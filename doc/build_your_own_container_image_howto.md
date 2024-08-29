# Build your own container image HOWTO

This document describes how to build your own container image for the StarRocks Operator. There are some reasons why you
might want to build your own container image, for example:

1. You want to build the Operator image based on the latest code.
2. You want to add fields to the StarRocksCluster CRD.
3. Your k8s version is too old to support the latest StarRocks Operator image.

## Prerequisites

1. Automated build tool `make` installed.
2. `git` installed.
3. `golang` installed.
4. [Docker](https://docs.docker.com/get-docker/) installed.
5. A container image repository, such as [Docker Hub](https://hub.docker.com/).

## 1. Build the container image based on the latest code

First, we describe how to build the container image based on the latest code on `main` branch.

```bash
# clone the repository
git clone https://github.com/StarRocks/starrocks-kubernetes-operator.git

# build your own image
docker build -t meaglekey/operator:your-branch-id .

# push the image to Container Registry
docker push meaglekey/operator:your-branch-id
```

## 2. Build the container image based on your own modification

In order to contextualize this process, we assume that your k8s does not support appProtocol, and you want to remove it
from the Operator.

## 2.1. Find the commit ID you want to revert

First, you need to find the PR that introduced the feature you want to remove. For example, we
find [this PR](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/288) by searching for `appProtocol`, and
its PR ID is 288.

Second, in the PR page, you can find the commit ID in the URL. For example, the commit ID
of [this PR](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/288) is `66e2f2c`.

## 2.2. Revert the commit

```bash
# clone the repository
git clone https://github.com/StarRocks/starrocks-kubernetes-operator.git

# checkout the latest released version, e.g., v1.9.8
git checkout v1.9.8

# revert the commit
git revert 66e2f2c

# During the Revert process, there may be conflicts. For conflicts in `_test.go`, you can directly delete the file.
git rm pkg/common/resource_utils/service_test.go
git revert --continue
```

## 2.3. Update CRD Definition and Dependency Code

You need to update the CRD definition and generate the dependency code. And it will install a dependency
tool `controller-gen` by `make` to your $GOPATH/bin directory.

```bash
# update the CRD definition
make manifests

# generate the dependency code
make generate
```

## 2.4. Build and Push the container image

```bash
# replace the image name with your own, e.g. replace meaglekey to the username of the Docker Hub account
docker build -t meaglekey/operator:v1.9.8-remove-appProtocol .

# push the image to the Docker Hub
# Note: you have to log in to the Docker Hub by `docker login` First.
docker push meaglekey/operator:v1.9.8-remove-appProtocol
```
