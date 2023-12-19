StarRocks 是一款高性能分析型数据仓库，使用向量化、MPP 架构、CBO、智能物化视图、可实时更新的列式存储引擎等技术实现多维、实时、高并发的数据分析。

本节将介绍：

1. 如何在本地安装包含 1 个 FE、1 个 BE 和 1 个 FE Proxy 的 StarRocks。
2. 如何将本地数据加载到 StarRocks。

> 我们假设你已经对 Kubernetes 有基本的了解。

下表列出了部署 StarRocks 的最小和推荐的硬件配置。

| 资源  | 最小值   | 推荐值   |
|-----|-------|-------|
| CPU | 4 CPU | 8 CPU |
| 内存  | 8 GB  | 16 GB |
| 磁盘  | 40 GB | 80 GB |

以下部分将涵盖：

1. [前提条件](./local_installation_how_to.md#1-前提条件) - 安装 StarRocks 的前提条件。
2. [手动从零开始安装](./local_installation_how_to.md#2-从零开始手动安装) - 从零开始使用 helm 安装 StarRocks。
3. [通过脚本安装](./local_installation_how_to.md#3-通过脚本安装) - 使用脚本安装 StarRocks。
4. [将数据加载到 StarRocks](./local_installation_how_to.md#4-将数据加载到-starrocks) - 将数据加载到 StarRocks。
5. [卸载 StarRocks](./local_installation_how_to.md#5-卸载-starrocks) - 卸载 StarRocks。

## 1. 前提条件

为了安装 StarRocks，你需要满足以下前提条件：

1. 安装 [docker](https://docs.docker.com/get-docker/)
2. 安装 [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
3. 安装 [helm](https://helm.sh/)
4. 安装 [kind](https://kind.sigs.k8s.io/)

> 你应该以具有 sudo 权限的用户身份登录，以安装上述工具。

### 1.1 安装 `docker`

以下步骤将指导你安装 docker：

以 **Ubuntu** 为例

```shell
# 查看 https://developer.aliyun.com/article/872508 了解如何安装 docker 的详细信息。

# 安装基本软件
sudo apt-get update
sudo apt-get install apt-transport-https ca-certificates curl software-properties-common lrzsz -y

# 使用阿里云的源
sudo curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"

# 安装 docker
sudo apt-get update
sudo apt-get install docker-ce -y
```

> 更多信息请查看 https://docs.docker.com/desktop/install/linux-install/ .

以**Mac**为例

下载并安装 [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)。

> 如果不允许安装 [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)，你可以
> 使用 [colima](https://github.com/abiosoft/colima) 作为另一个选项。

### 1.2 安装 `kubectl`

以下步骤将指导你安装 kubectl：

```shell
# 设置 URL
# 默认是官方的 URL 地址，如果你无法访问下面的 url，可以尝试使用下面的 url。
# KUBECTL_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
KUBECTL_URL="https://dl.k8s.io/release/v1.28.3/bin"

# 1. download kubectl
# Linux amd64
curl -LO "$KUBECTL_URL/linux/amd64/kubectl"
# Linux arm64
curl -LO "$KUBECTL_URL/linux/arm64/kubectl"
# MacOS amd64
curl -LO "$KUBECTL_URL/darwin/amd64/kubectl"
# MacOS arm64
curl -LO "$KUBECTL_URL/darwin/arm64/kubectl"

# 2. install kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

# 3. check the version you installed
kubectl version --client
```

> 更多信息请查看：[install kubectl on linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
> 或 [install kubectl on macos](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/)。

### 1.3 安装 `helm`

以下步骤将指导你安装 helm：

```shell
# 设置 URL
# 默认是官方的 URL 地址，如果你无法访问下面的 url，可以尝试使用下面的 url。
# HELM_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
HELM_URL="https://get.helm.sh"

# 1. download helm
# Linux amd64
curl -LO $HELM_URL/helm-v3.12.3-linux-amd64.tar.gz
# Linux arm64
curl -LO $HELM_URL/helm-v3.12.3-linux-arm64.tar.gz
# MacOS amd64
curl -LO $HELM_URL/helm-v3.12.3-darwin-amd64.tar.gz
# MacOS arm64
curl -LO $HELM_URL/helm-v3.12.3-darwin-arm64.tar.gz

# 2. install helm
# e.g. Linux arm64
tar -zxvf helm-v3.12.3-linux-amd64.tar.gz 
sudo mv linux-amd64/helm /usr/local/bin/helm

# 3. check the version you installed
helm version
```

### 1.4 安装 `kind`

以下步骤将指导你安装 kind：

```shell
# 设置 URL
# 默认是官方的 URL 地址，如果你无法访问下面的 url，可以尝试使用下面的 url。
# KIND_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
KIND_URL="https://kind.sigs.k8s.io/dl/v0.20.0"

# Linux AMD64 / x86_64
curl -Lo ./kind $KIND_URL/kind-linux-amd64
# Linux ARM64
curl -Lo ./kind $KIND_URL/kind-linux-arm64
# Linux Intel Macs
curl -Lo ./kind $KIND_URL/kind-darwin-amd64
# Linux M1 / ARM Macs
curl -Lo ./kind $KIND_URL/kind-darwin-arm64

# install kind
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# check the version you installed
kind version
```

## 2. 从零开始手动安装

### 2.1 通过 Kind 安装 Kubernetes 集群

在安装了上述工具后，你可以使用以下步骤手动安装 StarRocks。

```bash
# prepare a kind configuration file `kind.yaml`:
# 为 `kind` 集群设置端口映射，这样你就可以从 kubernetes 集群外部访问 `StarRocks` 集群了。
cat <<EOF > kind.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30001
    hostPort: 30001
    listenAddress: "0.0.0.0"
    protocol: TCP
  - containerPort: 30002
    hostPort: 30002
    listenAddress: "0.0.0.0"
    protocol: TCP
EOF

# kubernetes installation
unset http_proxy
unset https_proxy
# 你可能需要 sudo 权限来执行 kind create cluster 命令
sudo kind create cluster --image=kindest/node:v1.23.4 --name=starrocks --config=kind.yaml

# 你可能需要 sudo 权限来执行kubectl命令
# waiting for the new kubernetes cluster to be running healthily
sudo kubectl get pods -A
NAMESPACE            NAME                                              READY   STATUS    RESTARTS   AGE
kube-system          coredns-64897985d-brlj9                           1/1     Running   0          4m26s
kube-system          coredns-64897985d-m9kj6                           1/1     Running   0          4m26s
kube-system          etcd-starrocks-control-plane                      1/1     Running   0          4m39s
kube-system          kindnet-jsrg8                                     1/1     Running   0          4m26s
kube-system          kube-apiserver-starrocks-control-plane            1/1     Running   0          4m39s
kube-system          kube-controller-manager-starrocks-control-plane   1/1     Running   0          4m39s
kube-system          kube-proxy-8h6b4                                  1/1     Running   0          4m26s
kube-system          kube-scheduler-starrocks-control-plane            1/1     Running   0          4m39s
local-path-storage   local-path-provisioner-5ddd94ff66-9l2km           1/1     Running   0          4m26s
```

### 2.2 通过 Helm 安装 StarRocks

你可以在 https://github.com/StarRocks/starrocks-kubernetes-operator/releases 获取 StarRocks Helm Chart 的最新版本。

```shell
# Add the Helm Chart Repo
helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator

# update the repo
helm repo update starrocks-community

# View the Helm Chart Repo that you added.
# There are three charts in the repo, and kube-starrocks will be used to install StarRocks Operator and StarRocks cluster.
helm search repo starrocks-community
```

这里有 [kube-starrocks](../../helm-charts/charts/kube-starrocks/README.md) Helm Chart 的更多信息。

Prepare a `values.yaml` file to install, and you can refer
to [values.yaml](../../helm-charts/charts/kube-starrocks/values.yaml) to see the default values.

```shell
cat <<EOF >values.yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.8.6

starrocks:
  starrocksFESpec:
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.1.2
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 500m
        memory: 1Gi
    service:
      type: NodePort
      ports:
      - name: query   # fill the name from the fe service ports
        containerPort: 9030
        nodePort: 30002
        port: 9030
  starrocksBeSpec:
    image:
      repository: starrocks/be-ubuntu
      tag: 3.1.2
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 500m
        memory: 1Gi
  starrocksFeProxySpec:
    enabled: true
    resources:
      requests:
        cpu: 100m
        memory: 200Mi
    service:
      type: NodePort
      ports:
      - name: http-port   # fill the name from the fe proxy service ports
        nodePort: 30001
        containerPort: 8080
        port: 8080        
EOF

# install starrocks
# 如果你因为网络问题无法安装，可参见后面的操作
helm install -n starrocks starrocks -f values.yaml starrocks-community/kube-starrocks --create-namespace

############## 如果你因为网络问题无法直接安装(helm instal)，可以先下载再安装 ##############
# 如果无法下载可以使用下面的 URL
# HELM_CHART_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
HELM_CHART_URL="https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.8.6"
curl -LO "$HELM_CHART_URL/kube-starrocks-1.8.6.tgz"

helm install -n starrocks starrocks -f values.yaml ./kube-starrocks-1.8.6.tgz --create-namespace
#############################################################

# set alias for kubectl
alias k='kubectl -n starrocks'

# waiting for the starrocks cluster to be running healthily
ubuntu@vm:~$ k get pods
NAME READY STATUS RESTARTS AGE
kube-starrocks-be-0 1/1 Running 0 2m
kube-starrocks-fe-0 1/1 Running 0 3m
kube-starrocks-fe-proxy-5c7c7fc7b-wwvs6 1/1 Running 0 3m
kube-starrocks-operator-7498c7fbd-qsbgb 1/1 Running 0 3m
```

## 3. 通过脚本安装

请确保 `docker` 已经安装，更多信息请查看 [install docker](./local_installation_how_to.md#11-安装-docker)。

默认情况下，[script](../../scripts/local-install.sh) 将会做以下事情：

1. 在你的机器上安装 kubectl、helm、kind。
2. 创建一个名为 `starrocks` 的 kind 集群。
3. 通过 [kube-starrocks](../../helm-charts/charts/kube-starrocks/README.md) chart 安装 starrocks。

执行下面的命令：

```shell
sudo bash local-install.sh --helm-url https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com --kind-url https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com --kubectl-url https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com --helm-chart-url https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com/kube-starrocks-1.8.6.tgz
```

## 4. 将数据加载到 StarRocks

在 starrocks 集群运行正常后，你可以将数据加载到 starrocks。

```shell
# set alias for kubectl
alias k='kubectl -n starrocks'

# get the pod ip of kube-starrocks-fe-0
IP=$(k get pod kube-starrocks-fe-0 --template '{{.status.podIP}}')

# login to kube-starrocks-fe-0
k exec -it kube-starrocks-fe-0 -- env IP=$IP bash

# create database and table
mysql -h $IP -P 9030 -uroot

create database test_db;
use test_db;
CREATE TABLE `table1`
(
    `id` int(11) NOT NULL COMMENT "用户 ID",
    `name` varchar(65533) NULL COMMENT "用户姓名",
    `score` int(11) NOT NULL COMMENT "用户得分"
)
ENGINE=OLAP
PRIMARY KEY(`id`)
DISTRIBUTED BY HASH(`id`)
PROPERTIES ("replication_num" = "1");

# Note: back to your host, not FE
# create csv file
cat <<EOF > example1.csv
1,Lily,23
2,Rose,23
3,Alice,24
4,Julia,25
EOF

# load data to starrocks
curl --location-trusted -u root:"" -H "label:123" \
    -H "Expect:100-continue" \
    -H "column_separator:," \
    -H "columns: id, name, score" \
    -T example1.csv -XPUT \
    http://localhost:30001/api/test_db/table1/_stream_load
```

## 5. 卸载 StarRocks

```shell
# uninstall starrocks
helm uninstall starrocks -n starrocks

# uninstall kubernetes cluster
kind delete cluster --name starrocks
```
