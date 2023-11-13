# StarRocks-Kubernetes-Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> English | [中文](README_ZH-CN.md)

## Overview

StarRocks Kubernetes Operator is a project that implements the deployment and operation of StarRocks, a next-generation
sub-second MPP OLAP database, on Kubernetes. It facilitates the deployment of StarRocks' Frontend (FE), Backend (BE),
and Compute Node (CN) components within your Kubernetes environment. It also includes Helm chart for easy installation
and configuration. With StarRocks Kubernetes Operator, you can easily manage the lifecycle of StarRocks clusters, such
as installing, scaling, upgrading etc.

## Prerequisites

1. Kubernetes version >= 1.18
2. Helm version >= 3.0

## Features

### Operator Features

- Support deploying StarRocks FE, BE and CN components separately
  FE component is a must-have component, BE and CN components can be optionally deployed
- Support multiple StarRocks clusters in one Kubernetes cluster
- Support external clients outside the network of kubernetes to load data into StarRocks using STREAM LOAD
- Support automatic scaling for CN nodes based on CPU and memory usage
- Support mounting persistent volumes for StarRocks containers

### Helm Chart Features

- Support Helm Chart for easy installation and configuration
    - using kube-starrocks Helm chart to install both operator and StarRocks cluster
    - using operator Helm Chart to install operator, and using starrocks Helm Chart to install starrocks cluster
- Support initializing the password of root in your StarRocks cluster during installation.
- Support integration with other components in the Kubernetes ecosystem, such as Prometheus, Datadog, etc.

## Installation

In order to use StarRocks in Kubernetes, you need to install:

1. StarRocksCluster CRD
2. StarRocks Operator
3. StarRocksCluster CR

There are two ways to install Operator and StarRocks Cluster.

1. Install Operator and StarRocks Cluster by yaml Manifest.
2. Install Operator and StarRocks Cluster by Helm Chart.

> Note: In every release, we will provide the latest version of the yaml Manifest and Helm Chart. You can find them
> in https://github.com/StarRocks/starrocks-kubernetes-operator/releases

## Installation by yaml Manifest

Please see [Deploy StarRocks With Operator](./doc/deploy_starrocks_with_operator_howto.md) document for more details.

### 1. Apply the StarRocksCluster CRD

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```

### 2. Apply the Operator manifest

Apply the Operator manifest. By default, the Operator is configured to install in the starrocks namespace. To use the
Operator in a custom namespace, download
the [Operator manifest](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml)
and edit all instances of namespace: starrocks to specify your custom namespace.
Then apply this version of the manifest to the cluster with kubectl apply -f {local-file-path} instead of using the
command below.

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

### 3. Deploy the StarRocks cluster

You need to prepare a separate yaml file to deploy the StarRocks. The starrocks cluster CRD fields explains
in [api.md](./doc/api.md). The [examples](./examples/starrocks) directory contains some simple example for reference.

You can use any of the template yaml file as a starting point. You can further add more configurations into the template
yaml file following this deployment documentation.

For demonstration purpose, we use the [starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml) example
template to start a 3 FE and 3 BE StarRocks cluster.

```console
kubectl apply -f starrocks-fe-and-be.yaml
```

## Installation by Helm Chart

Please see [kube-starrocks](./helm-charts/charts/kube-starrocks/README.md) for how to install both operator and
StarRocks cluster by Helm Chart.

If you want more flexibility in managing your StarRocks clusters, you can deploy Operator
using [operator](./helm-charts/charts/kube-starrocks/charts/operator) Helm Chart and StarRocks
using [starrocks](./helm-charts/charts/kube-starrocks/charts/starrocks) Helm Chart separately.

## Other Documents

- In [doc](./doc) directory, you can find more documents about how to use StarRocks Operator.
- In [examples](./examples/starrocks) directory, you can find more examples about how to write StarRocksCluster CR.
