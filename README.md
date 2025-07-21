# StarRocks-Kubernetes-Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> English | [中文](README_ZH-CN.md)

## Overview

StarRocks Kubernetes Operator is a project that implements the deployment and operation of StarRocks, a next-generation
sub-second MPP OLAP database, on Kubernetes. It facilitates the deployment of StarRocks' Frontend (FE), Backend (BE),
and Compute Node (CN) components within your Kubernetes environment. It also includes Helm chart for easy installation
and configuration. With StarRocks Kubernetes Operator, you can easily manage the lifecycle of StarRocks clusters, such
as installing, scaling, upgrading etc.

> [!NOTE]  
> The StarRocks k8s operator was designed to be a level 2 operator.   See https://sdk.operatorframework.io/docs/overview/operator-capabilities/ to understand more about the capabilities of a level 2 operator. 

## Prerequisites

1. Kubernetes version >= 1.18
2. Helm version >= 3.0

## Features

### Operator Features

- Support deploying StarRocks FE, BE and CN components separately
  FE component is a must-have component, BE and CN components can be optionally deployed
- Support multiple StarRocks clusters in one Kubernetes cluster
- Support external clients outside the network of kubernetes to load data into StarRocks using STREAM LOAD
- Support automatic scaling for CN nodes based on CPU and memory usage
- Support mounting persistent volumes for StarRocks containers
- **Support PVC expansion for storage volumes without data loss** (New in v1.9.9+)

### Helm Chart Features

- Support Helm Chart for easy installation and configuration
    - using kube-starrocks Helm chart to install both operator and StarRocks cluster
    - using operator Helm Chart to install operator, and using starrocks Helm Chart to install starrocks cluster
- Support initializing the password of root in your StarRocks cluster during installation.
- Support integration with other components in the Kubernetes ecosystem, such as Prometheus, Datadog, etc.

## Installation

In order to use StarRocks in Kubernetes, you need to install:

1. StarRocksCluster CRD
2. StarRocks Operator
3. StarRocksCluster CR

There are two ways to install Operator and StarRocks Cluster.

1. Install Operator and StarRocks Cluster by yaml Manifest.
2. Install Operator and StarRocks Cluster by Helm Chart.

> Note: In every release, we will provide the latest version of the yaml Manifest and Helm Chart. You can find them
> in https://github.com/StarRocks/starrocks-kubernetes-operator/releases

## Installation by yaml Manifest

Please see [Deploy StarRocks With Operator](./doc/deploy_starrocks_with_operator_howto.md) document for more details.

### 1. Apply the StarRocksCluster CRD

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```

### 2. Apply the Operator manifest

Apply the Operator manifest. By default, the Operator is configured to install in the starrocks namespace. To use the
Operator in a custom namespace, download
the [Operator manifest](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml)
and edit all instances of namespace: starrocks to specify your custom namespace.
Then apply this version of the manifest to the cluster with kubectl apply -f {local-file-path} instead of using the
command below.

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

### 3. Deploy the StarRocks cluster

You need to prepare a separate yaml file to deploy the StarRocks. The starrocks cluster CRD fields explains
in [api.md](./doc/api.md). The [examples](./examples/starrocks) directory contains some simple example for reference.

You can use any of the template yaml file as a starting point. You can further add more configurations into the template
yaml file following this deployment documentation.

For demonstration purpose, we use the [starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml) example
template to start a 3 FE and 3 BE StarRocks cluster.

Here's an example yaml for Docker Desktop with local desktop access with StarRocks 3.2.1 so you can upgrade in later steps.
```
atwong@Albert-CelerData sroperatortest % cat starrocks-fe-and-be.yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.1
    replicas: 3
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 4
      memory: 16Gi
    service:            
      type: LoadBalancer
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.1
    replicas: 3
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 4
      memory: 8Gi
