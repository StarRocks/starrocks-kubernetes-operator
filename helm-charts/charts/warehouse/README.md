# Deploy CelerData Warehouse by warehouse Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/celerdata/celerdata-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a
Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This
topic describes how to use Helm to automatically deploy a CelerData warehouse on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://kubernetes.io/).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).
- [Install CelerData operator](../kube-celerdata/charts/operator/README.md).
- [Install CelerData cluster](../kube-celerdata/charts/celerdata/README.md).

> Note: Warehouse is an enterprise feature for CelerData.

## Install Warehouse Chart

1. Add the CelerData Helm repository.

    ```bash
    $ helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
    $ helm repo update celerdata
    $ helm search repo celerdata
    NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
    celerdata/kube-celerdata      1.9.0            3.1-latest   kube-celerdata includes two subcharts, operator and celerdata
    us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator            1.9.0            1.9.0        A Helm chart for CelerData operator
    celerdata/celerdata           1.9.0            3.1-latest   A Helm chart for CelerData cluster
    celerdata/warehouse           1.9.0            3.1-latest   A Helm chart for CelerData cluster
    ```

2. Prepare the values.yaml file.

   ```yaml
   # The name of warehouse in CelerData. You can execute `show warehouses` command in SQL to see the created warehouse.
   nameOverride: "wh1"
   spec:
     # Make sure the CelerData cluster exists in the same namespace.
     # You can check it by running `kubectl -n celerdata get celerdataclusters.celerdata.com`.
     celerDataClusterName: kube-celerdata
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
    # Use the above values.yaml to deploy a warehouse in namespace celerdata
    helm -n celerdata install warehouse celerdata/warehouse -f values.yaml

    # Restart the CelerData operator to make it aware of the new CRD
    kubectl -n celerdata rollout restart deployment kube-celerdata-operator
    ```

   Please see [values.yaml](./values.yaml) for more details.

## Uninstall Warehouse

```bash
helm uninstall warehouse
```
