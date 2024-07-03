# Deploy Operator by operator Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/StarRocks/starrocks-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This topic describes how to use Helm to automatically deploy a StarRocks operator on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://kubernetes.io/).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).


## Install operator Chart

1. Add the Helm Chart Repo for StarRocks. The Helm Chart contains the definitions of the StarRocks Operator and the custom resource StarRocksCluster.
   1. Add the Helm Chart Repo.

      ```Bash
      helm repo add starrocks https://starrocks.github.io/starrocks-kubernetes-operator
      ```

   2. Update the Helm Chart Repo to the latest version.

      ```Bash
      helm repo update starrocks
      ```

   3. View the Helm Chart Repo that you added.

      ```Bash
      $ helm search repo starrocks
      NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
      starrocks/kube-starrocks      1.8.0            3.1-latest   kube-starrocks includes two subcharts, starrock...
      starrocks/operator            1.8.0            1.8.0        A Helm chart for StarRocks operator
      starrocks/starrocks           1.8.0            3.1-latest   A Helm chart for StarRocks cluster
      ```

2. Install the operator Chart.

   ```Bash
   helm install starrocks-operator starrocks/operator
   ```

   Please see [values.yaml](./values.yaml) for more details.

## Uninstall operator Chart

```Bash
helm uninstall starrocks-operator
```
