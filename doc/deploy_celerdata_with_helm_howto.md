# Deploy CelerData With Helm

We have split the kube-celerdata chart into two subcharts: operator and CelerData since v1.8.0.

Installing kube-celerdata is equivalent to installing both operator and `celerdata` subcharts, and uninstalling
kube-celerdata is equivalent to uninstalling both operator and `celerdata` subcharts.

If you want more flexibility in managing your CelerData clusters, you can install operator and `celerdata` subcharts
separately.

There are three charts in this repository:

```bash
$ helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
$ helm repo update celerdata
$ helm search repo celerdata
NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
# install both operator and celerdata
celerdata/kube-celerdata      1.8.0            3.1-latest   kube-celerdata includes two subcharts, celerdata...
# install operator only
us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator            1.8.0            1.8.0        A Helm chart for CelerData operator
# install celerdata only
celerdata/celerdata           1.8.0            3.1-latest   A Helm chart for CelerData cluster
```

1. `kube-celerdata` includes two subcharts, `operator` and `celerdata`.
   See [kube-celerdata](../helm-charts/charts/kube-celerdata/README.md) for more details.
2. `operator` is the Helm chart for CelerData operator.
   See [operator](../helm-charts/charts/kube-celerdata/charts/operator/README.md) for more details.
3. `celerdata` is the Helm chart for CelerData cluster.
   See [starrocks](../helm-charts/charts/kube-celerdata/charts/celerdata/README.md) for more details.
