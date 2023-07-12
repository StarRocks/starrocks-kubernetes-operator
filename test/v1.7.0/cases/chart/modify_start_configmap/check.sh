#! /bin/bash

# when `.Values.starrocksFESpec.config`, `.Values.starrocksCnSpec.config`, `.Values.starrocksBeSpec.config` are changed,
# we expect the annotation `app.starrocks.io/fe-config-hash`, `app.starrocks.io/be-config-hash`, `app.starrocks.io/cn-config-hash` are changed,
# which belong the pods created by related statefulset.

# if command fails, exit
set -e

# get the current value of annotation
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f values.yaml ../../../../../helm-charts/charts/kube-starrocks/)
fe_config_hash=$(echo "${manifests}" | grep "app.starrocks.io/fe-config-hash" | awk '{print $2}')
be_config_hash=$(echo "${manifests}" | grep "app.starrocks.io/be-config-hash" | awk '{print $2}')
cn_config_hash=$(echo "${manifests}" | grep "app.starrocks.io/cn-config-hash" | awk '{print $2}')
echo "fe_config_hash: ${fe_config_hash}"
echo "be_config_hash: ${be_config_hash}"
echo "cn_config_hash: ${cn_config_hash}"

# modify the config of starrocksFESpec
sed "s/query_port = 9030/query_port = 19030/g" values.yaml >/tmp/values.yaml
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f /tmp/values.yaml ../../../../../helm-charts/charts/kube-starrocks/)
fe_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/fe-config-hash" | awk '{print $2}')
be_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/be-config-hash" | awk '{print $2}')
cn_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/cn-config-hash" | awk '{print $2}')
if [ "${fe_config_hash}" = "${fe_config_hash2}" ]; then
  echo "fe_config_hash is not changed!!!"
  exit 1
fi
if [ "${be_config_hash}" != "${be_config_hash2}" ]; then
  echo "be_config_hash is changed!!!"
  exit 1
fi
if [ "${cn_config_hash}" != "${cn_config_hash2}" ]; then
  echo "cn_config_hash is changed!!!"
  exit 1
fi

# modify the config of starrocksBeSpec
sed "s/be_port = 9060/be_port = 19060/g" values.yaml >/tmp/values.yaml
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f /tmp/values.yaml ../../../../../helm-charts/charts/kube-starrocks/)
fe_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/fe-config-hash" | awk '{print $2}')
be_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/be-config-hash" | awk '{print $2}')
cn_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/cn-config-hash" | awk '{print $2}')
if [ "${fe_config_hash}" != "${fe_config_hash2}" ]; then
  echo "fe_config_hash is not changed!!!"
  exit 1
fi
if [ "${be_config_hash}" = "${be_config_hash2}" ]; then
  echo "be_config_hash is changed!!!"
  exit 1
fi
if [ "${cn_config_hash}" != "${cn_config_hash2}" ]; then
  echo "cn_config_hash is changed!!!"
  exit 1
fi

# modify the config of starrocksCnSpec
sed "s/thrift_port = 9060/thrift_port = 19060/g" values.yaml >/tmp/values.yaml
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f /tmp/values.yaml ../../../../../helm-charts/charts/kube-starrocks/)
fe_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/fe-config-hash" | awk '{print $2}')
be_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/be-config-hash" | awk '{print $2}')
cn_config_hash2=$(echo "${manifests}" | grep "app.starrocks.io/cn-config-hash" | awk '{print $2}')
if [ "${fe_config_hash}" != "${fe_config_hash2}" ]; then
  echo "fe_config_hash is not changed!!!"
  exit 1
fi
if [ "${be_config_hash}" != "${be_config_hash2}" ]; then
  echo "be_config_hash is changed!!!"
  exit 1
fi
if [ "${cn_config_hash}" = "${cn_config_hash2}" ]; then
  echo "cn_config_hash is changed!!!"
  exit 1
fi
