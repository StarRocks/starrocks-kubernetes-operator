# Description

Please provide a detailed description of the changes you have made. Include the motivation for these changes, and any
additional context that may be important.
> the PR should include helm chart changes and operator changes.

# Related Issue(s)

Please list any related issues and link them here.

# Checklist

For operator, please complete the following checklist:

- [ ] run `make generate` to generate the code.
- [ ] run `golangci-lint run` to check the code style.
- [ ] run `make test` to run UT.
- [ ] run `make manifests` to update the yaml files of CRD.

For helm chart, please complete the following checklist:

- [ ] make sure you have updated the [values.yaml](../../helm-charts/charts/kube-starrocks/charts/starrocks/values.yaml)
  file of starrocks chart.
- [ ] In `scripts` directory, run `bash create-parent-chart-values.sh` to update the values.yaml file of the parent
  chart( kube-starrocks chart).
