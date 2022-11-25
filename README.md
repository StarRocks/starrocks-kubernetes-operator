# StarRocks-Kubernetes-Operator

## 1 Overview
**(under development)**  
This operator is developed with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), which can deploy StarRocks CRD resource in kubernetes.

The StarRocks cluster deployed by this operator incldues FE, BE and CN components. These components are running in FQDN mode by default.

If StorageVolume info is not specified in CRD spec, the operator will use emptydir mode to store FE meta and BE data. Please refer to the storageVolume section on how to enable customized storageVolume.

## 2 Requirements
 * kubernetes 1.18+
 * golang 1.18+

## 3 Supported Features
* FE decouples with CN and BE. FE is a must-have component, BE and CN can be optionally deployed.
* Support v2beta2 horizontalpodautoscalers for CN cluster.

## 3 Build operator images
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
## 4 Starrocks Components Image
Official FE/CN/BE components images can be found from [dockerhub](https://hub.docker.com/u/starrocks):
Those images contains default application configuration file, they can be overritten by configuring kubernetes configmap deployment crd. 

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
The ConfigMap value is in property format, the following command is an example to generate the configmap from an existing file.
```shell
# create fe-config-map from starrocks/fe/conf/fe.conf file
kubectl create configmap fe-config-map --from-file=starrocks/fe/conf/fe.conf
```
## 5 storageVolume
External storage can be used to store FE meta and BE data for persistence. `storageVolumes` can be specified in corresponding component spec to enable external storage volumes auto provisioning. Note that the specific `storageClassName` should be available in kubernetes cluster before enabling this storageVolume feature.


**FE meta storage example**
```yaml
starRocksFeSpec:
  storageVolumes:
    - name: fe-meta
      storageClassName: meta-storage
      storageSize: 10Gi
      mountPath: /opt/starrocks/fe/meta # overwrite the default meta path
```
**BE data storage example**
```yaml
starRocksBeSpec:
  storageVolumes:
    - name: be-data
      storageClassName: data-storage
      storageSize: 1Ti
      mountPath: /opt/starrocks/be/storage # overwrite the default data path
```

## 6 Deploy Operator in kubernetes
`deploy` directory contains all the necessary yamls to deploy the operator. Yaml files with `leader_` prefix are for operator election if willing to take multiples pods for backup. `manager.yaml` is a deployment yaml to deploy operator. Remember to update corresponding `image` before applying to kubernetes. Other yamls are facilities objects created for running the operator, include namespace, service account, rbac.
By default, the operator deploys the StarRocks cluster in `starrocks` namespace. Either specifying the namespace `-n <namespace>` when running `kubectl apply` or set the namespace meta field in every yaml files.
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

## 7 Example
[examples](./examples/starrocks) directory contains some simple example for reference.
