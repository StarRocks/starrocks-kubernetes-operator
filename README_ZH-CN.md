# StarRocks-Kubernetes-operator

[![许可证](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## 概述

StarRocks Kubernetes Operator 实现了在 Kubernetes 上部署和操作 StarRocks 的功能。StarRocks 是一款高性能分析型数据仓库，使用向量化、MPP
架构、CBO、智能物化视图、可实时更新的列式存储引擎等技术实现多维、实时、高并发的数据分析。Operator 便于在您的 Kubernetes 环境中部署
StarRocks 的 Frontend（FE）、Backend（BE）和计算节点（CN）组件。它还包括 Helm chart 以便于安装和配置。使用 StarRocks Kubernetes
Operator，您可以轻松管理 StarRocks 集群的生命周期，如安装、扩展、升级等。

## 先决条件

1. Kubernetes 版本 >= 1.18
2. Helm 版本 >= 3.0

## 特性

### Operator 特性

- 支持分别部署 StarRocks 的 FE、BE 和 CN 组件。FE 组件是必须的，BE 和 CN 组件可以选择性部署
- 支持在一个 Kubernetes 集群中部署多个 StarRocks 集群
- 支持外部客户端通过 STREAM LOAD 将数据加载到 StarRocks
- 支持根据 CPU 和内存使用情况自动扩展 CN 节点
- 支持为 StarRocks 容器挂载持久卷

### Helm Chart 特性

- 支持 Helm Chart 以便于安装和配置
- 使用 kube-starrocks Helm chart 同时安装 operator 和 StarRocks 集群
- 使用 operator Helm Chart 安装 operator，使用 starrocks Helm Chart 安装 starrocks 集群
- 支持在安装过程中初始化 StarRocks 集群的 root 密码
- 支持与 Kubernetes 生态系统中的其他组件集成，如 Prometheus、Datadog 等

## 安装

要在 Kubernetes 中使用 StarRocks，您需要安装：

1. StarRocksCluster CRD
2. StarRocks Operator
3. StarRocksCluster CR

有两种方式可以安装 Operator 和 StarRocks Cluster。

1. 通过 yaml Manifest 安装 Operator 和 StarRocks Cluster。
2. 通过 Helm Chart 安装 Operator 和 StarRocks Cluster。

> 注意：在每个版本中，我们都会提供最新版本的 yaml Manifest 和 Helm
> Chart。您可以在 https://github.com/StarRocks/starrocks-kubernetes-operator/releases 中找到它们。

### 通过 yaml Manifest 安装

请参阅 [使用 Operator 部署 StarRocks 文档](./doc/deploy_starrocks_with_operator_howto.md) 以获取更多详细信息。

首先，Apply 自定义资源定义 (CRD)：

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```

其次，Apply Operator manifest:

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

默认情况下，Operator 配置为在 starrocks 命名空间中安装。要在自定义命名空间中使用
Operator，下载 [Operator manifest](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml)
并编辑所有的 namespace: starrocks 以指定您的自定义命名空间。然后使用 kubectl apply -f {local-file-path} 将这个版本的
manifest 应用到集群。

最后，部署 StarRocks 集群。

您需要准备一个单独的 yaml 文件来部署 StarRocks。starrocks 集群 CRD 字段在 [api.md](./doc/api.md)
中有解释。 [examples](./examples/starrocks) 目录包含一些简单的示例供参考。 您可以使用任何模板 yaml
文件作为起点。您可以根据此部署文档将更多配置添加到模板 yaml 文件中。

为了演示目的，我们使用 [starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml) 示例模板启动一个包含 3 个
FE 和 3 个 BE 的 StarRocks 集群。

```console
kubectl apply -f starrocks-fe-and-be.yaml
```

### 通过 Helm Chart 安装

请参阅 [kube-starrocks](./helm-charts/charts/kube-starrocks/README.md) 了解如何通过 Helm Chart 安装 operator 和
StarRocks 集群。 如果您希望在管理 StarRocks
集群时有更多的灵活性，您可以使用 [operator](./helm-charts/charts/kube-starrocks/charts/operator) Helm Chart 部署
Operator，使用  [starrocks](./helm-charts/charts/kube-starrocks/charts/starrocks) Helm Chart 部署 StarRocks。

## 其它文档

- 在 [doc](./doc) 目录中，您可以找到更多关于如何使用 StarRocks Operator 的文档。
- 在 [examples](./examples/starrocks) 目录中，您可以找到更多关于如何编写 StarRocksCluster CR 的示例。
