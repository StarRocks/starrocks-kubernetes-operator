# StarRocks-Kubernetes-Operator
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview
**(under development)**  
This operator is developed with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), which can deploy StarRocks CRD resources in kubernetes.

This Kubernetes Operator is able to deploy StarRocks' Front End (FE), Back End (BE) and Compute Node (CN) components into your kubernetes environment. These components run in FQDN (fully qualified domain name) mode by default.

## Requirements
 * kubernetes 1.18+
 * golang 1.18+

## Supported Features
* FE decouples with CN and BE. FE is a must-have component, BE and CN can be optionally deployed.
* Support v2 horizontalpodautoscalers for CN cluster.

## （Optional) Build the operator images by yourself
Get the official operator image from [here](https://hub.docker.com/r/starrocks/centos-operator/tags).

### Build starrocks operator docker image
Follow below instructions if you want to build your own image.

```
DOCKER_BUILDKIT=1 docker build -t starrocks-kubernetes-operator/operator:<tag> .
```
E.g.
```bash
DOCKER_BUILDKIT=1 docker build -t starrocks-kubernetes-operator/operator:latest .
```

### Publish starrocks operator docker image
```
docker push ghcr.io/OWNER/starrocks-kubernetes-operator/operator:latest
```
E.g. 
Publish image to ghcr 
```shell
docker push ghcr.io/dengliu/starrocks-kubernetes-operator/operator:latest
```

## Install Operator in kubernetes
Apply the custom resource definition (CRD) for the Operator:
```shell
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```
Apply the Operator manifest. By default, the Operator is configured to install in the starrocks namespace. 
To use the Operator in a custom namespace, download the [Operator manifest](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml) and edit all instances of namespace: starrocks to specify your custom namespace.
Then apply this version of the manifest to the cluster with kubectl apply -f {local-file-path} instead of using the command below.
```shell
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## Deploy StarRocks
You need to prepare a separate yaml file to deploy the StarRocks FE, BE and CN components.
Thstarrockscluster CRD fields explains in [api.md](./doc/api.md).
The [examples](./examples/starrocks) directory contains some simple example for reference.

You can use any of the template yaml file as a starting point. You can further add more configurations into the template yaml file following this deployment documentation.

### Configure the StarRocks' components images
Official FE/CN/BE components images can be found from [dockerhub](https://hub.docker.com/u/starrocks):

You can specify the image name in the yaml file.
For example, the below configuration uses the `starrocks/alpine-fe:2.4.1` image for FE.
```yaml
starRocksFeSpec:
  image: starrocks/alpine-fe:2.4.1
```

### (Optional) Using ConfigMap to configure your StarRocks cluster

The official images contains default application configuration file, however, they can be overritten by configuring kubernetes configmap deployment crd. 

You can generate the configmap from an StarRocks configuration file.
Below is an example of creating a Kubernetes configmap `fe-config-map` from the `fe.conf` configuration file. You can do the same with BE and CN.
```shell
# create fe-config-map from starrocks/fe/conf/fe.conf file
kubectl create configmap fe-config-map --from-file=starrocks/fe/conf/fe.conf
```
Once the configmap is created, you can reference the configmap in the yaml file.
For example:
```yaml
# fe use configmap example
starRocksFeSpec:
  configMapInfo:
    configMapName: fe-config-map
    resolveKey: fe.conf
# cn use configmap example
starRocksCnSpec:
  configMapInfo:
    configMapName: cn-config-map
    resolveKey: cn.conf
# be use configmap example
  starRocksBeSpec:
    configMapInfo:
    configMapName: be-config-map
    resolveKey: be.conf
```
### (Optional) Configuring storage volume
External storage can be used to store FE meta and BE data for persistence. `storageVolumes` can be specified in corresponding component spec to enable external storage volumes auto provisioning. Note that the specific `storageClassName` should be available in kubernetes cluster before enabling this storageVolume feature.

If `StorageVolume` info is not specified in CRD spec, the operator will use emptydir mode to store FE meta and BE data. 

**FE storage example**
```yaml
starRocksFeSpec:
  storageVolumes:
    - name: fe-meta
      storageClassName: meta-storage
      storageSize: 10Gi
      mountPath: /opt/starrocks/fe/meta # overwrite the default meta path
```
**BE storage example**
```yaml
starRocksBeSpec:
  storageVolumes:
    - name: be-data
      storageClassName: data-storage
      storageSize: 1Ti
      mountPath: /opt/starrocks/be/storage # overwrite the default data path
```

### Deploy the StarRocks cluster
For demonstration purpose, we use the [starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml) example template to start a 3 FE and 3 BE StarRocks cluster.

```bash
kubectl apply -f starrocks-fe-and-be.yaml
```

### Connect to the deployed StarRocks Cluster

After deploying the StarRocks cluster, you can use `kubectl get svc -n <namespace>` to find the IP to connect to. For example if the namespace that starrocks is deployed into is `starrocks`, you can:
```bash
kubectl get svc -n starrocks
```
`<your-StarRocksCluster-name>-fe-service`'s clusterIP is the IP to use to connect to StarRocks FE.

## Stop the StarRocks cluster

Delete the custom resource:
```shell
kubectl delete -f starrocks-fe-and-be.yaml
```

Remove the Operator:
```shell
kubectl delete -f  https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## Others 
### helm
StarRocks have supported helm use.  
[helm chart](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks). [github repo](https://github.com/StarRocks/helm-charts)
