# starrocks-kubernetes-operator

## 1 Overview
**(under development)**  
develop with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), deploy starrocks in kubernetes,
operator can only deploy fe, also deploy be and cn. now operator only support deploy starrocks by FQDN mode. for storage when you don't clear statement storagevolume info, 
the operator will use emptydir mode for save fe meta and be data. how to use storageVolume, please reference the storageVolume.

## 2 Requirements
 * kubernetes 1.18+
 * golang 1.18+

## 3 Supported Features
* fe decouple with cn and be. you can only deploy fe or cn and be.
* support v2beta2 horizontalpodautoscalers for cn cluster.

## 3 Build operator images
the starrocks supported operator image, you can download from :[operator](https://hub.docker.com/repository/docker/fushilei/starrocks.operator/tags?page=1&ordering=last_updated) , if you want to build image by yourself, you can do it as follows:

in the root directory of project, you can compile operator by make command,
if you want build by yourself, you should prepare go environment.
before build image, you should execute 'make build' command for build operator. 
after build, you can build image and push image to your repository. the ops as follows:

```bash
# compile operator
make build 
# build docker image
make docker IMG="xxx"
# push image to docker hub
make push IMG="xxx"
```
## 4 starrocks image
the fe, cn be components you can find from [dockerhub](https://hub.docker.com/u/starrocks):
the image have default config, you can set your config with kubernetes configmap and specify configmap in deployment crd. examples:
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
the configmap value are property format, you can use starrocks config file generate.
```shell
kubectl create configmap fe-config-map --from-file=starrocks/fe/conf/fe.conf
```
## 5 storageVolume
when you specify the volume for storage fe meta or be data, you should configure the storageVolume in component spec and prepare create storageclass in kubernetes.
for fe meta storage example:
```yaml
starRocksFeSpec:
  storageVolumes:
    - name: fe-meta
      storageClassName: meta-storage
      storageSize: 10Gi
      mountPath: /opt/starrocks/fe/meta # is stable path
```
for be data storage example:
```yaml
starRocksBeSpec:
  storageVolumes:
    - name: be-data
      storageClassName: data-storage
      storageSize: 1Ti
      mountPath: /opt/starrocks/be/storage # is stable path
```

## 5 Deploy operator in kubernetes
the directory of deploy have deployment yaml.
the start with leader yamls are for operator election when have multi pods for backup.
the manager.yaml is the deployment crd yaml for deploy operator, before you create or apply, you should update the image field.
other yamls are for operator running.
the operator deploy in starrocks namespace, if you want to deploy in another namespace, you should modify the yamls in deploy.
the follows display the resources create or apply.
```bash
cd deploy
# create crd
kubectl apply -f starrocks.com_starrocksclusters.yaml
# create namespace
kubectl apply -f namespace.yam;
# create rbac-roles
kubectl apply -f leader_election_role.yaml
kubectl apply -f role.yaml
# create rbac-role-binding
kubectl apply -f role_binding.yaml
kubectl apply -f leader_election_role_binding.yaml
# create rbac-service-account
kubectl apply -f service_account.yaml
# create operator deployment
# replace image field with image which build in[3]
kubectl apply -f manager.yaml
```

## 6 example
the directory of examples has some simple example for deploy starrocks.

