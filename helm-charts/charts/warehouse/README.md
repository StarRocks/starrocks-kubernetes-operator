# Deploy StarRocks Warehouse by starrocks Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/StarRocks/starrocks-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a
Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This
topic describes how to use Helm to automatically deploy a StarRocks cluster on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://kubernetes.io/).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).
- [Install StarRocks operator](../kube-starrocks/charts/operator/README.md).
- [Install StarRocks cluster](../kube-starrocks/charts/starrocks/README.md).

> Note: Warehouse is an enterprise feature for StarRocks.

## Install Warehouse Chart

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

2. Prepare the values.yaml file.

   ```yaml
   # The name of warehouse in StarRocks. You can execute `show warehouses` command in SQL to see the created warehouse.
   nameOverride: "wh1"
   spec:
     # Make sure the StarRocks cluster exists in the same namespace.
     # You can check it by running `kubectl -n starrocks get starrocksclusters.starrocks.com`.
     starRocksClusterName: kube-starrocks
     replicas: 1
     image: your-enterprise-image-version-for-cn
     resources:
       limits:
         cpu: 8
         memory: 8Gi
       requests:
         cpu: 8
         memory: 8Gi
   ```

3. Install the warehouse Chart.

    ```bash
    # Use the above values.yaml to deploy a warehouse in namespace starrocks
    helm -n starrocks install warehouse starrocks-community/warehouse -f values.yaml

    # Restart the StarRocks operator to make it aware of the new CRD
    kubectl -n starrocks rollout restart deployment kube-starrocks-operator
    ```

   Please see [values.yaml](./values.yaml) for more details.

## Uninstall Warehouse

```bash
helm uninstall warehouse
```
