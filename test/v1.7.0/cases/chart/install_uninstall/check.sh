#! /bin/bash

# if command fails, exit
set -e

# install the helm chart
helm install starrocks -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f values.yaml ../../../../../helm-charts/charts/kube-starrocks/