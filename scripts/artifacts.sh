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

# parse the package name
# get the name from the path, e.g. kube-starrocks-1.6.1.tgz
package_name=$(ls -t $HOME_PATH/helm-charts/charts/kube-starrocks/*.tgz | head -1 | awk -F '/' '{print $NF}')
# remove the suffix from package_name, e.g. kube-starrocks-1.6.1
package_name=${package_name%.tgz}
# remove the version from package_name, e.g. kube-starrocks
name=${package_name%-*}
# get the version from package_name, e.g. 1.6.1
version=${package_name##*-}

# helm repo index
url=https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/${release_tag}/${name}-chart-${version}.tgz
if [ -f $HOME_PATH/helm-charts/charts/index.yaml ]; then
  helm repo index --merge $HOME_PATH/helm-charts/charts/index.yaml --url $url ..
else
  helm repo index --url $url ..
fi

# copy to artifacts
mkdir -p $HOME_PATH/artifacts
# helm chart
mv $HOME_PATH/helm-charts/charts/kube-starrocks/${package_name}.tgz $HOME_PATH/artifacts/${name}-chart-${version}.tgz
# yaml files for operator and crd
cp $HOME_PATH/deploy/*.yaml $HOME_PATH/artifacts/

# gh release upload
# gh release upload $1 $HOME_PATH/artifacts/*.tgz $HOME_PATH/artifacts/*.yaml
