# StarRock 的日志及相关配置

StarRocks 包含了三个组件，FE、BE、CN，其中，FE 是必须的组件。这篇文档将介绍 StarRocks 组件的日志及相关配置。主要包括：

1. [日志的存储位置](./logging_and_related_configurations_howto.md#1-日志的存储位置)；
2. [默认存储介质](./logging_and_related_configurations_howto.md#2-默认存储介质)；
3. [持久化日志](./logging_and_related_configurations_howto.md#3-持久化日志)；
4. [将日志打印到控制台](./logging_and_related_configurations_howto.md#4-将日志打印到控制台)；
5. [将日志收集到 datadog](././logging_and_related_configurations_howto.md#5-将日志收集到-datadogs)。

## 1. 日志的存储位置

1. FE 组件的日志存储在：`/opt/starrocks/fe/log` 目录下，关键日志文件包括：`fe.out`, `fe.log`, `fe.warn.log`。
2. BE 组件的日志存储在：`/opt/starrocks/be/log` 目录下，关键日志文件包括：`be.out`, `be.INFO`, `be.WARNING`。
3. CN 组件的日志存储在：`/opt/starrocks/cn/log` 目录下，关键日志文件包括：`cn.out`, `cn.INFO`, `cn.WARNING`。

## 2. 默认存储介质

默认情况下，所有组件都采用 `emptyDir` 存储卷。这带来的一个问题是：当 Pod 重启后，无法查看到重启前的 Pod 的日志， 导致不利于排查问题。
下面提供了两种方法来解决该问题：

1. [持久化日志](./logging_and_related_configurations_howto.md#3-持久化日志)。当 Pod 重启后，Pod 之前的日志仍然存在。
2. [将日志打印到控制台](./logging_and_related_configurations_howto.md#4-将日志打印到控制台)。通过 `kubectl logs my-pod -p` 查看重启前的 Pod 的日志。

## 3. 持久化日志

在所有组件的Spec定义中，都存在`storageVolumes`字段，以允许用户自定义存储卷。以 FE 为例：

```yaml
spec:
  starRocksFeSpec:
    storageVolumes:
    - mountPath: /opt/starrocks/fe/log
      name: fe-storage-log
      storageSize: 10Gi
      storageClassName: ""
```

`storageClassName` 如果为空，Kubernetes 将使用默认的存储卷类型。你也可以通过 `kubectl get storageclass` 查看 Kubernetes
集群内可使用的存储卷类型的列表。**注意：选择一个合适的存储卷类型非常重要。**
详情参见：https://kubernetes.io/zh-cn/docs/concepts/storage/persistent-volumes/#types-of-persistent-volumes
> 注意：Operator 会为 StarRocks 集群创建 PVC 资源，storageclass 的控制器会自动地创建具体的存储卷。

### 3.1 Helm Chart 支持持久化日志

如果使用了 Helm Chart 的方式部署 StarRocks 集群，可以通过修改 `values.yaml` 来持久化日志。以 FE 组件为例：

针对 kube-starrocks Helm Chart，可以这样配置：

```yaml
starrocks:
  starrocksFESpec:
    storageSpec:
      name: "fe-storage"
      storageSize: 10Gi
      logStorageSize: 10Gi
      storageClassName: ""
```

针对 starrocks Helm Chart，可以这样配置：

```yaml
starrocksFESpec:
  storageSpec:
    name: "fe-storage"
    storageSize: 10Gi
    logStorageSize: 10Gi
    storageClassName: ""
```

> 注意：在 FE 中，`storageSize` 指定了元数据的存储卷的大小，`logStorageSize` 指定了日志的存储卷的大小。

## 4. 将日志打印到控制台

通过为组件设置环境变量`LOG_CONSOLE = 1` 可以将组件的日志打印到控制台上。以 FE 为例：

```yaml
spec:
  starRocksFeSpec:
    feEnvVars:
    - name: LOG_CONSOLE
      value: "1"
```

### 4.1 Helm Chart 支持设置环境变量

如果使用了 Helm Chart 的方式部署 StarRocks 集群，可以修改 `values.yaml` 的内容，来设置环境变量。下面以 FE 组件为例：

针对 kube-starrocks Helm Chart，配置如下：

```yaml
starrocks:
  starrocksFESpec:
    feEnvVars:
    - name: LOG_CONSOLE
      value: "1"
```

针对 starrocks Helm Chart，配置如下：

```yaml
starrocksFESpec:
  feEnvVars:
  - name: LOG_CONSOLE
    value: "1"
```

## 5. 将日志收集到 datadog

参见：[Datadog](../integration/integration-with-datadog.md)
