# Deploy Warehouse

From StarRocks Operator v1.9.0, StarRocksWarehouse CRD is introduced to manage the warehouse. This document describes
how to deploy a warehouse.

> Note: Warehouse is an enterprise feature for StarRocks.

## 1. Prerequisites

1. StarRocks Operator >= v1.9.0. The latest version of the operator is recommended.
2. An installed StarRocks Cluster.
   See [deploy_starrocks_with_operator_howto.md](./deploy_starrocks_with_operator_howto.md)
   or [deploy_starrocks_with_helm_howto.md](./deploy_starrocks_with_helm_howto.md) for more details.
3. Starrocks enterprise version >= v3.2.0.

## 2. Deploy Warehouse

You can choose one of the following methods to deploy a warehouse:

1. Deploy Warehouse by YAML Manifest
2. Deploy Warehouse by Helm Chart

> If you deploy StarRocks Cluster with BE and CN nodes, they are added to the `default_warehouse` in StarRocks by
> default. You can also define a Warehouse CR named `default-warehouse` to add more BE and CN nodes to the warehouse.

### 2.1 Deploy Warehouse by YAML Manifest

First, we need to install StarRocksWarehouse CRD and restart the StarRocks operator to make it aware of the new CRD.

```console
# install crd
kubectl apply -f https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.9.6/starrocks.com_starrockswarehouses.yaml

# restart operator
kubectl rollout restart deployment kube-starrocks-operator
```

Then, we need to deploy a warehouse by the following YAML manifest.

```yaml
# wh1.yaml
apiVersion: starrocks.com/v1
kind: StarRocksWarehouse
metadata:
  # A warehouse will be created with this name in StarRocks Cluster. If you are using dash(-) in the name, the warehouse
  # name created by StarRocks will be replaced with underscore(_).
  name: wh1

spec:
  # Make sure the StarRocks cluster exists in the same namespace.
  # You can check it by running `kubectl get starrocksclusters.starrocks.com`.
  starRocksCluster: kube-starrocks
  template:
    envVars:
      - name: TZ
        value: Asia/Shanghai
    image: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/cn-ubuntu:3.2.6-ee
    replicas: 1
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 8
      memory: 8Gi
```

You can see [api.md](./api.md) for more details about the StarRocksWarehouse CRD fields. The spec part is very similar
to the StarRocksCnSpec of StarRocksCluster, so
see [deploy_a_starrocks_cluster_with_cn.yaml](../examples/starrocks/deploy_a_starrocks_cluster_with_cn.yaml) for more
fields.

Apply the YAML manifest:

```bash
kubectl -n starrocks apply -f wh1.yaml
```

### 2.2. Deploy Warehouse by Helm Chart

We also support deploying a warehouse by Helm chart.
You can also see [Warehouse Chart](../helm-charts/charts/warehouse/README.md) for how to deploy it.

First, prepare a values.yaml file for Warehouse chart.

```yaml
# wh1-values.yaml
spec:
  # Make sure the StarRocks cluster exists in the same namespace.
  # You can check it by running `kubectl get starrocksclusters.starrocks.com`.
  starRocksClusterName: kube-starrocks
  replicas: 1
  image:
    repository: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/cn-ubuntu
    tag: "3.2.6-ee"
  resources:
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 8
      memory: 8Gi
```

Then deploy a warehouse by the following command:

```console
# Use the above values.yaml to deploy a warehouse in namespace starrocks
helm -n starrocks install wh1 starrocks-community/warehouse -f wh1-values.yaml

# Restart the StarRocks operator to make it aware of the new CRD
kubectl -n starrocks rollout restart deployment kube-starrocks-operator
```

## 3. Manage Warehouse

### 3.1. Show the deployed warehouse

If you have deployed the above warehouse, you can see it by using the following SQL command:

```console
# A warehouse has been created with the name `wh1`.
mysql> show warehouses;
+-------+-------------------+-----------+-----------+---------------------+-----------------+-----------------+------------+-----------+-----------+---------------------+---------------------+----------------------------------------------+
| Id    | Name              | State     | NodeCount | CurrentClusterCount | MaxClusterCount | StartedClusters | RunningSql | QueuedSql | CreatedOn | ResumedOn           | UpdatedOn           | Comment                                      |
+-------+-------------------+-----------+-----------+---------------------+-----------------+-----------------+------------+-----------+-----------+---------------------+---------------------+----------------------------------------------+
| 0     | default_warehouse | AVAILABLE | 0         | 1                   | 1               | 1               | 0          | 0         | NULL      | 2024-05-11 16:49:37 | 2024-05-11 17:53:30 | An internal warehouse init after FE is ready |
| 35030 | wh1               | AVAILABLE | 1         | 1                   | 1               | 1               | 0          | 0         | NULL      | NULL                | NULL                | NULL                                         |
+-------+-------------------+-----------+-----------+---------------------+-----------------+-----------------+------------+-----------+-----------+---------------------+---------------------+----------------------------------------------+
2 rows in set (0.00 sec)
```

### 3.2 Upgrade Deployment

We strongly recommend you to upgrade deployment by modifying the YAML Manifest file or values.yaml file. For example,
you can update any fields in the file, e.g. the image version, replicas, and resources.

> We don't suggest you to modify the deployment of warehouse by `kubectl edit`.

#### 3.2.1 Update the YAML manifest

For example, upgrade the image version:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksWarehouse
metadata:
  # A warehouse will be created with this name in StarRocks Cluster. If you are using dash(-) in the name, the warehouse
  # name created by StarRocks will be replaced with underscore(_).
  name: wh1

spec:
  # Make sure the StarRocks cluster exists in the same namespace.
  # You can check it by running `kubectl get starrocksclusters.starrocks.com`.
  starRocksCluster: kube-starrocks
  template:
    envVars:
      - name: TZ
        value: Asia/Shanghai
    image: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/cn-ubuntu:3.2.7-ee  # this line is updated
    replicas: 1
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 8
      memory: 8Gi
```

Apply the updated YAML manifest:

```console
kubectl -n starrocks apply -f wh1.yaml
```

### 3.2.2 Update values.yaml for Helm chart

For example, upgrade the image version:

```yaml
# wh1-values.yaml
spec:
  # Make sure the StarRocks cluster exists in the same namespace.
  # You can check it by running `kubectl get starrocksclusters.starrocks.com`.
  starRocksClusterName: kube-starrocks
  replicas: 1
  image:
    repository: us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/cn-ubuntu
    tag: "3.2.7-ee" # this line is updated
  resources:
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 8
      memory: 8Gi
```

Then upgrade the warehouse by the following command:

```console
helm -n starrocks upgrade wh1 starrocks-community/warehouse -f wh1-values.yaml
```

## 4. Delete the Warehouse

If you deployed the warehouse by YAML manifest, you can delete it by running the following command:

```console
kubectl delete -f wh1.yaml
```

If you deployed the warehouse by Helm chart, you can delete it by running the following command:

```console
helm -n starrocks uninstall wh1
```
