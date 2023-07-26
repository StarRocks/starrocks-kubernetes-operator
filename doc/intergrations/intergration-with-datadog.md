Datadog provides easy integration of log and metrics collection in the kubernetes environment.
This document describes how to integrate the Datadog with Starrocks.

> see https://docs.datadoghq.com/containers/kubernetes/installation/?tab=helm for more details.

## 1. Install the Datadog chart

If this is a fresh install, add the Helm Datadog repo:

```bash
helm repo add datadog https://helm.datadoghq.com
helm repo update
```

Create a Kubernetes Secret to store your Datadog API key and app key:

```bash
export DD_API_KEY=xxx
export DD_APP_KEY=yyy
kubectl create secret generic datadog-secret --from-literal api-key=$DD_API_KEY --from-literal app-key=$DD_APP_KEY
```

Set the following parameters in your `datadog-values.yaml` to reference the secret:

```bash
datadog:
  apiKeyExistingSecret: datadog-secret
  appKeyExistingSecret: datadog-secret
  site: <DATADOG_SITE>   # Replace <DATADOG_SITE> with your Datadog site.
  logs:
    enabled: true
  prometheusScrape:
    enabled: true
    serviceEndpoints: true
```

Install the chart with the release name `datadog-agent`:

```bash
helm install datadog-agent  -f datadog-values.yaml datadog/datadog
```

## 2. Install or upgrade the starrocks cluster

If this is a fresh install, add the Helm Starrocks Operator repo
```bash
helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
helm repo update
```

Prepare the `sr-values.yaml` file for starrocks cluster:

```bash
# sr-values.yaml
starrocksFESpec:
  annotations:
    # change the value of ad.datadoghq.com/fe.logs to your own value
    ad.datadoghq.com/fe.logs: '[{"source": "fe","service": "starrocks","tags": ["env:test"]}]'
  service:
    annotations:
      prometheus.io/path: "/metrics"
      prometheus.io/port: "8030"
      prometheus.io/scrape: "true"
  feEnvVars:
    - name: LOG_CONSOLE
      value: "1"
starrocksBeSpec:
  annotations:
    # change the value of ad.datadoghq.com/be.logs to your own value
    ad.datadoghq.com/be.logs: '[{"source": "be","service": "starrocks","tags": ["env:test"]}]'
  service:
    annotations:
      prometheus.io/path: "/metrics"
      prometheus.io/port: "8040"
      prometheus.io/scrape: "true"
  beEnvVars:
    - name: LOG_CONSOLE
      value: "1"
```

> the `prometheus.io/port` in `starrocksFESpec` need to be the same value of `http_port` in `starrocksFESpec.confg`.
> the `prometheus.io/port` in `starrocksBESpec` need to be the same value of `webserver_port` in `starrocksBESpec.confg`.

Install the starrocks cluster:
```bash
helm install -n starrocks starrocks -f sr-values.yaml starrocks-community/kube-starrocks
```