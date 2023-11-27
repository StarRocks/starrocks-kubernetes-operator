# Mount Persistent Volume

StarRocks Kubernetes Operator supports mounting persistent volumes to StarRocks FE and BE pods. If not specified, the
operator will use emptyDir mode to store FE meta and BE data. **When container restarts, the data will be lost.**

This document describes how to mount persistent volumes to StarRocks FE and BE pods. There are two ways to mount
persistent volumes to StarRocks FE and BE pods:

1. Mounting persistent volumes to StarRocks FE and BE pods by the StarRocks CRD YAML file.
2. Mounting persistent volumes to StarRocks FE and BE pods by Helm chart.

> Note: Starrocks Operator will create a new PVC for each storageVolume. You should not create PVC manually.

## 1. Mounting Persistent Volumes by StarRocks CRD YAML File

If you want to use external storage to store FE meta and BE data for persistence, you can specify `storageVolumes` in
the corresponding component spec.

The following is an example of mounting persistent volumes to StarRocks FE and BE.

```bash
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: kube-starrocks
  namespace: starrocks
  labels:
    cluster: kube-starrocks
spec:
  starRocksFeSpec:
    image: "starrocks/fe-ubuntu:3.1-latest"
    replicas: 1
    storageVolumes:
    - name: fe-storage-meta
      storageClassName: data-storage
      # fe container stop running if the disk free space which the fe meta directory residents, is less than 5Gi.
      storageSize: 10Gi
      mountPath: /opt/starrocks/fe/meta
    - name: fe-storage-log
      storageClassName: data-storage
      storageSize: 5Gi
      mountPath: /opt/starrocks/fe/log
  starRocksBeSpec:
    image: "starrocks/be-ubuntu:3.1-latest"
    replicas: 1
    storageVolumes:
    - name: be-storage-data
      storageClassName: data-storage
      storageSize: 1Ti
      mountPath: /opt/starrocks/be/storage
    - name: be-storage-log
      storageClassName: data-storage
      storageSize: 1Gi
      mountPath: /opt/starrocks/be/log
```

Note that the specific `storageClassName` should be available in kubernetes cluster before enabling this storageVolume
feature. If `StorageVolume` info is not specified in CRD spec, the operator will use emptydir mode to store FE meta and
BE data.

## 2. Mounting Persistent Volumes by Helm Chart

See [helm_repo_add_howto](./add_helm_repo_howto.md) to learn how to add the Helm Chart Repo for StarRocks. In this
guide, we will use `starrocks-community/kube-starrocks` chart to deploy both StarRocks operator and cluster.

### 2.1. Download the values.yaml file for the kube-starrocks chart

The values.yaml file contains the default configurations for the StarRocks Operator and the StarRocks cluster.

```Bash
helm show values starrocks-community/kube-starrocks > values.yaml
```

The following is a snippet of the values.yaml file:

```yaml
starrocks:
  starrocksFESpec:
    # fe storageSpec for persistent metadata.
    storageSpec:
      name: ""
      # the storageClassName represent the used storageclass name. if not set will use k8s cluster default storageclass.
      # you must set name when you set storageClassName
      storageClassName: ""
      # the persistent volume size， default 10Gi.
      # fe container stop running if the disk free space which the fe meta directory residents, is less than 5Gi.
      storageSize: 10Gi
      # Setting this parameter can persist log storage
      logStorageSize: 5Gi

  starrocksBeSpec:
    # specify storageclass name and request size.
    storageSpec:
      # the name of volume for mount. if not will use emptyDir.
      name: ""
      # the storageClassName represent the used storageclass name. if not set will use k8s cluster default storageclass.
      # you must set name when you set storageClassName
      storageClassName: ""
      storageSize: 1Ti
      # Setting this parameter can persist log storage
      logStorageSize: 1Gi
```

### 2.2. Configure a YAML File for Custom Settings.

You can configure the following parameters in the YAML file:

```yaml
starrocks:
  starrocksFESpec:
    storageSpec:
      name: ""
      storageSize: 10Gi
      logStorageSize: 5Gi

  starrocksBeSpec:
    storageSpec:
      name: ""
      storageSize: 1Ti
      logStorageSize: 5Gi
```

### 2.3. Deploy StarRocks Operator and Cluster

See [Install StarRocks by kube-starrocks chart](../helm-charts/charts/kube-starrocks/README.md) to learn how to deploy
StarRocks Operator and Cluster
