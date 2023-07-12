#! /bin/bash

# this script is used to install starrocks-operator by kind
# you should also install kubectl and plugins, including kubectl-neat, kubectl-ns.

# check the number of arguments
if [ $# -ne 2 ]; then
  echo "Usage: $0 <release_tag> <fe_be_cn_tag>"
  exit 1
fi
release_tag=$1
fe_be_cn_tag=$2

# check if the cluster exists
kind get clusters | grep starrocks >/dev/null
if [ $? -eq 0 ]; then
  echo "The cluster starrocks already exists, please delete it first."
else
  # delete the old cluster if exists
  kind delete cluster --name starrocks

  # create a new kubernetes cluster
  unset http_proxy
  unset https_proxy
  kind create cluster --name starrocks
fi

# download yaml files of crd and operator
curl -LO https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}/operator.yaml >operator.yaml
curl -LO https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}/starrocks.com_starrocksclusters.yaml >starrocks.com_starrocksclusters.yaml

# install crd
kubectl apply -f starrocks.com_starrocksclusters.yaml

# install operator
# NOTE:
#   1. the imagePullPolicy is set to IfNotPresent to avoid pulling image from docker hub
#   2. change the image to the related version image, not latest
kubectl apply -f operator.yaml

# prepare for images
docker images starrocks/fe-ubuntu | grep ${fe_be_cn_tag} >/dev/null
if [ $? -ne 0 ]; then
  docker pull starrocks/fe-ubuntu:${fe_be_cn_tag}
fi
kind load docker-image starrocks/fe-ubuntu:${fe_be_cn_tag} --name starrocks

docker images starrocks/be-ubuntu | grep ${fe_be_cn_tag} >/dev/null
if [ $? -ne 0 ]; then
  docker pull starrocks/be-ubuntu:${fe_be_cn_tag}
fi
kind load docker-image starrocks/be-ubuntu:${fe_be_cn_tag} --name starrocks

docker images starrocks/cn-ubuntu | grep ${fe_be_cn_tag} >/dev/null
if [ $? -ne 0 ]; then
  docker pull starrocks/cn-ubuntu:${fe_be_cn_tag}
fi
kind load docker-image starrocks/cn-ubuntu:${fe_be_cn_tag} --name starrocks

# install starrocks cluster
kubectl-ns starrocks
kubectl apply -f ./cluster/starrockscluster_sample.yaml

## sleep for waiting operator to reconcile
#sleep 20
#
## collect resources
#version=$1
#mkdir -p "./${version}"
#cd "./${version}"
#cluster=$(kubectl get starrockscluster | grep sample | awk '{print $1}')
#kubectl get sts ${cluster}-be -oyaml | kubectl-neat >./${cluster}-be.yaml
#kubectl get sts ${cluster}-fe -oyaml | kubectl-neat >./${cluster}-fe.yaml
#kubectl get sts ${cluster}-cn -oyaml | kubectl-neat >./${cluster}-cn.yaml
#kubectl get svc ${cluster}-be-search -oyaml | kubectl-neat >./${cluster}-be-search.yaml
#kubectl get svc ${cluster}-be-service -oyaml | kubectl-neat >./${cluster}-be-service.yaml
#kubectl get svc ${cluster}-cn-search -oyaml | kubectl-neat >./${cluster}-cn-search.yaml
#kubectl get svc ${cluster}-cn-service -oyaml | kubectl-neat >./${cluster}-cn-service.yaml
#kubectl get svc ${cluster}-fe-search -oyaml | kubectl-neat >./${cluster}-fe-search.yaml
#kubectl get svc ${cluster}-fe-service -oyaml | kubectl-neat >./${cluster}-fe-service.yaml
#kubectl get hpa ${cluster}-cn-autoscaler -oyaml | kubectl-neat >./${cluster}-cn-autoscaler.yaml
