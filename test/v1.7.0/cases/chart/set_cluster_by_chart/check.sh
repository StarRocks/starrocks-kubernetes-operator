#! /bin/zsh

# check the rendered manifests, make sure annotations, nodeSelector, tolerations are rendered correctly.
# note: you should use zsh to run this script.

# if command fails, exit
set -e

# make manifests
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f values.yaml ../../../../../helm-charts/charts/kube-starrocks/)
echo $manifests | kubectl apply -f - --dry-run='none'

cluster=$(echo "${manifests}" | awk '/^---$/{i++}i==11' | sed 's/^---$//g')
echo ${cluster}
