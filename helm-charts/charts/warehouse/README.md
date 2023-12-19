# Deploy StarRocks Warehouse by starrocks Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/StarRocks/starrocks-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a
Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This
topic describes how to use Helm to automatically deploy a StarRocks cluster on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://docs.starrocks.io/en-us/latest/deployment/sr_operator#create-kubernetes-cluster).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).
- [Install StarRocks operator](../kube-starrocks/charts/operator/README.md).
- [Install StarRocks cluster](../kube-starrocks/charts/starrocks/README.md).

## Install starrocks Chart

1. Add the StarRocks Helm repository.

    ```bash
    $ helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
    $ helm repo update
    $ helm search repo starrocks-community
    NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
    starrocks-community/kube-starrocks      1.9.0            3.1-latest   kube-starrocks includes two subcharts, starrock...
    starrocks-community/operator            1.9.0            1.9.0        A Helm chart for StarRocks operator
    starrocks-community/starrocks           1.9.0            3.1-latest   A Helm chart for StarRocks cluster
    starrocks-community/warehouse           1.9.0            3.1-latest   A Helm chart for StarRocks cluster
    ```

2. Install the starrocks Chart.

    ```bash
    helm install warehouse starrocks-community/warehouse
    ```

   Please see [values.yaml](./values.yaml) for more details.

## Uninstall starrocks Chart

```bash
helm uninstall warehouse
```
