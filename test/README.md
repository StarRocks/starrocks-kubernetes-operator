NOTE: This directory has some data and scripts to make test ease，but not automatically tested.
To run these test cases, you probably need to modify these scripts. In the future, we will integrate these test cases to
github workflow.

we use the following tools for test:

1. use [kind](https://kind.sigs.k8s.io/) to create a kubernetes cluster.
2. use [kustomize](https://kustomize.io/) to create the resources of each case.
3. use [helm](https://helm.sh/) to test our helm chart.

# install.sh

This script will do the following things:

1. install a kubernetes cluster using kind.
2. deploy the crd and starrocks operator to the cluster by version.
3. deploy the StarRocksCluster resource to the cluster defined in './cluster'.

run the script:

```bash
bash install.sh v1.7.0
```

# diff.sh

This is used to compare the resources of starrocks cluster created by the operator.

NOTE: you must run `install.sh` first, and collect the resources of the cluster.

e.g.

```bash
bash diff.sh starrockscluster-sample v1.6.1 v1.7.0  >/tmp/diff.data
```

# cases

For every release, we will add some cases to test the new features.

In `v1.7.0/cases`

1. test cases for operator。we use `kustomize` to create the resources of each case, and use
   the `kubectl kustomize . | kubectl apply -f-` to deploy the resources. It makes test ease, but not automatically
   tested.
    1. `mount_configmaps`: mount configmaps to starrocks cluster.
    2. `mount_secrets`: mount secrets to starrocks cluster.
    3. `support_fileds`: add the new added support fields of starrocks cluster.
2. test cases for helm chart。We use customized values.yaml to render the helm chart, and check the correctness.
    1. `modify_start_configmap`: modify the start configmap of starrocks cluster, and expect the related pod will be
       restarted.
    2. `operator_support_mode_fields`: add annotations, nodeSelector, tolerations to operator. use zsh to
       run `check.sh`.
    3. `init_password`: init password for starrocks cluster.
    4. `install_uninstall`: install and uninstall the starrocks cluster by helm.
    5. `set_cluster_by_chart`: set the starrocks cluster by modify the values.yaml, not modify the StarRocksCluster
       resource directly.