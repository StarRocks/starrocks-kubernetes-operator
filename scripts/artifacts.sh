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

function helm_package() {
  chart_name=$1
  if [[ $chart_name = "kube-starrocks" ]]; then
    cd $HOME_PATH/helm-charts/charts/kube-starrocks/
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  elif [[ $chart_name = "operator" ]]; then
    cd $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  elif [[ $chart_name = "starrocks" ]]; then
    # helm package for starrocks
    cd $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  fi
}

function get_package_name() {
  chart_name=$1
  if [[ $chart_name = "kube-starrocks" ]]; then
    # get the package name for kube-starrocks from Chart.yaml
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
    echo $package_name
  elif [[ $chart_name = "operator" ]]; then
    # get the package name for kube-starrocks from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator/${package_name} ]; then
      echo "package ${package_name} not found"
      exit 1
    fi
    echo $package_name
  elif [[ $chart_name = "starrocks" ]]; then
    # get the package name for kube-starrocks from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks/${package_name} ]; then
      echo "package ${package_name} not found"
      exit 1
    fi
    echo $package_name
  else
    echo "no such ${chart_name} chart"
    exit 1
  fi
}

function helm_repo_index() {
  chart_name=$1
  release_tag=$RELEASE_TAG
  url=https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}
  if [[ $chart_name = "kube-starrocks" ]]; then
    if [ -f $HOME_PATH/index.yaml ]; then
      helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH/helm-charts/charts/$chart_name
    else
      helm repo index --url $url $HOME_PATH/helm-charts/charts/$chart_name
    fi
    mv $HOME_PATH/helm-charts/charts/kube-starrocks/index.yaml $HOME_PATH/index.yaml
  else
    # for starrocks and operator
    if [ -f $HOME_PATH/index.yaml ]; then
      helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH/helm-charts/charts/kube-starrocks/charts/$chart_name
    else
      helm repo index --url $url $HOME_PATH/helm-charts/charts/kube-starrocks/charts/$chart_name
    fi
    mv $HOME_PATH/helm-charts/charts/kube-starrocks/charts/$chart_name/index.yaml $HOME_PATH/index.yaml
  fi
}

# check parameter
if [ $# -ne 1 ]; then
  echo "Usage: $0 <release_tag>"
  exit 1
fi
export RELEASE_TAG=$1

# include common.sh
source ./common.sh
export HOME_PATH=$(printHomePath)

# use the latest chart values.yaml
bash create-parent-chart-values.sh
# use the latest operator.yaml
bash operator.sh

# artifacts are stored in $HOME_PATH/artifacts
mkdir -p $HOME_PATH/artifacts

echo "mkdir artifacts for kube-starrocks chart"
helm_package kube-starrocks
package_name=$(get_package_name kube-starrocks)
helm_repo_index kube-starrocks
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "mkdir artifacts for operator chart"
helm_package operator
package_name=$(get_package_name operator)
helm_repo_index operator
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-starrocks/charts/operator/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "mkdir artifacts for starrocks chart"
helm_package starrocks
package_name=$(get_package_name starrocks)
helm_repo_index starrocks
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-starrocks/charts/starrocks/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "copy yaml files for operator and crd"
cp $HOME_PATH/deploy/*.yaml $HOME_PATH/artifacts/

echo "generate api.md"
cd $HOME_PATH/scripts
bash gen-api-reference-docs.sh
