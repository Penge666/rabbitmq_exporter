# RabbitMQ Exporter

Prometheus exporter for RabbitMQ metrics, based on RabbitMQ HTTP API.

### Dependencies

* Prometheus [client](https://github.com/prometheus/client_golang) for Golang
* [Logging](https://github.com/Sirupsen/logrus)

### Setting up locally

1. You need **RabbitMQ**. For local setup I recommend this [docker box](https://github.com/mikaelhg/docker-rabbitmq). It's "one-click" solution.

2. For OS-specific **Docker** installation checkout these [instructions](https://docs.docker.com/installation/).

3. Building rabbitmq_exporter:【使用 **Docker** 构建一个名为 `rabbitmq_exporter` 的 Docker 镜像。】

    ```
    $ docker build -t rabbitmq_exporter .
    ```

4. Running:

        $ docker run --publish 6060:9672 --rm rabbitmq_exporter

上述命令会启动一个名为 `rabbitmq_exporter` 的 Docker 容器，并执行以下操作：

- 将容器内的 `9672` 端口映射到宿主机的 `6060` 端口。
- 在容器停止后，自动删除该容器。
- 启动该镜像的默认服务，通常是暴露 RabbitMQ 相关的监控数据。

初始化项目

```bash
go mod init rabbitmq_exporter
go mod tidy
```

Now your metrics are available through [http://localhost:9672/metrics](http://localhost::9672/metrics).

### Metrics

Total number of:

* channels
* connections
* consumers
* exchanges
* queues
* messages

## 简介

这个项目是一个 **RabbitMQ Exporter**，用于从 RabbitMQ 实例中收集各种指标，并将其暴露为 Prometheus 可抓取的指标格式。这样，Prometheus 就可以监控 RabbitMQ 的性能和健康状态，并将数据用于后续的可视化和报警。

## 步骤

该代码实现了一个Prometheus监控RabbitMQ的Exporter，以下是其主要步骤：

1. **配置加载**：从指定的配置文件（默认为`config.json`）读取RabbitMQ节点信息，包含节点URL、用户名、密码及请求间隔。
2. **API请求**：通过HTTP请求获取RabbitMQ的各种统计信息，如连接数、通道数、队列数、消费者数和交换器数等。
3. **Prometheus指标**：使用Prometheus的`GaugeVec`定义多个监控指标，按节点区分，实时更新RabbitMQ的各项指标数据。
4. **定期数据更新**：根据配置文件中的请求间隔，周期性地获取RabbitMQ的最新状态并更新Prometheus指标。
5. **Prometheus服务**：启动一个HTTP服务，提供`/metrics`端点，供Prometheus抓取数据。
6. **错误处理与重试**：对异常情况进行捕获与处理，确保Exporter稳定运行。

通过此Exporter，Prometheus能够监控RabbitMQ的性能和健康状况。