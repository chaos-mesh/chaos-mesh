# Monitoring Chaos Mesh

Most of Chaos Mesh components has some built-in metrics collectors/exporters for monitoring with [Prometheus](https://prometheus.io/).

## Configure with Prometheus Operator

With Prometheus Operator, we could easily monitoring Kubernetes Concepts like Service, Pod, etc with CRD.

We provide several `ServiceMonitor` examples under directory `prometheus-operator`, you could build your own monitoring rules with them as references.

## Configure with Prometheus Kubernetes Service Discovery

There are several common-used annotations like `prometehus.io/scrape` and `prometehus.io/path` on Chaos Mesh components to configure Prometheus Kubernetes Discovery. After installing Chaos Mesh, you could use these following annotations to configure Prometheus Kubernetes Service Discovery.

We provide a configuration example under directory `prometheus-kubernetes-service-discovery`.
## Setup Grafana Dashboard

There are several Grafana Dashboards available for monitoring Chaos Mesh components:

- Chaos Mesh Overview, [Grafana Dashboard](https://grafana.com/grafana/dashboards/15918)
- Chaos Mesh | Chaos Daemon, [Grafana Dashboard](https://grafana.com/grafana/dashboards/15919)

And there are also other dashboards available for monitoring [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime), which drives Chaos Mesh:

- Controller Runtime Controllers Detail, [Grafana Dashboard](https://grafana.com/grafana/dashboards/15920)
- Controller Runtime Webhooks Detail, [Grafana Dashboard](https://grafana.com/grafana/dashboards/15921)
