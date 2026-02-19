# Deploy Multiple Clusters HOWTO

If you have deployed a CelerData cluster by YAML manifests, you can write a new CelerDataCluster CR YAML to deploy
another CelerData cluster.

We have split the `kube-celerdata` chart into two subcharts: `operator` and `celerdata`. Installing `kube-celerdata` is
equivalent to installing both `operator` and `celerdata` subcharts, and uninstalling `kube-celerdata` is equivalent to
uninstalling both `operator` and `celerdata` subcharts.

If you have deployed a CelerData cluster by `operator` + `celerdata` helm chart, you can deploy another CelerData
cluster by the `celerdata` helm chart.

If you have deployed a CelerData cluster by `kube-celerdata` helm chart, you have two ways to deploy another CelerData
cluster.

1. Deploy another CelerData cluster by `celerdata` helm chart.
2. Deploy another CelerData cluster by `kube-celerdata` Helm chart.

This document will guide you through the process of deploying multiple CelerData clusters by `kube-celerdata` helm
chart.

## Deploy another CelerData cluster by `kube-celerdata` Helm chart

By default, the operator will watch all namespaces. If you want to deploy another CelerData cluster
by `kube-celerdata`, you should limit `each operator` to watch a specific namespace.

```yaml
operator:
  celerDataOperator:
    watchNamespace: "your-namespace"
```

> you can also add `--set operator.celerDataOperator.watchNamespace="your-namespace"` to the `helm` command which has
> higher priority.

So, the steps to deploy multiple CelerData clusters by `kube-celerdata` are:

1. update `values.yaml` file of the first deployed CelerData cluster to limit the operator to watch a specific
   namespace.
2. upgrade the first CelerData cluster.
3. install the second CelerData cluster by the same `kube-celerdata` chart, and do not forget to specify the namespace.
