#! /bin/bash

# if command fails, exit
set -e

# apply the helm chart
helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f values.yaml ../../../../../helm-charts/charts/kube-starrocks/ | kubectl apply -f -