# StarRocks-Kubernetes-operator

[![许可证](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## 概述

本文档描述了如何使用StarRocks Operator在Kubernetes环境中部署StarRocks集群。更多文档请查看 [doc](./doc/README.md) 目录。

## 要求

* kubernetes 1.18+
* golang 1.18+

## 支持的功能

* FE与CN和BE解耦。FE是必须的组件，BE和CN可选部署。
* CN集群支持v2版本的HPA。

## 安装operator

为operator应用自定义资源定义 (CRD)：

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```

应用operator清单。默认情况下，operator配置为在starrocks命名空间中安装。如果您要在自定义命名空间中使用operator，请下载 [operator yaml](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml)
并编辑所有的命名空间，然后，使用`kubectl apply -f {local-file-path}`命令，而不是下面的命令，将这个版本的清单应用到集群中。

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## 部署StarRocks集群

您需要准备一个独立的yaml文件，以部署StarRocks的FE、BE和CN组件。StarRocksCluster CRD 的字段含义在[api.md](./doc/api.md)
中解释。[examples](./examples/starrocks) 目录包含了一些简单的参考示例。

为了演示目的，我们使用[starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml)
示例模板启动一个3FE和3BE的StarRocks集群。

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/examples/starrocks/starrocks-fe-and-be.yaml
```

## 连接到部署的StarRocks集群

部署StarRocks集群后，您可以使用`kubectl get svc -n <namespace>`来查找连接的IP。例如，如果starrocks部署在`starrocks`
命名空间中，您可以：

```console
kubectl get svc -n starrocks
```

使用`<your-StarRocksCluster-name>-fe-service`的clusterIP作为连接到StarRocks FE的IP。

## 停止StarRocks集群

删除自定义资源：

```console
kubectl delete -f starrocks-fe-and-be.yaml
```

删除operator：

```console
kubectl delete -f  https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## 使用Helm安装StarRocks

StarRocks已经采用Helm进行部署需求。您可以在[artifacthub](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks)
找到StarRocks的Helm图表。

更多详情，请查看[deploy_starrocks_with_helm.md](./doc/deploy_starrocks_with_helm_howto.md)。
