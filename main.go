package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	namespace         = "rabbitmq"    // Prometheus中使用的命名空间
	defaultConfigPath = "config.json" // 配置文件的默认路径
)

var log = logrus.New() // 初始化日志记录器

// 定义Prometheus度量指标（metrics）
var (
	connectionsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connections_total",
			Help:      "Total number of open connections.", // 总连接数
		},
		[]string{
			"node", // 节点标签
		},
	)
	channelsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "channels_total",
			Help:      "Total number of open channels.", // 总通道数
		},
		[]string{
			"node", // 节点标签
		},
	)
	queuesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "queues_total",
			Help:      "Total number of queues in use.", // 总队列数
		},
		[]string{
			"node", // 节点标签
		},
	)
	consumersTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumers_total",
			Help:      "Total number of message consumers.", // 消费者总数
		},
		[]string{
			"node", // 节点标签
		},
	)
	exchangesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exchanges_total",
			Help:      "Total number of exchanges in use.", // 总交换机数
		},
		[]string{
			"node", // 节点标签
		},
	)
	messagesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "messages_total",
			Help:      "Total number of messages in all queues.", // 所有队列的消息总数
		},
		[]string{
			"node", // 节点标签
		},
	)
)

// 配置文件结构体
type Config struct {
	Nodes    *[]Node `json:"nodes"`        // 节点数组
	Port     string  `json:"port"`         // 监听端口
	Interval string  `json:"req_interval"` // 请求间隔
}

// 节点结构体，表示RabbitMQ的一个节点
type Node struct {
	Name     string `json:"name"`                   // 节点名称
	Url      string `json:"url"`                    // 节点URL
	Uname    string `json:"uname"`                  // 用户名
	Password string `json:"password"`               // 密码
	Interval string `json:"req_interval,omitempty"` // 请求间隔，可选字段
}

// 发送API请求，返回JSON解码器
func sendApiRequest(hostname, username, password, query string) *json.Decoder {
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostname+query, nil)
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)

	if err != nil {
		log.Error(err)
		panic(err)
	}
	return json.NewDecoder(resp.Body)
}

// 获取RabbitMQ概览信息，更新Prometheus度量指标
func getOverview(hostname, username, password string) {
	decoder := sendApiRequest(hostname, username, password, "/api/overview")
	response := decodeObj(decoder)

	metrics := make(map[string]float64)
	for k, v := range response["object_totals"].(map[string]interface{}) {
		metrics[k] = v.(float64) // 将各项度量数据保存到metrics中
	}
	nodename, _ := response["node"].(string)

	// 更新Prometheus度量指标
	channelsTotal.WithLabelValues(nodename).Set(metrics["channels"])
	connectionsTotal.WithLabelValues(nodename).Set(metrics["connections"])
	consumersTotal.WithLabelValues(nodename).Set(metrics["consumers"])
	queuesTotal.WithLabelValues(nodename).Set(metrics["queues"])
	exchangesTotal.WithLabelValues(nodename).Set(metrics["exchanges"])
}

// 获取所有队列的消息数，更新Prometheus度量指标
func getNumberOfMessages(hostname, username, password string) {
	decoder := sendApiRequest(hostname, username, password, "/api/queues")
	response := decodeObjArray(decoder)
	nodename := response[0]["node"].(string)

	total_messages := 0.0
	for _, v := range response {
		total_messages += v["messages"].(float64) // 累加所有队列的消息数
	}
	messagesTotal.WithLabelValues(nodename).Set(total_messages)
}

// 解码JSON对象
func decodeObj(d *json.Decoder) map[string]interface{} {
	var response map[string]interface{}

	if err := d.Decode(&response); err != nil {
		log.Error(err)
	}
	return response
}

// 解码JSON数组
func decodeObjArray(d *json.Decoder) []map[string]interface{} {
	var response []map[string]interface{}

	if err := d.Decode(&response); err != nil {
		log.Error(err)
	}
	return response
}

// 更新节点的统计信息
func updateNodesStats(config *Config) {
	for _, node := range *config.Nodes {

		// 如果节点没有指定请求间隔，使用全局配置的间隔
		if len(node.Interval) == 0 {
			node.Interval = config.Interval
		}
		go runRequestLoop(node) // 启动异步请求
	}
}

// 请求数据，更新Prometheus指标
func requestData(node Node) {
	defer func() {
		if r := recover(); r != nil {
			dt := 10 * time.Second
			time.Sleep(dt) // 如果发生异常，休眠一段时间再重试
		}
	}()

	getOverview(node.Url, node.Uname, node.Password)
	getNumberOfMessages(node.Url, node.Uname, node.Password)

	log.Info("Metrics updated successfully.") // 日志记录更新成功

	// 解析请求间隔
	dt, err := time.ParseDuration(node.Interval)
	if err != nil {
		log.Warn(err)
		dt = 30 * time.Second // 默认间隔为30秒
	}
	time.Sleep(dt) // 按照间隔时间进行休眠
}

// 持续请求节点数据
func runRequestLoop(node Node) {
	for {
		requestData(node)
	}
}

// 加载配置文件
func loadConfig(path string, c *Config) bool {
	defer func() {
		if r := recover(); r != nil {
			dt := 10 * time.Second
			time.Sleep(dt) // 如果加载配置文件时发生异常，休眠一段时间再重试
		}
	}()

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	err = json.Unmarshal(file, c)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	return true
}

// 持续加载配置文件
func runLoadConfigLoop(path string, c *Config) {
	for {
		is_ok := loadConfig(path, c)
		if is_ok == true {
			break // 配置文件加载成功后退出循环
		}
	}
}

// 主函数
func main() {
	configPath := defaultConfigPath // 默认配置文件路径
	if len(os.Args) > 1 {
		configPath = os.Args[1] // 如果提供了命令行参数，使用第一个参数作为配置文件路径
	}
	log.Out = os.Stdout // 设置日志输出到标准输出

	var config Config

	runLoadConfigLoop(configPath, &config) // 加载配置文件
	updateNodesStats(&config)              // 更新节点状态

	// 配置HTTP服务器，提供Prometheus度量指标
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>RabbitMQ Exporter</title></head>
             <body>
             <h1>RabbitMQ Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Infof("Starting RabbitMQ exporter on port: %s.", config.Port)
	http.ListenAndServe(":"+config.Port, nil) // 启动HTTP服务器
}

// 初始化时注册Prometheus度量指标
func init() {
	prometheus.MustRegister(channelsTotal)
	prometheus.MustRegister(connectionsTotal)
	prometheus.MustRegister(queuesTotal)
	prometheus.MustRegister(exchangesTotal)
	prometheus.MustRegister(consumersTotal)
	prometheus.MustRegister(messagesTotal)
}
