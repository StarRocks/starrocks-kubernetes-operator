# Deploy Operator and CelerData Cluster by kube-celerdata Chart

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Release Charts](https://img.shields.io/badge/Release-helmcharts-green.svg)](https://github.com/celerdata/celerdata-kubernetes-operator/releases)

[Helm](https://helm.sh/) is a package manager for Kubernetes. A [Helm Chart](https://helm.sh/docs/topics/charts/) is a Helm package and contains all of the resource definitions necessary to run an application on a Kubernetes cluster. This topic describes how to use Helm to automatically deploy a CelerData operator and cluster on a Kubernetes cluster.

## Before you begin

- [Create a Kubernetes cluster](https://kubernetes.io/).
- [Install Helm](https://helm.sh/docs/intro/quickstart/).

## Install kube-celerdata Chart

1. Add the Helm Chart Repo for CelerData. The Helm Chart contains the definitions of the CelerData Operator and the custom resource CelerDataCluster.
   1. Add the Helm Chart Repo.

      ```Bash
      helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
      ```

   2. Update the Helm Chart Repo to the latest version.

      ```Bash
      helm repo update celerdata
      ```

   3. View the Helm Chart Repo that you added.

      ```Bash
      $ helm search repo celerdata
      NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
      celerdata/kube-celerdata      1.8.0            3.1-latest   kube-celerdata includes two subcharts, operator and celerdata
      us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator            1.8.0            1.8.0        A Helm chart for CelerData operator
      celerdata/celerdata           1.8.0            3.1-latest   A Helm chart for CelerData cluster
      ```

2. Use the default **[values.yaml](https://github.com/celerdata/celerdata-kubernetes-operator/blob/main/helm-charts/charts/kube-celerdata/values.yaml)** of the Helm Chart to deploy the CelerData Operator and CelerData cluster, or create a YAML file to customize your deployment configurations.
   1. Deployment with default configurations

      Run the following command to deploy the CelerData Operator and the CelerData cluster which consists of one FE and one BE:

      ```Bash
      $ helm install celerdata celerdata/kube-celerdata
      # If the following result is returned, the CelerData Operator and CelerData cluster are being deployed.
      NAME: celerdata
      LAST DEPLOYED: Tue Aug 15 15:12:00 2023
      NAMESPACE: celerdata
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```

   2. Deployment with custom configurations
      - Create a YAML file, for example, **my-values.yaml**, and customize the configurations for the CelerData Operator and CelerData cluster in the YAML file. For the supported parameters and descriptions, see the comments in the default **[values.yaml](https://github.com/celerdata/celerdata-kubernetes-operator/blob/main/helm-charts/charts/kube-celerdata/values.yaml)** of the Helm Chart.
      - Run the following command to deploy the CelerData Operator and CelerData cluster with the custom configurations in **my-values.yaml**.

        ```Bash
        helm install -f my-values.yaml celerdata celerdata/kube-celerdata
        ```

    Deployment takes a while. During this period, you can check the deployment status by using the prompt command in the returned result of the deployment command above. The default prompt command is as follows:

    ```Bash
    $ kubectl --namespace default get celerdatacluster -l "cluster=kube-celerdata"
    # If the following result is returned, the deployment has been successfully completed.
    NAME             FESTATUS   CNSTATUS   BESTATUS
    kube-celerdata   running               running
    ```

    You can also run `kubectl get pods` to check the deployment status. If all Pods are in the `Running` state and all containers within the Pods are `READY`, the deployment has been successfully completed.

    ```Bash
    $ kubectl get pods
    NAME                                       READY   STATUS    RESTARTS   AGE
    kube-celerdata-be-0                        1/1     Running   0          2m50s
    kube-celerdata-fe-0                        1/1     Running   0          4m31s
    kube-celerdata-operator-69c5c64595-pc7fv   1/1     Running   0          4m50s
    ```

## Upgrade kube-celerdata Chart

If you need to upgrade the CelerData Operator and CelerData cluster, run the following command:
```bash
helm upgrade -f my-values.yaml celerdata celerdata/kube-celerdata
```

## Uninstall kube-celerdata Chart

If you need to uninstall the CelerData Operator and CelerData cluster, run the following command:
```bash
helm uninstall celerdata
```

Search Helm Chart maintained by CelerData on Artifact Hub. See [kube-celerdata](https://github.com/StarRocks/starrocks-kubernetes-operator/tree/main/helm-charts/charts/kube-starrocks).
