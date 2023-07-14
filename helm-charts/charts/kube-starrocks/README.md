# kube-starrocks

Install the kube-starrocks, a collection of Kubernetes manifests to provide easy to operate end-to-end Kubernetes
cluster deploy starrocks cluster use [starrocks Operator](https://github.com/StarRocks/starrocks-kubernetes-operator).

## Prerequisites

- Kubernetes 1.18.3+
- Helm 3+

## Get Helm Repository Info

```console
helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
helm repo update
```

_See [`helm repo`](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Helm Chart

1. view the package names in repo.
   ```console
   helm search repo starrocks-community
   ```
2. install specify package.
   ```
   helm install [RELEASE_NAME] starrocks-community/[PACKAGE_NAME]
   ```

_See [configuration](#configuration) below._

## Uninstall Helm Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

CRDs created by this chart are not removed by default and should be manually cleaned up:

```console
kubectl delete crd starrocksclusters.starrocks.com
```

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing).
To see all configurable options with detailed comments:

```console
helm show values starrocks-community/kube-starrocks
```

## Documentation

You can see our documentation at StarRocks website for more in-depth installation and instructions for production:

- [English](https://docs.starrocks.io/en-us/latest/introduction/StarRocks_intro)
- [简体中文](https://docs.starrocks.io/zh-cn/latest/introduction/StarRocks_intro)