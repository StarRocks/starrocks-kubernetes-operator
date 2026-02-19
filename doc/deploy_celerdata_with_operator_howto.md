# Deploy CelerData Cluster with Operator

This document introduces how to use the CelerData Operator to automate the deployment and management of a StarRocks
cluster on a Kubernetes cluster.

It includes the following parts:

1. Deploy CelerData Operator
2. Deploy CelerData cluster
3. Manage CelerData Cluster
    1. Access CelerData cluster
    2. Upgrade CelerData cluster
    3. Scale CelerData cluster
    4. Using ConfigMap to configure your CelerData cluster
  
> [!NOTE]  
> The StarRocks k8s operator was designed to be a level 2 operator.   See https://sdk.operatorframework.io/docs/overview/operator-capabilities/ to understand more about the capabilities of a level 2 operator. 

## Prerequisites

1. Kubernetes cluster version >= 1.23.0+
2. kubelet version >= 1.23.0+

## 1. Deploy CelerData Operator

It includes the following main steps:

1. Apply CelerDataCluster CRD.
2. Deploy CelerData Operator.

### 1.1. Apply CelerDataCluster CRD

CelerDataCluster CRD is a custom resource definition (CRD) that defines the CelerData cluster. It is used to create and
manage CelerData clusters by using the CelerData Operator. Please refer to [api.md](./api.md) for the detailed
description of the CelerDataCluster CRD.

Apply the CelerDataCluster CRD by using the following command:

```bash
kubectl apply -f https://raw.githubusercontent.com/celerdata/celerdata-kubernetes-operator/main/deploy/celerdata.com_celerdataclusters.yaml
```

### 1.2. Deploy CelerData Operator

You can choose to deploy the CelerData Operator by using a default configuration file or a custom configuration file.

