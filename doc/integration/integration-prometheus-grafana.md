# StarRocks Cluster Integration With Prometheus and Grafana Service

This document describes how to integrate StarRocks cluster with Prometheus and Grafana service in a kubernetes
environment. From this document, you will learn,

* How to turn on prometheus metrics scrape for the StarRocks cluster
* How to import StarRocks Grafana dashboard

## 1. Prerequisites

+ Kubernetes: v1.18+ - v1.26
+ Prometheus helm chart: latest available version
+ Grafana helm chart: latest available version
+ StarRocks operator and helm chart: v1.7.1+

> see [deploy-prometheus-grafana.md](./deploy-prometheus-grafana.md) for how to deploy Prometheus and Grafana.

## 2. Deploy StarRocks Cluster

There are two ways to turn on the prometheus metrics scrape for the StarRocks cluster.

1. Turn on the prometheus metrics scrape by adding annotations
2. Turn on the prometheus metrics scrape by using ServiceMonitor CRD

### 2.1 Turn on the prometheus metrics scrape by adding annotations

Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks)
with some customized values.

Following is an example of the content of the `sr-values.yaml`.

* For chart v1.7.1 and below,

```yaml
# sr-values.yaml
starrocksFESpec:
  service:
    annotations:
      prometheus.io/path: "/metrics"
      prometheus.io/port: "8030"
      prometheus.io/scrape: "true"
  resources:
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 4
      memory: 4Gi
starrocksBESpec:
  service:
    annotations:
      prometheus.io/path: "/metrics"
      prometheus.io/port: "8040"
      prometheus.io/scrape: "true"
  resources:
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 4
      memory: 4Gi
```

* For chart v1.8.0 and above,

```yaml
# sr-values.yaml
starrocks:
  starrocksFESpec:
    service:
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/port: "8030"
        prometheus.io/scrape: "true"
    resources:
      requests:
        cpu: 1
        memory: 2Gi
      limits:
        cpu: 4
        memory: 4Gi
  starrocksBESpec:
    service:
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/port: "8040"
        prometheus.io/scrape: "true"
    resources:
      requests:
        cpu: 1
        memory: 2Gi
      limits:
        cpu: 4
        memory: 4Gi
```

Note that `"prometheus.io/*` annotations are the must items to be added, this will allow prometheus to auto discover
StarRocks PODs and to collect the metrics.
This method will restart the StarRocks cluster.

An equivalent StarRocks CRD may look like,

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: kube-starrocks
  namespace: default
spec:
  starRocksBeSpec:
    configMapInfo:
      configMapName: kube-starrocks-be-cm
      resolveKey: be.conf
    image: starrocks/be-ubuntu:3.5-latest
    limits:
      cpu: 4
      memory: 4Gi
    replicas: 1
    requests:
      cpu: 1
      memory: 2Gi
    service:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "8040"
        prometheus.io/scrape: "true"
  starRocksFeSpec:
    configMapInfo:
      configMapName: kube-starrocks-fe-cm
      resolveKey: fe.conf
    image: starrocks/fe-ubuntu:3.5-latest
    limits:
      cpu: 4
      memory: 4Gi
    replicas: 1
    requests:
      cpu: 1
      memory: 2Gi
    service:
      annotations:
        prometheus.io/path: /metrics
        prometheus.io/port: "8030"
        prometheus.io/scrape: "true"
```

Run the following commands to deploy StarRocks operator and StarRocks cluster,

```shell
helm repo add starrocks https://starrocks.github.io/starrocks-kubernetes-operator
helm repo update starrocks
helm install starrocks -f sr-values.yaml starrocks/kube-starrocks
```

### 2.2 Turn on the prometheus metrics scrape by using ServiceMonitor CRD

Compared to the annotation approach, ServiceMonitor allows for more flexible definition of selector and relabeling rules
in the future.
> Make
> sure [Deploy Prometheus and Grafana Service by Operator](./deploy-prometheus-grafana.md#2-deploy-prometheus-and-grafana-service-by-operator)

Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks)
with some customized values.

```shell
starrocks:
  metrics:
    serviceMonitor:
      enabled: true
```

Note: This only works for chart v1.8.4 and above.

## 3. Import StarRocks Grafana Dashboard

StarRocks grafana dashboard configuration for kubernetes environment is available
at https://github.com/StarRocks/starrocks/blob/main/extra/grafana/kubernetes/StarRocks-Overview-kubernetes-3.0.json

Detailed instruction can be
found [here](https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/#import-a-dashboard).

* Click **Dashboards** in the left-side menu.
* Click New and select **Import** in the dropdown menu.
* Paste dashboard JSON text directly into the text area

The import process enables you to change the name of the dashboard, pick the data source you want the dashboard to use,
and specify any metric prefixes (if the dashboard uses any).
