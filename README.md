# StarRocks-Kubernetes-Operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> English | [中文](README_ZH-CN.md)

## Overview

Using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), a framework that enables the deployment of
StarRocks Custom Resource Definition (CRD) resources within a Kubernetes environment.

The Kubernetes Operator provided by StarRocks facilitates the deployment of StarRocks' Front End (FE), Back End (BE),
and Compute Node (CN) components within your Kubernetes environment. By default, these components operate in FQDN (fully
qualified domain name) mode.

## Requirements

* kubernetes 1.18+
* golang 1.18+

## Supported Features

* FE decouples with CN and BE. FE is a must-have component, BE and CN can be optionally deployed.
* Support v2 horizontalpodautoscalers for CN cluster.

## Install Operator in kubernetes

Apply the custom resource definition (CRD) for the Operator:

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/starrocks.com_starrocksclusters.yaml
```

Apply the Operator manifest. By default, the Operator is configured to install in the starrocks namespace.
To use the Operator in a custom namespace, download
the [Operator manifest](https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml)
and edit all instances of namespace: starrocks to specify your custom namespace.
Then apply this version of the manifest to the cluster with kubectl apply -f {local-file-path} instead of using the
command below.

```console
kubectl apply -f https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## Deploy the StarRocks cluster

You need to prepare a separate yaml file to deploy the StarRocks FE, BE and CN components.
The starrocks cluster CRD fields explains in [api.md](./doc/api.md).
The [examples](./examples/starrocks) directory contains some simple example for reference.

You can use any of the template yaml file as a starting point. You can further add more configurations into the template
yaml file following this deployment documentation.

For demonstration purpose, we use the [starrocks-fe-and-be.yaml](./examples/starrocks/starrocks-fe-and-be.yaml) example
template to start a 3 FE and 3 BE StarRocks cluster.

```console
kubectl apply -f starrocks-fe-and-be.yaml
```

## Connect to the deployed StarRocks Cluster

After deploying the StarRocks cluster, you can use `kubectl get svc -n <namespace>` to find the IP to connect to. For
example if the namespace that starrocks is deployed into is `starrocks`, you can:

```console
kubectl get svc -n starrocks
```

`<your-StarRocksCluster-name>-fe-service`'s clusterIP is the IP to use to connect to StarRocks FE.

## Stop the StarRocks cluster

Delete the custom resource:

```console
kubectl delete -f starrocks-fe-and-be.yaml
```

Remove the Operator:

```console
kubectl delete -f  https://raw.githubusercontent.com/StarRocks/starrocks-kubernetes-operator/main/deploy/operator.yaml
```

## Install StarRocks with Helm

StarRocks has embraced Helm for its deployment needs. You can find the Helm chart for StarRocks
at [artifacthub](https://artifacthub.io/packages/helm/kube-starrocks/kube-starrocks).

See [deploy_starrocks_with_helm.md](./doc/deploy_starrocks_with_helm_howto.md) for more details.

There are more documents in the [doc](./doc/README.md) directory.
