#! /bin/sh

#
# Copyright 2021-present, StarRocks Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
#  limitations under the License.
#

set -ex

# check parameter
if [ $# -ne 1 ]; then
  echo "Usage: $0 <release_tag>"
  exit 1
fi
release_tag=$1

# set the home path
HOME_PATH=$(
  cd "$(dirname "$0")"
  cd ..
  pwd
)
echo "HOME_PATH: ${HOME_PATH}"

# helm package
cd $HOME_PATH/helm-charts/charts/kube-starrocks/
# must be executed before helm index operation
helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .

# get the package name from Chart.yaml
chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
# get the chart version from Chart.yaml
chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
# make the package name
package_name=${chart_name}-${chart_version}.tgz
# make sure the package exists
if [ ! -f $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name} ]; then
  echo "package ${package_name} not found"
  exit 1
fi

# helm repo index
url=https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}/${chart_name}-${chart_version}.tgz
if [ -f $HOME_PATH/index.yaml ]; then
  helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH/helm-charts/charts/kube-starrocks
else
  helm repo index --url $url $HOME_PATH/helm-charts/charts/kube-starrocks
fi
mv $HOME_PATH/helm-charts/charts/kube-starrocks/index.yaml $HOME_PATH/index.yaml

# the generated index.yaml is not correct, so we need to fix it
# the wrong one, e.g. https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.7.0/kube-starrocks-1.7.0.tgz/artifacts/kube-starrocks-1.7.0.tgz
# first get the url in index.yaml
old=$(cat $HOME_PATH/index.yaml | grep "$url")
new=${old%/*/*}
# then replace the url with the correct one, and do not use sed
sed "s|$old|$new|g" $HOME_PATH/index.yaml >/tmp/index.yaml
cp /tmp/index.yaml $HOME_PATH/index.yaml

# copy to artifacts
mkdir -p $HOME_PATH/artifacts
# helm chart
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

# yaml files for operator and crd
cp $HOME_PATH/deploy/*.yaml $HOME_PATH/artifacts/

# build migrate-chart-value tool
cd $HOME_PATH/scripts/migrate-chart-value
CGO_ENABLED=0 GOOS=linux go build -o migrate-chart-value main.go
cp $HOME_PATH/scripts/migrate-chart-value/migrate-chart-value $HOME_PATH/artifacts/

# gh release upload
# gh release upload $1 $HOME_PATH/artifacts/*.tgz $HOME_PATH/artifacts/*.yaml
