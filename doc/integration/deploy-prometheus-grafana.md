# Deploy Prometheus and Grafana

There are two ways to deploy Prometheus and Grafana service in the Kubernetes environment.

1. Deploy Prometheus and Grafana Service Directly.
2. Deploy Prometheus and Grafana Service by Operator.

Inorder to scrape the metrics by ServiceMonitor CRD, you must install the Prometheus Operator in the Kubernetes.
That is to say, you must choose the second way to deploy Prometheus and Grafana Service.

## 1. Deploy Prometheus and Grafana Service Directly

### 1.1 Deploy Prometheus Service

> SKIP this step if the Prometheus service is already available in the Kubernetes environment.

Follow the instructions
from [Prometheus Helm Chart](https://artifacthub.io/packages/helm/prometheus-community/prometheus) to deploy the
Prometheus service into the Kubernetes environment.

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
helm install prometheus prometheus-community/prometheus
```

### 1.2 Deploy Grafana Service

> SKIP this step if the Grafana service is already available in the Kubernetes environment.

Follow the instructions from [Grafana Helm Chart](https://artifacthub.io/packages/helm/grafana/grafana) to deploy the
Grafana service into the Kubernetes environment.

Need override the default value of the grafana service, change the Grafana service type from default `ClusterIP`
to `LoadBalancer` in order to access the Grafana service from the outside Kubernetes network.

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

## 2. Deploy Prometheus and Grafana Service by Operator

you can also install `kube-prometheus-stack` chart which will install Prometheus operator, Prometheus service
and Grafana service.

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update prometheus-community
helm install prometheus prometheus-community/kube-prometheus-stack
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

From the navigation path, `"Administration" -> "Configuration" -> "Add Data Source"`, set the HTTP.URL
to `http://prometheus-server.default`. Change `default` to the correct namespace if the prometheus service is not
deployed in the default namespace. Refer to Grafana
doc [here](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/) for
detailed instructions.

