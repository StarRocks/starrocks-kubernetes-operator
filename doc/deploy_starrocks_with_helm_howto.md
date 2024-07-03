# Deploy StarRocks With Helm

We have split the kube-starrocks chart into two subcharts: operator and starrocks since v1.8.0.

Installing kube-starrocks is equivalent to installing both operator and starrocks subcharts, and uninstalling
kube-starrocks is equivalent to uninstalling both operator and starrocks subcharts.

If you want more flexibility in managing your StarRocks clusters, you can install operator and starrocks subcharts
separately.

There are three charts in this repository:

```bash
$ helm repo add starrocks https://starrocks.github.io/starrocks-kubernetes-operator
$ helm repo update starrocks
$ helm search repo starrocks
NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
# install both operator and starrocks
starrocks/kube-starrocks      1.8.0            3.1-latest   kube-starrocks includes two subcharts, starrock...
# install operator only
starrocks/operator            1.8.0            1.8.0        A Helm chart for StarRocks operator
# install starrocks only
starrocks/starrocks           1.8.0            3.1-latest   A Helm chart for StarRocks cluster
```

1. `kube-starrocks` includes two subcharts, `operator` and `starrocks`.
   See [kube-starrocks](../helm-charts/charts/kube-starrocks/README.md) for more details.
2. `operator` is the Helm chart for StarRocks operator.
   See [operator](../helm-charts/charts/kube-starrocks/charts/operator/README.md) for more details.
3. `starrocks` is the Helm chart for StarRocks cluster.
   See [starrocks](../helm-charts/charts/kube-starrocks/charts/starrocks/README.md) for more details.
