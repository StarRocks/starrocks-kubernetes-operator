StarRocks is a next-gen, high-performance analytical data warehouse that enables real-time, multi-dimension, and
highly concurrent data analysis. This section describes:

1. How to install starrocks locally with 1 FE, 1 BE, and 1 FE Proxy.
2. How to load local data to starrocks.

> We are assuming you have a basic understanding of what the Kubernetes is and how it runs.

The following table lists the minimum and recommended hardware configurations for deploying StarRocks.

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU      | 4 CPU   | 8 CPU       |
| Mem      | 8 GB    | 16 GB       |
| Disk     | 40 GB   | 80 GB       |

The following section will cover:

1. [Prerequisites](#1-prerequisites) - The prerequisites for installing StarRocks.
2. [Installation from scratch manually](#2-installation-from-scratch-manually) - Install StarRocks with helm from
   scratch.
3. [Installation by scripts](./local_installation_how_to.md#3-installation-by-script) - Install StarRocks with scripts.
4. [Load data to StarRocks](./local_installation_how_to.md#4-load-data-to-starrocks) - Load data to StarRocks.
5. [Uninstall StarRocks](./local_installation_how_to.md#5-uninstall-starrocks) - Uninstall StarRocks.

## 1. Prerequisites

In order to install StarRocks, you need to have the following prerequisites:

1. Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool.
2. Install [helm](https://helm.sh/) command-line tool.
3. Install [docker](https://docs.docker.com/get-docker/) command-line tool.
4. Install [kind](https://kind.sigs.k8s.io/) command-line tool.

> You should be logged in as a user with sudo privileges to install the above tools.

### 1.1 Install `kubectl`

The following steps will guide you to install kubectl:

```shell
# 1. download kubectl
# Linux amd64
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
# Linux arm64
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"
# MacOS amd64
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
# MacOS arm64
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"

# 2. install kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
sudo chown root: /usr/local/bin/kubectl

# 3. check the version you installed
kubectl version --client
```

> see [install kubectl on linux](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)
> or [install kubectl on macos](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/) for more details.

### 1.2 Install `helm`

The following steps will guide you to install helm:

```shell
# 1. download helm
# Linux amd64
curl -LO https://get.helm.sh/helm-v3.12.3-linux-amd64.tar.gz
# Linux arm64
curl -LO https://get.helm.sh/helm-v3.12.3-linux-arm64.tar.gz
# MacOS amd64
curl -LO https://get.helm.sh/helm-v3.12.3-darwin-amd64.tar.gz
# MacOS arm64
curl -LO https://get.helm.sh/helm-v3.12.3-darwin-arm64.tar.gz

# 2. install helm
# e.g. Linux arm64
tar -zxvf helm-v3.12.3-linux-arm64.tar.gz 
sudo mv linux-arm64/helm /usr/local/bin/helm

# 3. check the version you installed
helm version
```

> see https://helm.sh/docs/intro/install/ for more details.

### 1.3 Install `docker`

The following steps will guide you to install docker:

**Linux**

```shell
# update your package manager
sudo apt-get update

# you can use it to get the latest docker package
curl -fsSL https://get.docker.com/ | sh

# Add your account to the docker group.
# NOTE: You will have to log out and log back in for the change to take effect.
sudo usermod -aG docker $USER

# Verify that Docker is installed by running the hello-world container
docker run hello-world
```

> see https://docs.docker.com/desktop/install/linux-install/ for more details.

**Mac**

Download and install [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/). After the
installation is complete, the Docker icon will be displayed in the menu bar.

> If installing [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/) is not allowed, you can
> use
[colima](https://github.com/abiosoft/colima) as another option.

### 1.4 Install `kind`

The following steps will guide you to install `kind`:

```bash
# download kind
# Linux AMD64 / x86_64
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
# Linux ARM64
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-arm64
# Linux Intel Macs
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-amd64
# Linux M1 / ARM Macs
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-arm64

# install kind
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# check the version you installed
kind version
```

## 2. Installation from scratch manually

### 2.1 Deploy a `kubernetes` cluster

The following steps will guide you to deploy a `kubernetes` cluster with `kind`.

```bash
# prepare a kind configuration file `kind.yaml`:
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
kind create cluster --image=kindest/node:v1.23.4 --name=starrocks --config=kind.yaml

# waiting for the new kubernetes cluster to be running healthily
```

> Set port mappings for `kind` cluster, so that you can access the `StarRocks` cluster from outside the kubernetes
> cluster.

### 2.2 Install StarRocks by helm

You can find the latest version of StarRocks Helm Chart
in https://github.com/StarRocks/starrocks-kubernetes-operator/releases.

```shell
# Add the Helm Chart Repo
helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator

# update the repo
helm repo update starrocks-community

# View the Helm Chart Repo that you added.
# There are three charts in the repo, and kube-starrocks will be used to install StarRocks Operator and StarRocks cluster.
helm search repo starrocks-community
```

see [kube-starrocks](../helm-charts/charts/kube-starrocks/README.md) chart for more details.

Prepare a `values.yaml` file to install, and you can refer
to [values.yaml](../helm-charts/charts/kube-starrocks/values.yaml) to see the default values.

```shell
cat <<EOF >values.yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.8.5

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
      - containerPort: 9030
        name: query
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
      - name: http-port
        nodePort: 30001
        containerPort: 8080
        port: 8080        
EOF

# install starrocks
helm install -n starrocks -f values.yaml starrocks starrocks-community/kube-starrocks --create-namespace

# set alias for kubectl
alias k='kubectl -n starrocks'

# waiting for the starrocks cluster to be running healthily
ubuntu@vm:~$ k get pods
NAME                                      READY   STATUS    RESTARTS   AGE
kube-starrocks-be-0                       1/1     Running   0          2m
kube-starrocks-fe-0                       1/1     Running   0          3m
kube-starrocks-fe-proxy-5c7c7fc7b-wwvs6   1/1     Running   0          3m
kube-starrocks-operator-7498c7fbd-qsbgb   1/1     Running   0          3m
```

## 3. Installation by script

Make sure `docker` is installed，See [install docker](#13-install-docker) for more details.

By default, the [script](../scripts/local-install.sh) will do the following things:

1. install kubectl, helm, kind on your machine.
2. create a kind cluster named `starrocks`.
3. install starrocks by [kube-starrocks](../helm-charts/charts/kube-starrocks/README.md) chart.

## 4. Load data to starrocks

After the starrocks cluster is running healthily, you can load data to starrocks.

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

## 5. Uninstall starrocks

```shell
# uninstall starrocks
helm uninstall starrocks -n starrocks

# uninstall kubernetes cluster
kind delete cluster --name starrocks
```
