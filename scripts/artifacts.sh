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
  if [[ $chart_name = "kube-celerdata" ]]; then
    cd $HOME_PATH/helm-charts/charts/kube-celerdata/
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  elif [[ $chart_name = "operator" ]]; then
    cd $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  elif [[ $chart_name = "celerdata" ]]; then
    # helm package for celerdata
    cd $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  elif [[ $chart_name = "warehouse" ]]; then
    # helm package for warehouse
    cd $HOME_PATH/helm-charts/charts/warehouse
    # must be executed before helm index operation
    helm package --sign --key 'yandongxiao' --keyring ~/.gnupg/secring.gpg .
  fi
}

function get_package_name() {
  chart_name=$1
  if [[ $chart_name = "kube-celerdata" ]]; then
    # get the package name for kube-celerdata from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/kube-celerdata/${package_name} ]; then
      echo "package ${package_name} not found"
      exit 1
    fi
    echo $package_name
  elif [[ $chart_name = "operator" ]]; then
    # get the package name for operator from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator/${package_name} ]; then
      echo "package ${package_name} not found"
      exit 1
    fi
    echo $package_name
  elif [[ $chart_name = "celerdata" ]]; then
    # get the package name for celerdata from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata/${package_name} ]; then
      echo "package ${package_name} not found"
      exit 1
    fi
    echo $package_name
  elif [[ $chart_name = "warehouse" ]]; then
    # get the package name for warehouse from Chart.yaml
    chart_name=$(cat $HOME_PATH/helm-charts/charts/warehouse/Chart.yaml | grep '^name: ' | awk -F ': ' '{print $NF}')
    # get the chart version from Chart.yaml
    chart_version=$(cat $HOME_PATH/helm-charts/charts/warehouse/Chart.yaml | grep '^version: ' | awk -F ': ' '{print $NF}')
    # make the package name
    package_name=${chart_name}-${chart_version}.tgz
    # make sure the package exists
    if [ ! -f $HOME_PATH/helm-charts/charts/warehouse/${package_name} ]; then
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
  url=https://github.com/celerdata/celerdata-kubernetes-operator/releases/download/${release_tag}
  if [[ $chart_name = "kube-celerdata" || $chart_name = "warehouse" ]]; then
    if [ -f $HOME_PATH/index.yaml ]; then
      helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH/helm-charts/charts/$chart_name
    else
      helm repo index --url $url $HOME_PATH/helm-charts/charts/$chart_name
    fi
    mv $HOME_PATH/helm-charts/charts/$chart_name/index.yaml $HOME_PATH/index.yaml
  else
    # for celerdata and operator
    if [ -f $HOME_PATH/index.yaml ]; then
      helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH/helm-charts/charts/kube-celerdata/charts/$chart_name
    else
      helm repo index --url $url $HOME_PATH/helm-charts/charts/kube-celerdata/charts/$chart_name
    fi
    mv $HOME_PATH/helm-charts/charts/kube-celerdata/charts/$chart_name/index.yaml $HOME_PATH/index.yaml
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

echo "mkdir artifacts for kube-celerdata chart"
helm_package kube-celerdata
package_name=$(get_package_name kube-celerdata)
helm_repo_index kube-celerdata
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-celerdata/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-celerdata/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "mkdir artifacts for operator chart"
helm_package operator
package_name=$(get_package_name operator)
helm_repo_index operator
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-celerdata/charts/operator/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "mkdir artifacts for celerdata chart"
helm_package celerdata
package_name=$(get_package_name celerdata)
helm_repo_index celerdata
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/kube-celerdata/charts/celerdata/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "mkdir artifacts for warehouse chart"
helm_package warehouse
package_name=$(get_package_name warehouse)
helm_repo_index warehouse
# move the package to artifacts
mv $HOME_PATH/helm-charts/charts/warehouse/${package_name} $HOME_PATH/artifacts/${package_name}
mv $HOME_PATH/helm-charts/charts/warehouse/${package_name}.prov $HOME_PATH/artifacts/${package_name}.prov

echo "copy yaml files for operator and crd"
cp $HOME_PATH/deploy/*.yaml $HOME_PATH/artifacts/

echo "generate api.md"
cd $HOME_PATH/scripts
bash gen-api-reference-docs.sh
