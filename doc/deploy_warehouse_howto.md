# Deploy Warehouse

From StarRocks Operator v1.9.0, StarRocksWarehouse CRD is introduced to manage the warehouse. This document describes
how to deploy a warehouse.

> Note: Warehouse is an enterprise feature for StarRocks.

## Prerequisites

1. StarRocks Operator >= v1.9.0. Using the latest version of the operator is recommended.
2. An installed StarRocks Cluster.
   See [deploy_starrocks_with_operator_howto.md](./deploy_starrocks_with_operator_howto.md)
   or [deploy_starrocks_with_helm_howto.md](./deploy_starrocks_with_helm_howto.md) for more details.
3. Starrocks enterprise version >= v3.2.0.

You can choose one of the following methods to deploy a warehouse:

1. Deploy Warehouse by YAML Manifest
2. Deploy Warehouse by Helm Chart

## Deploy Warehouse by YAML Manifest

First, we should install StarRocksWarehouse CRD.

```console
kubectl apply -f https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.9.6/starrocks.com_starrockswarehouses.yaml
```

Then, we can deploy a warehouse by the following YAML manifest.

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksWarehouse
metadata:
  # A warehouse will be created with this name in StarRocks Cluster. If you are using dash(-) in the name, the warehouse
  # name created by StarRocks will be replaced with underscore(_).
  name: wh1
  namespace: starrocks
spec: # Make sure the StarRocks cluster exists in namespace, e.g. starrocks.
  # You can check it by running `kubectl get starrocksclusters.starrocks.com`.
  starRocksCluster: kube-starrocks
  template:
    envVars:
      - name: TZ
        value: Asia/Shanghai
    image: your-enterprise-image-version-for-cn
    limits:
      cpu: 8
      memory: 8Gi
    requests:
      cpu: 8
      memory: 8Gi
```

You can see [./api.md](api.md) for more details about the StarRocksWarehouse CRD fields. The spec part is very similar
to the StarRocksCnSpec of StarRocksCluster.

## Deploy Warehouse by Helm Chart

`warehouse` is the Helm chart for StarRocks warehouse.
See [Warehouse Chart](../helm-charts/charts/kube-starrocks/charts/starrocks/README.md) for how to deploy it.
