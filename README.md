# starrocks-kubernetes-operator

## 1 Overview
**(under development)**  
develop with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), control starrocks in kubernetes

## 2 Requirements
 * kubernetes 1.18+
 * network: fe can access the ip of cn-pod

## 3 Supported Features
* start cn clusters，register cn-pod's ip to fe
* auto-scaling for cn cluster

## 3 Build images
### 3.1 Operator
```bash
# build image
make docker IMG="xxx"
# push image to docker hub
make push IMG="xxx"
```
### 3.2 Auxiliary Components
contains register-sidecar and cn-offline job
```bash
cd components
# build image
make docker IMG="xxx"
# push image to docker hub
make push IMG="xxx"
```

## 4 Deploy operator in kubernetes
```bash
cd deploy
# create crd
kubectl apply -f starrocks.com_computenodegroups.yaml
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
# replace image field with image which build in[3.1]
kubectl apply -f manager.yaml
```

## 5 Running
### 5.1 Build a cn-cluster
```bash 
cd examples/cn
# configure fe-account
kubectl apply -f fe-account.yaml
# configure cn-parameter
kubectl apply -f cn-config.yaml
# create ComputeNodeGroup
# replace component-image field with image which build in[3.2]
kubectl apply -f cn.yaml
```
## 6 Design
[docs](./doc)
