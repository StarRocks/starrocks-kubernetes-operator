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
  # You can check it by running `kubectl -n starrocks get starrocksclusters.starrocks.com`.
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
  # You can check it by running `kubectl -n starrocks get starrocksclusters.starrocks.com`.
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
helm -n starrocks install wh1 starrocks/warehouse -f wh1-values.yaml

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
  # You can check it by running `kubectl -n starrocks get starrocksclusters.starrocks.com`.
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
  # You can check it by running `kubectl -n starrocks get starrocksclusters.starrocks.com`.
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
helm -n starrocks upgrade wh1 starrocks/warehouse -f wh1-values.yaml
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

Deleting a warehouse CR normally also removes the warehouse from the cluster: before it lets the
warehouse's CN StatefulSet go, the operator runs `DROP WAREHOUSE` in the FE. Everything set on the
warehouse inside the database (its properties and the privileges granted on it) is dropped with it.

This is best effort. If the FE is unreachable or rejects the statement, the operator logs the error and
lets the StatefulSet go anyway, so the warehouse stays in the cluster. After deleting a warehouse CR,
run `SHOW WAREHOUSES` to confirm it is gone, and drop it with `DROP WAREHOUSE` if it is still listed.

### 4.1 Delete the CR but keep the warehouse in the cluster

Sometimes the CR is deleted only to be applied again under the same name, for example when moving the
warehouse between operators or rolling such a move back. To keep the warehouse, its properties and its
grants in the cluster across that, annotate the CR first:

```console
kubectl -n starrocks annotate starrockswarehouse wh1 starrocks.com/skip-fe-cleanup=true
```

Or put the annotation in the manifest:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksWarehouse
metadata:
  name: wh1
  namespace: starrocks
  annotations:
    starrocks.com/skip-fe-cleanup: "true"
```

For a warehouse deployed by Helm chart, use the `kubectl annotate` command above; `helm upgrade` leaves
the annotation in place.

The annotation takes effect once the operator has copied it onto the warehouse's CN StatefulSet and
removed its protection finalizer there. Check it before you delete the CR:

```console
kubectl -n starrocks get sts wh1-warehouse-cn \
  -o jsonpath='{.metadata.finalizers}{"\n"}{.metadata.annotations.starrocks\.com/skip-fe-cleanup}{"\n"}'
```

The first line must not list `starrocks.com.starrockswarehouse/protection`, and the second line must
print `true`. Adding the annotation changes only the StatefulSet's metadata, so no CN pod restarts.

From then on, deleting the CR is an ordinary Kubernetes delete. The operator sends nothing to the FE,
and the CN StatefulSet is removed even if the operator is not running at that moment. In the cluster,
the warehouse stays as it was and its compute nodes show `Alive = false` in `SHOW COMPUTE NODES` until a
same-named warehouse brings them back. If the new CR has fewer replicas than the old one, the extra
compute nodes stay offline; drop them with `ALTER SYSTEM DROP COMPUTE NODE`.

To delete the warehouse for good later, remove the annotation first:

```console
kubectl -n starrocks annotate starrockswarehouse wh1 starrocks.com/skip-fe-cleanup-
```

Then wait until the operator has put the protection finalizer back, and only then delete the CR. Run the
same check as above: the first line must list `starrocks.com.starrockswarehouse/protection`, and the
second line must be empty. If you delete the CR before that, the StatefulSet still carries the opt-out,
so the operator skips `DROP WAREHOUSE` and the warehouse stays in the cluster.

If you delete an annotated CR and no longer need the warehouse, drop it yourself with `DROP WAREHOUSE`.
