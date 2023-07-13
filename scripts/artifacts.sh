#! /bin/sh

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
helm package .

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
url=https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}/${chart_name}-chart-${chart_version}.tgz
if [ -f $HOME_PATH/index.yaml ]; then
  helm repo index --merge $HOME_PATH/index.yaml --url $url $HOME_PATH
else
  helm repo index --url $url $HOME_PATH
fi

# the generated index.yaml is not correct, so we need to fix it
# the wrong one, e.g. https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.7.0/kube-starrocks-chart-1.7.0.tgz/artifacts/kube-starrocks-1.7.0.tgz
# first get the url in index.yaml
old=$(cat $HOME_PATH/index.yaml | grep "$url")
new=${old%/*/*}
# then replace the url with the correct one, and do not use sed
sed "s|$old|$new|g" $HOME_PATH/index.yaml >/tmp/index.yaml
cp /tmp/index.yaml $HOME_PATH/index.yaml

# copy to artifacts
mkdir -p $HOME_PATH/artifacts
# helm chart
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name} $HOME_PATH/artifacts/${chart_name}-chart-${chart_version}.tgz
# yaml files for operator and crd
cp $HOME_PATH/deploy/*.yaml $HOME_PATH/artifacts/

# gh release upload
# gh release upload $1 $HOME_PATH/artifacts/*.tgz $HOME_PATH/artifacts/*.yaml
