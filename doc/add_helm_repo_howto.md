# Add the Helm Chart Repo for CelerData

1. Add the Helm Chart Repo.

   ```Bash
   helm repo add celerdata https://celerdata.github.io/celerdata-kubernetes-operator
   ```

2. Update the Helm Chart Repo to the latest version.

   ```Bash
   helm repo update celerdata
   ```

3. View the Helm Chart Repo that you added.

   ```Bash
   $ helm search repo celerdata
   NAME                                    CHART VERSION    APP VERSION  DESCRIPTION
   celerdata/kube-celerdata      1.8.0            3.1-latest   kube-celerdata includes two subcharts, celerdata...
   us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator            1.8.0            1.8.0        A Helm chart for CelerData operator
   celerdata/celerdata           1.8.0            3.1-latest   A Helm chart for CelerData cluster
   ```

See [Deploy CelerData With Helm](deploy_celerdata_with_helm_howto.md) to learn how to deploy CelerData with Helm.
