# Observability With Grafana Stack

This repository contains the accompanying code to my series on [Observability With Grafana Stack](https://riyazali.net/series/grafana-stack).

This series shows how to setup [the three of observability](https://www.datadoghq.com/knowledge-center/observability) on [Grafana](http://grafana.com)

1. Logs with [Loki](https://grafana.com/oss/loki/) (using [Alloy](https://grafana.com/oss/alloy-opentelemetry-collector/) as an agent)
2. Metrics with [Prometheus](http://prometheus.io)
3. Traces with [Tempo](https://grafana.com/oss/tempo/)

The sample Golang application here also uses [OpenTelemetry's `go` sdk](https://github.com/open-telemetry/opentelemetry-go) 

<br />

![stack architecture](https://github.com/user-attachments/assets/f0744ba9-e139-40dd-b5f1-32fd3f545fa3)

## Running the stack

To deploy the stack along with the application, run:

```bash
> docker-compose up -d --build
```

To generate some artificial load using [`k6`](http://k6.io), run:

```bash
> k6 run k6-script.js
```
