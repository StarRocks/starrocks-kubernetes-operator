# Mount external configMaps or secrets

StarRocks Kubernetes Operator supports mounting multiple external configmaps or secrets into StarRocks. This document
describes how to mount configmaps into StarRocks.
> You can mount secrets in the same way.

## 1. Mount configMaps by starRocks CRD YAML file

You can specify `configMaps` in the corresponding component spec. The following is an example

```shell
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: kube-starrocks
  namespace: kb-system
spec:
  starRocksFeSpec:
    image: "starrocks/fe-ubuntu:2.5.4"
    replicas: 1
    configMaps:
      - name: my-configmap
        mountPath: /etc/my-configmap
  starRocksBeSpec:
    image: "starrocks/be-ubuntu:2.5.4"
    replicas: 1
    configMaps:
      - name: my-configmap
        mountPath: /etc/my-configmap
```

> Note: The specific `ConfigMap` resources should be available in kubernetes cluster before enabling this feature.

## 2. Mount configMaps by helm chart

By using Helm chart, you can also mount multiple external configmaps into StarRocks. You can specify `configMaps` in
the corresponding component spec. The following is an example by using `kube-starrocks` Helm chart.

```shell
starrocks:
  starrocksBeSpec:
    configMaps:
      # mount the whole configmap `my-configmap` to `/etc/my-configmap`
      - name: my-configmap
        mountPath: /etc/my-configmap

  # a configmap named `my-configmap` will be created with the following content.
  configMaps:
  - name: my-configmap
    data:
      key.conf: |
        this is the content of the configmap
        when mounted, key will be the name of the file
```

## 3. Mount configMaps to a subPath by Helm Chart

You can also mount external configmaps into StarRocks with a subPath. The following is an example by
using `kube-starrocks` Helm chart.

```shell
starrocks:
  starrocksBeSpec:
    configMaps:
      # mount the file `key.conf` in configmap `my-configmap` to `/opt/starrocks/be/conf/key.conf`
      - name: my-configmap
        mountPath: /opt/starrocks/be/conf/key.conf
        subPath: key.conf

  # a configmap named `my-configmap` will be created with the following content.
  configMaps:
  - name: my-configmap
    data:
      key.conf: |
        this is the content of the configmap
        when mounted, key will be the name of the file
```

In `/opt/starrocks/be/conf`, the original file if existed will be overwritten, but other files will not be affected.
