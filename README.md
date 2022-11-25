# StarRocks-Kubernetes-Operator

## 1 Overview
**(under development)**  
This operator is developed with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), which can deploy StarRocks CRD resources in kubernetes.

This Kubernetes Operator is able to deploy StarRocks' Front End (FE), Back End (BE) and Compute Node (CN) components together or separately. These components run in FQDN (fully qualified domain name) mode by default.

## 2 Requirements
 * kubernetes 1.18+
 * golang 1.18+

## 3 Supported Features
* FE decouples with CN and BE. FE is a must-have component, BE and CN can be optionally deployed.
* Support v2beta2 horizontalpodautoscalers for CN cluster.

## 3 Build the operator images
Get the official operator image from:[operator](https://hub.docker.com/repository/docker/fushilei/starrocks.operator/tags?page=1&ordering=last_updated).

Follow below instructions if you want to build your own image.

```bash
# under root directory, compile operator
make build 
# build docker image
make docker IMG="xxx"
# push image to docker hub
make push IMG="xxx"
```


## Deploy the StarRocks Operator in kubernetes
The first step is to deploy the StarRocks operator in your Kubernetes environment. The [deploy](./deploy) directory contains all the necessary yamls to deploy the operator. 

* Yaml files with `leader_` prefix are for operator election if willing to take multiples pods for backup. 

* The [manager.yaml](./deploy/manager.yaml) template is a deployment yaml to deploy the StarRocks operator. Remember to update corresponding `image` before applying to kubernetes. 

* Other yamls are facilities objects created for running the operator, include namespace, service account, rbac.

By default, the operator deploys the StarRocks cluster in `starrocks` namespace. Either specifying the namespace `-n <namespace>` when running `kubectl apply` or set the namespace meta field in every yaml files.

This example deploys StarRocks operator in the default `starrocks` namespace.
```bash
cd deploy
# create crd
kubectl apply -f starrocks.com_starrocksclusters.yaml
# create namespace
kubectl apply -f namespace.yaml;
# create rbac-roles the namespace starrocks  
kubectl apply -n starrocks -f leader_election_role.yaml
kubectl apply -n starrocks -f role.yaml
# create rbac-role-binding
kubectl apply -n starrocks -f role_binding.yaml
kubectl apply -n starrocks -f leader_election_role_binding.yaml
# create rbac-service-account
kubectl apply -n starrocks -f service_account.yaml
# create operator deployment
# replace image field with image which build in[3]
kubectl apply -n starrocks -f manager.yaml
```

## Deploy StarRocks with the StarRocks Operator
You need to prepare a separate yaml file to deploy the StarRocks FE, BE and CN components.

The [examples](./examples/starrocks) directory contains some simple example for reference.

You can use any of the yaml file as a starting point.

### Configure the StarRocks' components images
Official FE/CN/BE components images can be found from [dockerhub](https://hub.docker.com/u/starrocks):

You can specify the image name in the yaml file.
For example, the below configuration uses the `starrocks/alpine-fe:2.4.1` image for FE.
```yaml
starRocksFeSpec:
  image: starrocks/alpine-fe:2.4.1
```


### (Optional) Using ConfigMap to configure your StarRocks cluster

Those images contains default application configuration file, they can be overritten by configuring kubernetes configmap deployment crd. 

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

```commandline
kubectl apply -f starrocks-fe-and-be.yaml
```