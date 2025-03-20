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