```

```console
kubectl apply -f starrocks-fe-and-be.yaml
```

### 4. Connect the StarRocks cluster

To connect, just use the mysql client and connect to the StarRocks cluster port 9030.  An example of a connection is shown below. 

> [!NOTE]  
>  If you want to connect remotely or through your desktop, you will need to enable the k8s Load Balander.

```
atwong@Albert-CelerData sroperatortest % kubectl -n starrocks get svc
NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                       AGE
starrockscluster-sample-be-search    ClusterIP      None            <none>        9050/TCP                                                      5m2s
starrockscluster-sample-be-service   ClusterIP      10.103.248.52   <none>        9060/TCP,8040/TCP,9050/TCP,8060/TCP                           5m2s
starrockscluster-sample-fe-search    ClusterIP      None            <none>        9030/TCP                                                      6m22s
starrockscluster-sample-fe-service   LoadBalancer   10.99.14.222    localhost     8030:32326/TCP,9020:32578/TCP,9030:30774/TCP,9010:32505/TCP   6m22s
atwong@Albert-CelerData sroperatortest % mysql -h 127.0.0.1 -P 9030 -uroot
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 3
Server version: 5.1.0 3.2.1-79ee91d

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>
```

### 5. Upgrade the StarRocks cluster

To upgrade, just patch the StarRocks cluster. 

```console
kubectl -n starrocks patch starrockscluster starrockscluster-sample --type='merge' -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:latest"}}}'
kubectl -n starrocks patch starrockscluster starrockscluster-sample --type='merge' -p '{"spec":{"starRocksBeSpec":{"image":"starrocks/be-ubuntu:latest"}}}'
```

### 6. Resize the StarRocks cluster

To resize, just patch the StarRocks cluster.

> [!IMPORTANT]
>  Once you deploy with 3 FE nodes, you are in HA mode.  Do not resize FE nodes below 3 since that will affect cluster quorum.  This rule doesn't apply to CN nodes.

```console
kubectl -n starrocks patch starrockscluster starrockscluster-sample --type='merge' -p '{"spec":{"starRocksBeSpec":{"replicas":9}}}'
```

### 7. Expand storage volumes (PVC Expansion)

You can expand storage volumes for FE, BE, and CN components without data loss. The operator supports expanding PVC sizes by patching existing PVCs and recreating StatefulSets when necessary.

> [!IMPORTANT]
> - Storage sizes can only be **increased**, never decreased
> - Your storage class must support volume expansion (`allowVolumeExpansion: true`)
> - The expansion process preserves all data

Example of expanding FE metadata storage from 10Gi to 20Gi:

```console
kubectl -n starrocks patch starrockscluster starrockscluster-sample --type='merge' -p '{"spec":{"starRocksFeSpec":{"storageVolumes":[{"name":"fe-meta","storageSize":"20Gi","mountPath":"/opt/starrocks/fe/meta"}]}}}'
```

Example of expanding BE data storage from 100Gi to 200Gi:

```console
kubectl -n starrocks patch starrockscluster starrockscluster-sample --type='merge' -p '{"spec":{"starRocksBeSpec":{"storageVolumes":[{"name":"be-data","storageSize":"200Gi","mountPath":"/opt/starrocks/be/storage"}]}}}'
```

For detailed information about PVC expansion, see the [PVC Expansion How-To Guide](./doc/pvc_expansion_howto.md).

### 8. Delete/stop the StarRocks cluster

To delete/stop the StarRocks cluster, just execute the delete command.

```console
kubectl delete -f starrocks-fe-and-be.yaml
```
or
```console
kubectl delete starrockscluster starrockscluster-sample -n starrocks
```

### 9. Delete/stop the StarRocks Operator

To delete/stop the StarRocks Operate, just execute the delete command.

```console
kubectl delete -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```


## Installation by Helm Chart

Please see [kube-starrocks](./helm-charts/charts/kube-starrocks/README.md) for how to install both operator and
StarRocks cluster by Helm Chart.

If you want more flexibility in managing your StarRocks clusters, you can deploy Operator
using [operator](./helm-charts/charts/kube-starrocks/charts/operator) Helm Chart and StarRocks
using [starrocks](./helm-charts/charts/kube-starrocks/charts/starrocks) Helm Chart separately.

## Other Documents

- In [doc](./doc) directory, you can find more documents about how to use StarRocks Operator.
- In [examples](./examples/starrocks) directory, you can find more examples about how to write StarRocksCluster CR.
- [Documentation on docs.starrocks.io](https://docs.starrocks.io/docs/deployment/sr_operator/)