1. **Deploy the CelerData Operator by using a default configuration file.**
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/celerdata/celerdata-kubernetes-operator/main/deploy/operator.yaml
   ```
   The CelerData Operator is deployed to the namespace `celerdata` and manages all CelerData clusters under all
   namespaces. After `operator.yaml` is applied, The following resources will be created:
    ```bash
    namespace/celerdata created
    serviceaccount/celerdata created
    clusterrole.rbac.authorization.k8s.io/kube-celerdata-operator created
    clusterrolebinding.rbac.authorization.k8s.io/kube-celerdata-operator created
    role.rbac.authorization.k8s.io/celerdata-leader-election-role created
    rolebinding.rbac.authorization.k8s.io/celerdata-leader-election-rolebinding created
    deployment.apps/kube-celerdata-operator created
    ```
2. **Deploy the CelerData Operator by using a custom configuration file.** By default, the Operator is configured to
   install in the starrocks namespace. To use the Operator in a custom namespace, download the Operator manifest and
   substitute all instances of namespace to your custom namespace.
    1. Download the configuration file **operator.yaml**, which is used to deploy the CelerData Operator.
       ```bash
       curl -O https://raw.githubusercontent.com/celerdata/celerdata-kubernetes-operator/main/deploy/operator.yaml
       ```
    2. Modify the configuration file **operator.yaml** to suit your needs.
    3. Deploy the CelerData Operator.
       ```bash
       kubectl apply -f operator.yaml
       ```

3. **Check the running status of the CelerData Operator.** If the pod is in the `Running` state and all containers
   inside the pod are `READY`, the CelerData Operator is running as expected.

    ```bash
    $ kubectl -n celerdata get pods
    NAME                                  READY   STATUS    RESTARTS   AGE
    celerdata-controller-65bb8679-jkbtg   1/1     Running   0          5m6s
    ```

## 2. Deploy CelerData Cluster

You need to prepare a separate yaml file to deploy the StarRocks FE, BE and CN components. You can directly use
the [sample configuration files](https://github.com/celerdata/celerdata-kubernetes-operator/tree/main/examples/celerdata)
provided by StarRocks to deploy a CelerData cluster (an object instantiated by using the custom resource StarRocks
Cluster). For example, you can use **celerdata-fe-and-be.yaml** to deploy a CelerData cluster that consists of three FE
nodes and three BE nodes.

```bash
kubectl apply -f https://raw.githubusercontent.com/celerdata/celerdata-kubernetes-operator/main/examples/celerdata/celerdata-fe-and-be.yaml
```

The following table describes a few important fields in the **celerdata-fe-and-be.yaml** file.

| **Field** | **Description**                                                                                                                                                                                                                                    |
|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Kind      | The resource type of the object. The value must be `CelerDataCluster`.                                                                                                                                                                             |
| Metadata  | Metadata, in which the following sub-fields are nested:<ul><li>`name`: the name of the object. Each object name uniquely identifies an object of the same resource type.</li><li>`namespace`: the namespace to which the object belongs.</li></ul> |
| Spec      | The expected status of the object. Valid values are `celerDataFeSpec`, `celerDataBeSpec`, and `celerDataCnSpec`.                                                                                                                                   |

You can also deploy the CelerData cluster by using a modified configuration file. For supported fields and detailed
descriptions, see [api.md](https://github.com/celerdata/celerdata-kubernetes-operator/blob/main/doc/api.md).

Deploying the CelerData cluster takes a while. During this period, you can use the
command `kubectl -n celerdata get pods` to check the starting status of the CelerData cluster. If all the pods are in
the `Running` state and all containers inside the pods are `READY`, the CelerData cluster is running as expected.

> **NOTE**
>
> If you customize the namespace in which the CelerData cluster is located, you need to replace `starrocks` with the
> name of your customized namespace.

```bash
$ kubectl -n celerdata get pods
NAME                                  READY   STATUS    RESTARTS   AGE
celerdata-controller-65bb8679-jkbtg   1/1     Running   0          22h
celerdatacluster-sample-be-0          1/1     Running   0          23h
celerdatacluster-sample-be-1          1/1     Running   0          23h
celerdatacluster-sample-be-2          1/1     Running   0          22h
celerdatacluster-sample-fe-0          1/1     Running   0          21h
celerdatacluster-sample-fe-1          1/1     Running   0          21h
celerdatacluster-sample-fe-2          1/1     Running   0          22h
```

> **Note**
>
> If some pods cannot be up after a long period of time, you can use `kubectl logs -n celerdata <pod_name>` to view the
> log information or use `kubectl -n celerdata describe pod <pod_name>` to view the event information to address the
> problem.

## 3. Manage CelerData Cluster

### 3.1. Access CelerData Cluster

The components of the CelerData cluster can be accessed through their associated Services, such as the FE Service. For
detailed descriptions of Services and their access addresses,
see [api.md](https://github.com/celerdata/celerdata-kubernetes-operator/blob/main/doc/api.md)
and [Services](https://kubernetes.io/docs/concepts/services-networking/service/).

The following table describes the FE Services of the CelerData cluster. `celerdatacluster-sample-fe-service` is the
Service that user can configure it from CelerDataCluster CR, and user should only use it to access the StarRocks.
`celerdatacluster-sample-fe-search` is the internal Service that is used by CelerData Cluster to discover the FE nodes.

```bash
$ kubectl get svc
NAME                                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                               AGE
celerdatacluster-sample-fe-search    ClusterIP   None           <none>        9030/TCP                              76s
celerdatacluster-sample-fe-service   ClusterIP   10.96.26.146   <none>        8030/TCP,9020/TCP,9030/TCP,9010/TCP   76s
```

#### 3.1.1. Access CelerData Cluster from within Kubernetes Cluster

From within the Kubernetes cluster, the CelerData cluster can be accessed through the FE Service's ClusterIP.

1. Obtain the internal virtual IP address `CLUSTER-IP` and port `PORT(S)` of the FE Service.

    ```Bash
    $ kubectl -n celerdata get svc 
    NAME                                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                               AGE
    celerdatacluster-sample-be-search    ClusterIP   None           <none>        9050/TCP                              66s
    celerdatacluster-sample-be-service   ClusterIP   10.96.86.207   <none>        9060/TCP,8040/TCP,9050/TCP,8060/TCP   66s
    celerdatacluster-sample-fe-search    ClusterIP   None           <none>        9030/TCP                              2m27s
    celerdatacluster-sample-fe-service   ClusterIP   10.96.26.146   <none>        8030/TCP,9020/TCP,9030/TCP,9010/TCP   2m27s
    ```

2. Access the CelerData cluster by using the MySQL client from within the Kubernetes cluster.

   ```Bash
   mysql -h 10.100.162.xxx -P 9030 -uroot
   ```

   Upon deploying a fresh CelerData cluster, the `root` user's password remains unset, potentially posing a security
   risk. See [Change root user password HOWTO](./change_root_password_howto.md) for details on how to set
   the `root` user's password.

#### 3.1.2. Access CelerData Cluster from outside Kubernetes Cluster by using LoadBalancer or NodePort

From outside the Kubernetes cluster, you can access the CelerData cluster through the FE Service's LoadBalancer or
NodePort. This topic uses LoadBalancer as an example:

1. Run the command `kubectl -n celerdata edit cdc celerdatacluster-sample` to update the CelerData cluster configuration
   file, and add `service` field to the `celerDataFeSpec` field.

    ```YAML
    spec:
      celerDataFeSpec:
        service:            
          type: LoadBalancer # specified as LoadBalancer
    ```

2. Obtain the IP address `EXTERNAL-IP` and port `PORT(S)` that the FE Service exposes to the outside.

    ```Bash
    $ kubectl -n celerdata get svc
    NAME                                 TYPE           CLUSTER-IP       EXTERNAL-IP                                                              PORT(S)                                                       AGE
    celerdatacluster-sample-be-search    ClusterIP      None           <none>        9050/TCP                                                      6m39s
    celerdatacluster-sample-be-service   ClusterIP      10.96.86.207   <none>        9060/TCP,8040/TCP,9050/TCP,8060/TCP                           6m39s
    celerdatacluster-sample-fe-search    ClusterIP      None           <none>        9030/TCP                                                      8m
    celerdatacluster-sample-fe-service   LoadBalancer   10.96.26.146   a7509284bf3784983a596c6eec7fc212-618xxxxxx.us-west-2.elb.amazonaws.com     8030:30028/TCP,9020:32241/TCP,9030:32640/TCP,9010:32384/TCP   8m
    ```

3. Log in to your machine host and access the CelerData cluster by using the MySQL client.

    ```Bash
    mysql -h a7509284bf3784983a596c6eec7fc212-618xxxxxx.us-west-2.elb.amazonaws.com -P9030 -uroot
    ```

#### 3.1.3. Access CelerData Cluster from outside Kubernetes Cluster by port forwarding

From outside the Kubernetes cluster, you can access the CelerData cluster through the FE Service's port forwarding.

1. Make sure that you have installed the `kubectl` command-line tool and configured access to the Kubernetes cluster.
2. Run the command `kubectl -n celerdata port-forward service/celerdatacluster-sample-fe-service 9030:9030` to forward
   local port `9030` to FE Service's port `9030`.
3. Access the CelerData cluster by using the MySQL client.
    ```Bash
    mysql -h 127.0.0.1 -P9030 -uroot
    ```

### 3.2. Upgrade CelerData Cluster

#### 3.2.1. Upgrade BE nodes

Run the following command to specify a new BE image file, such as `us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/be-ubuntu:latest`:

```bash
kubectl -n celerdata patch celerdatacluster celerdatacluster-sample --type='merge' -p '{"spec":{"celerDataBeSpec":{"image": us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/be-ubuntu:latest"}}}'
```

#### 3.2.2. Upgrade FE nodes

Run the following command to specify a new FE image file, such as `us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/fe-ubuntu:latest`:

```bash
kubectl -n celerdata patch celerdatacluster celerdatacluster-sample --type='merge' -p '{"spec":{"celerDataFeSpec":{"image": us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/fe-ubuntu:latest"}}}'
```

The upgrade process lasts for a while. You can run the command `kubectl -n celerdata get pods` to view the upgrade
progress.

### 3.3. Scale CelerData cluster

This topic takes scaling out the BE and FE clusters as examples.

#### 3.3.1. Scale out BE cluster

Run the following command to scale out the BE cluster to 9 nodes:

```bash
kubectl -n celerdata patch celerdatacluster celerdatacluster-sample --type='merge' -p '{"spec":{"celerDataBeSpec":{"replicas":9}}}'
```

#### 3.3.2. Scale out FE cluster

Run the following command to scale out the FE cluster to 4 nodes:

```bash
kubectl -n celerdata patch celerdatacluster celerdatacluster-sample --type='merge' -p '{"spec":{"celerDataFeSpec":{"replicas":4}}}'
```

The scaling process lasts for a while. You can use the command `kubectl -n celerdata get pods` to view the scaling
progress.

**Add cautions on scale-in FE nodes**:

FE nodes can be scaled-in, but there are some limitations:

1. FE nodes can only be scaled-in step by step. If the last scale-in operation is not completed, the next scale-in
   operation cannot be performed.
2. Each time less than half of the nodes can be scaled-in.
3. You can't do 3->1 scale in.

### 3.4. Using ConfigMap to configure your CelerData cluster

The official images contains default application configuration file, however, they can be overwritten by configuring
kubernetes configmap deployment crd.

You can generate the configmap from an StarRocks configuration file.
Below is an example of creating a Kubernetes configmap `fe-config-map` from the `fe.conf` configuration file. You can do
the same with BE and CN.

```console
# create fe-config-map from starrocks/fe/conf/fe.conf file
kubectl create configmap fe-config-map --from-file=starrocks/fe/conf/fe.conf
```

Once the configmap is created, you can reference the configmap in the yaml file.
For example:

```yaml
# fe use configmap example
celerDataFeSpec:
  configMapInfo:
    configMapName: fe-config-map
    resolveKey: fe.conf
# cn use configmap example
celerDataCnSpec:
  configMapInfo:
    configMapName: cn-config-map
    resolveKey: cn.conf
  # be use configmap example
  celerDataBeSpec:
    configMapInfo:
    configMapName: be-config-map
    resolveKey: be.conf
```

## FAQ

**Issue description:** When a custom resource CelerDataCluster is installed using `kubectl apply -f xxx`, an error is
returned `The CustomResourceDefinition 'celerdataclusters.celerdata.com' is invalid: metadata.annotations: Too long: must have at most 262144 bytes`.

**Cause analysis:** Whenever `kubectl apply -f xxx` is used to create or update resources, a metadata
annotation `kubectl.kubernetes.io/last-applied-configuration` is added. This metadata annotation is in JSON format and
records the *last-applied-configuration*. `kubectl apply -f xxx`" is suitable for most cases, but in rare situations ,
such as when the configuration file for the custom resource is too large, it may cause the size of the metadata
annotation to exceed the limit.

**Solution:** If you install the custom resource CelerDataCluster for the first time, it is recommended to
use `kubectl create -f xxx`. If the custom resource CelerDataCluster is already installed in the environment, and you
need to update its configuration, it is recommended to use `kubectl replace -f xxx`.
