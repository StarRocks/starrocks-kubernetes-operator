This directory contains scripts to help with the development of the StarRocks Operator.

`artifacts.sh` is used to generate the artifacts for the StarRocks Operator when a new release is made.
It includes:
```console
bash artifacts.sh <version>

# output
# for example
kube-starrocks-1.7.1.tgz
kube-starrocks-1.7.1.tgz.prov
operator.yaml
starrocks.com_starrocksclusters.yaml
```

`create-parent-chart-values.sh` is used to generate the parent chart values. It just contains the values of the child charts.
```console
# the output file will be saved to `helm-charts/charts/kube-starrocks/values.yaml`.
bash create-parent-chart-values.sh
```

`operator.sh` is used to generate the operator.yaml file. It will be saved to `deploy/operator.yaml`.
```console
# the output file will be saved to `deploy/operator.yaml`.
bash operator.sh
```