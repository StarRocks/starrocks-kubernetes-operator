# starrocks-kubernetes-operator

## 1 项目介绍
使用 [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) 构建，用于在k8s中管控starrocks。

## 2 环境要求
 * k8s: 1.18及以上
 * 网络: fe可访问cn-pod的ip

## 3 特性
* 启动cn集群，并向fe注册节点
* cn节点弹性扩缩

## 3 构建镜像
### 3.1 operator
```bash
# 生成镜像
make docker IMG="xxx"
# 推送镜像
make push IMG="xxx"
```
### 3.2 辅助组件
目前包含cn的注册sidecar和清理下线实例的job
```bash
cd components
# 生成镜像
make docker IMG="xxx"
# 推送镜像
make push IMG="xxx"
```

## 4 operator部署
```bash
cd deploy
# 创建crd
kubectl apply -f starrocks.com_computenodegroups.yaml
# 创建namespace
kubectl apply -f namespace.yam;
# 创建rbac roles
kubectl apply -f leader_election_role.yaml
kubectl apply -f role.yaml
# 创建rbac role binding
kubectl apply -f role_binding.yaml
kubectl apply -f leader_election_role_binding.yaml
# 创建rbac service account
kubectl apply -f service_account.yaml
# 部署operator镜像
# 注意将yaml中的镜像替换成[3.1]中构建的镜像
kubectl apply -f manager.yaml
```

## 5 运行
### 5.1 构建一个cn集群
```bash 
cd examples/cn
# 配置连接fe的账号密码
kubectl apply -f fe-account.yaml
# 配置cn节点参数
kubectl apply -f cn-config.yaml
# 创建cn集群
# 注意将yaml中的辅助组件镜像替换成[3.2]中构建的镜像
kubectl apply -f cn.yaml
```
## 6 详细设计
参考[设计文档](./doc)
