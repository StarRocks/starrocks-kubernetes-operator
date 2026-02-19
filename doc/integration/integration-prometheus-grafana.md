# CelerData Cluster Integration With Prometheus and Grafana Service

This document describes how to integrate CelerData cluster with Prometheus and Grafana service in a kubernetes
environment. From this document, you will learn,

* How to turn on prometheus metrics scrape for the CelerData cluster
* How to import StarRocks Grafana dashboard

## 1. Prerequisites

+ Kubernetes: v1.23.0+
+ Prometheus helm chart: latest available version
+ Grafana helm chart: latest available version
+ CelerData operator and helm chart: v1.7.1+

> see [deploy-prometheus-grafana.md](./deploy-prometheus-grafana.md) for how to deploy Prometheus and Grafana.

## 2. Deploy CelerData Cluster

There are two ways to turn on the prometheus metrics scrape for the CelerData cluster.

1. Turn on the prometheus metrics scrape by adding annotations
2. Turn on the prometheus metrics scrape by using ServiceMonitor CRD

### 2.1 Turn on the prometheus metrics scrape by adding annotations

Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-celerdata/kube-celerdata)
with some customized values.

Following is an example of the content of the `sr-values.yaml`.

* For chart v1.7.1 and below,

```yaml
# sr-values.yaml
celerDataFeSpec:
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
celerDataBeSpec:
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
  celerDataFeSpec:
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
  celerDataBeSpec:
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
This method will restart the CelerData cluster.

An equivalent StarRocks CRD may look like,

```yaml
apiVersion: celerdata.com/v1
kind: CelerDataCluster
metadata:
  name: kube-celerdata
  namespace: default
spec:
  celerDataBeSpec:
    configMapInfo:
      configMapName: kube-celerdata-be-cm
      resolveKey: be.conf
    image: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/be-ubuntu:3.5-latest
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
  celerDataFeSpec:
    configMapInfo:
      configMapName: kube-celerdata-fe-cm
      resolveKey: fe.conf
    image: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/fe-ubuntu:3.5-latest
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

Run the following commands to deploy CelerData operator and CelerData cluster,

```shell
helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
helm repo update celerdata
helm install celerdata -f sr-values.yaml celerdata/kube-celerdata
```

### 2.2 Turn on the prometheus metrics scrape by using ServiceMonitor CRD

Compared to the annotation approach, ServiceMonitor allows for more flexible definition of selector and relabeling rules
in the future.
> Make
> sure [Deploy Prometheus and Grafana Service by Operator](./deploy-prometheus-grafana.md#2-deploy-prometheus-and-grafana-service-by-operator)

Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-celerdata/kube-celerdata)
with some customized values.

```shell
celerdata:
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
