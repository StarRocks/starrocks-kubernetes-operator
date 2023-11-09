# Add the Helm Chart Repo for StarRocks

1. Add the Helm Chart Repo.

   ```Bash
   helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
   ```

2. Update the Helm Chart Repo to the latest version.

   ```Bash
   helm repo update starrocks-community
   ```

3. View the Helm Chart Repo that you added.

   ```Bash
   $ helm search repo starrocks-community
   NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
   starrocks-community/kube-starrocks      1.8.0            3.1-latest   kube-starrocks includes two subcharts, starrock...
   starrocks-community/operator            1.8.0            1.8.0        A Helm chart for StarRocks operator
   starrocks-community/starrocks           1.8.0            3.1-latest   A Helm chart for StarRocks cluster
   ```

See [Deploy StarRocks With Helm](deploy_starrocks_with_helm_howto.md) to learn how to deploy StarRocks with Helm.
