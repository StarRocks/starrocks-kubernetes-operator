# Build your own container image HOWTO

This document describes how to build your own container image for the StarRocks Operator. There are some reasons why you
might want to build your own container image, for example:

1. You want to add fields to the StarRocksCluster CRD.
2. Your k8s version is too old to support the latest StarRocks Operator image.

This document will guide you through the process of building your own container image for the StarRocks Operator.

## Prerequisites

1. Automated build tool Make installed.
2. Git installed.
3. Go installed.
4. [Docker](https://docs.docker.com/get-docker/) installed.
5. A container image repository, such as [Docker Hub](https://hub.docker.com/).

In order to contextualize this process, we assume that your k8s does not support appProtocol, and you want to remove it
from the Operator.

## 1. Find the commit ID you want to revert

First, you need to find the PR that introduced the feature you want to remove. For example, we
find [this PR](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/288) by searching for `appProtocol`, and
its PR ID is 288.

Second, in the PR page, you can find the commit ID in the URL. For example, the commit ID
of [this PR](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/288) is `66e2f2c`.

## 2. Revert the commit

```bash
# clone the repository
git clone https://github.com/StarRocks/starrocks-kubernetes-operator.git

# checkout the latest released verion
git checkout v1.9.7

# revert the commit
git revert 66e2f2c

# During the Revert process, there may be conflicts. For conflicts in `_test.go`, you can directly delete the file.
git rm pkg/common/resource_utils/service_test.go
git revert --continue
```

## 3. Use Make to update CRD

```bash
# update the CRD
# You can get the latest CRD by running `make manifests`
make manifests

# generate the dependency code
make generate
```

> Note, It will install the dependency tool `controller-gen` to your $GOPATH/bin directory.

## 4. Build and Push the container image

```bash
# replace the image name with your own
# replace meaglekey to the username of the Docker Hub account
docker build -t meaglekey/operator:v1.9.7-remove-appProtocol .

# push the image to the Docker Hub
# Image you have logged in to the Docker Hub by `docker login`
docker push meaglekey/operator:v1.9.7-remove-appProtocol
```
