# cn

ComputeNode（cn）是starrocks的计算节点，不带本地存储数据，
可以执行除了OlapScanNode 和 OlapTableSink之外的查询计划。starrocks-operator即是基于K8s云原生技术，实现对ComputeNode节点的管控

## 名词解释
1. Pod: K8s中创建和管理的、最小的可部署的计算单元，包括一组（一个或多个）容器，这些容器共享存储、网络、以及怎样运行这些容器的声明
2. Deployment: 一种k8s原生资源，主要用于管控无状态服务
3. CR：CustomResource，k8s中用户自定义的一种资源
4. HPA：一种k8s原生资源，用于控制业务负载进行伸缩
5. CronJob：一种k8s原生资源，可以定时执行任务
6. Operator：基于上述的CustomResource，并通过自己编写控制器（Custom Controller）来将特定于某应用的领域知识组织起来，以编码的形式构造对 Kubernetes API 的扩展


## k8s资源

1. ComputeNodeGroup CR
2. cn-operator Deployment
3. CN Deployment (CN、RegisterSidecar)
4. Offline CronJob
5. HPA
6. ClusterRoleBinding
7. ServiceAccount

## operator主要功能及实现

### ComputeNodeGroup创建

* 根据CR的描述渲染出与该cr相关的资源(deployment, cronJob, hpa, clusterRoleBinding, serviceAccount)
* 设置finalizers, 用于删除时做一些清理操作

### ComputeNodeGroup删除

* 将deployment的副本数设为0，pod全部被删除后，去fe节点中将alive=false的cn节点摘除
* 清理finalizers

### ComputeNodeGroup扩容

会去修改deployment的副本数，每个CN pod 会绑定一个RegisterSidecar，通过RegisterSidecar将该CN节点信息向FE节点注册

### ComputeNodeGroup缩容

会去修改deployment的副本数，Offline CronJob 会定期执行任务，将现存的cn pod与fe中的信息进行比对，把不存在的cn从fe上摘除

### 弹性伸缩

通过CR中的hpa字段描述集群弹性伸缩的策略

## TODO

- node表需增加labels，在多个集群的场景下，可以把一个cr(一个deployment)标记为一个集群
