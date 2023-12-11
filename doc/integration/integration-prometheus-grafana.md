# StarRocks Cluster Integration With Prometheus and Grafana Service

This document describes how to integrate StarRocks cluster with Prometheus and Grafana service in a kubernetes environment. From this document, you will learn,
* How to turn on prometheus metrics scrape for the StarRocks cluster
* How to import StarRocks Grafana dashboard


## 1. Prerequisites

+ Kubernetes: v1.18+ - v1.26
+ Prometheus helm chart: latest available version
+ Grafana helm chart: latest available version
+ StarRocks operator and helm chart: v1.7.1+

## 2. Deploy Prometheus Service

> SKIP this step if the Prometheus service is already available in the Kubernetes environment.

Follow the instructions from [Prometheus Helm Chart](https://artifacthub.io/packages/helm/prometheus-community/prometheus) to deploy the Prometheus service into the Kubernetes environment.

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
helm install prometheus prometheus-community/prometheus
```

## 3. Deploy Grafana Service

> SKIP this step if the Grafana service is already available in the Kubernetes environment.

Follow the instructions from [Grafana Helm Chart](https://artifacthub.io/packages/helm/grafana/grafana) to deploy the Grafana service into the Kubernetes environment.

Need override the default value of the grafana service, change the Grafana service type from default `ClusterIP` to `LoadBalancer` in order to access the Grafana service from the outside Kubernetes network.

An example `grafana-values.yaml` may look like as the following snippet.
```yaml
service:
  enabled: true
  type: "LoadBalancer"
```

Install the Grafana service with the following commands.
```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update grafana
helm install grafana -f grafana-values.yaml grafana/grafana
```

## 4. Login And Configure Grafana Service

> SKIP this section if the prometheus data source is already configured in Grafana.

### 4.1 Get Grafana Service External Address

Run the following command to get the external ip address
```shell
kubectl get svc grafana
```
A possible result may look like the follows
```text
$ kubectl get svc grafana
NAME      TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)        AGE
grafana   LoadBalancer   10.40.6.237   35.239.53.124   80:31362/TCP   4m25s
```
### 4.2 Get Grafana Admin Password

Run the following command to get the login password for admin user.
```shell
kubectl get secrets grafana -o yaml | grep admin-password | awk '{print $2}' | base64 -d
```
### 4.3 Login Grafana and Configure Prometheus Data Source

Access Grafana web GUI with `http://<external_ip>/`, login with `admin:<password>`.

From the navigation path, `"Administration" -> "Configuration" -> "Add Data Source"`, set the HTTP.URL to `http://prometheus-server.default`. Change `default` to the correct namespace if the prometheus service is not deployed in the default namespace. Refer to Grafana doc [here](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/) for detailed instructions.

## 5. Deploy StarRocks Cluster

There are two ways to turn on the prometheus metrics scrape for the StarRocks cluster.

1. Turn on the prometheus metrics scrape by adding annotations
2. Turn on the prometheus metrics scrape by using ServiceMonitor CRD

### 5.1 Turn on the prometheus metrics scrape by adding annotations

Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks) with some customized values.

Following is an example of the content of the `sr-values.yaml`.

* For chart v1.7.1 and below,
```yaml
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

Note that `"prometheus.io/*` annotations are the must items to be added, this will allow prometheus to auto discover StarRocks PODs and to collect the metrics.
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
    image: starrocks/be-ubuntu:3.1-latest
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
    image: starrocks/fe-ubuntu:3.1-latest
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
helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
helm repo update starrocks-community
helm install starrocks -f sr-values.yaml starrocks-community/kube-starrocks
```

### 5.2 Turn on the prometheus metrics scrape by using ServiceMonitor CRD

Compared to the annotation approach, ServiceMonitor allows for more flexible definition of selector and relabeling rules in the future.
Follow the instructions from [StarRocks Helm Chart](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks) with some customized values.

```shell
starrocks:
  metrics:
    serviceMonitor:
      enabled: true
```

Note: This only works for chart v1.8.4 and above.


## 6. Import StarRocks Grafana Dashboard

StarRocks grafana dashboard configuration for kubernetes environment is available at https://github.com/StarRocks/starrocks/blob/main/extra/grafana/kubernetes/StarRocks-Overview-kubernetes-3.0.json

Detailed instruction can be found [here](https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/#import-a-dashboard).
* Click **Dashboards** in the left-side menu.
* Click New and select **Import** in the dropdown menu.
* Paste dashboard JSON text directly into the text area

The import process enables you to change the name of the dashboard, pick the data source you want the dashboard to use, and specify any metric prefixes (if the dashboard uses any).
