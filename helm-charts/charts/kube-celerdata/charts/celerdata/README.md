# Deploy CelerData Cluster by celerdata Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/celerdata/celerdata-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This topic describes how to use Helm to automatically deploy a CelerData cluster on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://kubernetes.io/).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).
- [Install CelerData operator](../operator/README.md#install-operator-chart).

## Install celerdata Chart

1. Add the CelerData Helm repository.

    ```bash
    $ helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
    $ helm repo update celerdata
    $ helm search repo celerdata
    NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
    celerdata/kube-celerdata      1.8.0            3.1-latest   kube-celerdata includes two subcharts, celerdata...
    us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator            1.8.0            1.8.0        A Helm chart for CelerData operator
    celerdata/celerdata           1.8.0            3.1-latest   A Helm chart for CelerData cluster
    ```

2. Install the celerdata Chart.

    ```bash
    helm install celerdata celerdata/celerdata
    ```

    Please see [values.yaml](./values.yaml) for more details.

## Uninstall celerdata Chart

```bash
helm uninstall celerdata
```
