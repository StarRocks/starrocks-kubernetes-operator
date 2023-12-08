#! /bin/bash

# It use gen-crd-api-reference-docs to provide API reference docs for Custom Resource Definitions: StarRockCluster and StarRocksWarehouse

# include common.sh
source ./common.sh
export HOME_PATH=$(printHomePath)

# Download the gen-crd-api-reference-docs tool
curl -sLO https://github.com/ahmetb/gen-crd-api-reference-docs/archive/refs/heads/master.zip

# Unzip the gen-crd-api-reference-docs tool
unzip master.zip

# build the gen-crd-api-reference-docs tool
cd gen-crd-api-reference-docs-master
go build .

# Generate the API reference docs for StarRockCluster and StarRocksWarehouse
cd $HOME_PATH
./scripts/gen-crd-api-reference-docs-master/gen-crd-api-reference-docs \
  -config=./scripts/gen-crd-api-reference-docs-master/example-config.json \
  -template-dir=./scripts/gen-crd-api-reference-docs-master/template \
  -api-dir=github.com/StarRocks/starrocks-kubernetes-operator/pkg/apis/starrocks/v1 \
  -out-file=./doc/api.md

# remove the gen-crd-api-reference-docs tool
cd $HOME_PATH/scripts
rm -rf master.zip
rm -rf gen-crd-api-reference-docs-master
