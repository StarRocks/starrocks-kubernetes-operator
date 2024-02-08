# Deploy Multiple Clusters HOWTO

If you have deployed a StarRocks cluster by YAML manifests, you can write a new StarRocksCluster CR YAML to deploy
another StarRocks cluster.

We have split the `kube-starrocks` chart into two subcharts: `operator` and `starrocks`. Installing `kube-starrocks` is
equivalent to installing both `operator` and `starrocks` subcharts, and uninstalling `kube-starrocks` is equivalent to
uninstalling both `operator` and `starrocks` subcharts.

If you have deployed a StarRocks cluster by `operator` + `starrocks` helm chart, you can deploy another StarRocks
cluster by the `starrocks` helm chart.

If you have deployed a StarRocks cluster by `kube-starrocks` helm chart, you have two ways to deploy another StarRocks
cluster.

1. Deploy another StarRocks cluster by `starrocks` helm chart.
2. Deploy another StarRocks cluster by `kube-starrocks` Helm chart.

This document will guide you through the process of deploying multiple StarRocks clusters by `kube-starrocks` helm
chart.

## Deploy another StarRocks cluster by `kube-starrocks` Helm chart

By default, the operator will watch all namespaces. If you want to deploy another StarRocks cluster
by `kube-starrocks`, you should limit `each operator` to watch a specific namespace.

```yaml
operator:
  starrocksOperator:
    watchNamespace: "your-namespace"
```

> you can also add `--set operator.starrocksOperator.watchNamespace="your-namespace"` to the `helm` command which has
> higher priority.

So, the steps to deploy multiple StarRocks clusters by `kube-starrocks` are:

1. update `values.yaml` file of the first deployed StarRocks cluster to limit the operator to watch a specific
   namespace.
2. upgrade the first StarRocks cluster.
3. install the second StarRocks cluster by the same `kube-starrocks` chart, and do not forget to specify the namespace.
