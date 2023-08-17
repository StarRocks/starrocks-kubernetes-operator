# StarRocks Kubernetes Helm Charts

There are three charts in this repository:

```bash
$ helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
$ helm repo update
$ helm search repo starrocks-community
NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
starrocks-community/kube-starrocks      1.8.0            3.1-latest   kube-starrocks includes two subcharts, starrock...
starrocks-community/operator            1.8.0            1.8.0        A Helm chart for StarRocks operator
starrocks-community/starrocks           1.8.0            3.1-latest   A Helm chart for StarRocks cluster
```

1. `kube-starrocks` includes two subcharts, `operator` and `starrocks`. See [kube-starrocks](./charts/kube-starrocks/README.md) for more details.
2. `operator` is the Helm chart for StarRocks operator. See [operator](./charts/kube-starrocks/charts/operator/README.md) for more details.
3. `starrocks` is the Helm chart for StarRocks cluster. See [starrocks](./charts/kube-starrocks/charts/starrocks/README.md) for more details.
