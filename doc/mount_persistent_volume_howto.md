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
    image: "starrocks/fe-ubuntu:3.3-latest"
    replicas: 1
    storageVolumes:
    - name: fe-meta
      storageClassName: standard-rwo  # standard-rwo is the default storageClassName in GKE.
      # fe container stop running if the disk free space which the fe meta directory residents, is less than 5Gi.
      storageSize: 10Gi
      mountPath: /opt/starrocks/fe/meta
    - name: fe-log
      storageClassName: standard-rwo
      storageSize: 5Gi
      mountPath: /opt/starrocks/fe/log
  starRocksBeSpec:
    image: "starrocks/be-ubuntu:3.3-latest"
    replicas: 3
    storageVolumes:
    - name: be-data
      storageClassName: standard-rwo
      storageSize: 1Ti
      mountPath: /opt/starrocks/be/storage
    - name: be-log
      storageClassName: standard-rwo
      storageSize: 1Gi
      mountPath: /opt/starrocks/be/log
```

Note that the specific `storageClassName` should be available in kubernetes cluster before enabling this storageVolume
feature. If `StorageVolume` info is not specified in CRD spec, the operator will use emptydir mode to store FE meta and
BE data.

## 2. Mounting Persistent Volumes by Helm Chart

See [helm_repo_add_howto](./add_helm_repo_howto.md) to learn how to add the Helm Chart Repo for StarRocks. In this
guide, we will use `starrocks/kube-starrocks` chart to deploy both StarRocks operator and cluster.

### 2.1. Download the values.yaml file for the kube-starrocks chart

The values.yaml file contains the default configurations for the StarRocks Operator and the StarRocks cluster.

```Bash
helm show values starrocks/kube-starrocks > values.yaml
```

The following is a snippet of the values.yaml file:

```yaml
starrocks:
  starrocksFESpec: # fe storageSpec for persistent metadata.
    storageSpec:
      name: ""
      # the storageClassName represent the used storageclass name. if not set will use k8s cluster default storageclass.
      # you must set name when you set storageClassName
      # storageClassName: ""
      # the persistent volume size， default 10Gi.
      # fe container stop running if the disk free space which the fe meta directory residents, is less than 5Gi.
      storageSize: 10Gi
      # Setting this parameter can persist log storage
      logStorageSize: 5Gi

  starrocksBeSpec: # specify storageclass name and request size.
    storageSpec: # the name of volume for mount. if not will use emptyDir.
      name: ""
      # the storageClassName represent the used storageclass name. if not set will use k8s cluster default storageclass.
      # you must set name when you set storageClassName
      storageClassName: ""
      storageSize: 1Ti
      # Setting this parameter can persist log storage
      logStorageSize: 1Gi
```

### 2.2. Configure a YAML File with storageSpec settings

The following is an example of a custom values.yaml with storageSpec settings:

```yaml
starrocks:
   starrocksFESpec:
      image:
         repository: starrocks/fe-ubuntu
         tag: 3.3-latest
      storageSpec:
         name: fe-data
         storageClassName: standard-rwo   # standard-rwo is the default storageClassName in GKE.
         logStorageSize: 10Gi
         storageSize: 100Gi
   starrocksBeSpec:
      image:
         repository: starrocks/be-ubuntu
         tag: 3.3-latest
      replicas: 3
      storageSpec:
         name: be-storage
         storageClassName: standard-rwo
         logStorageSize: 10Gi
         storageSize: 500Gi
```

### 2.3. Deploy StarRocks Operator and Cluster

See [Install StarRocks by kube-starrocks chart](../helm-charts/charts/kube-starrocks/README.md) to learn how to deploy
StarRocks Operator and Cluster

## 3. Some Special storageClassName

Normally, the `storageClassName` is the name of the StorageClass that you want to use for the PersistentVolumeClaim.
We have also provided some special `storageClassName` for you to use:

1. `emptyDir`. It is a good choice when you want to mount a volume into the container for temporary usage, e.g. /tmp. Be aware that the files and directories written to the volume will be completely lost upon container restarting.
2. `hostPath`. It is a good choice when you want to the host's storage for the container, the data will be still there as along as the container is still running on the host. The data will be unavailable upon the container rescheduling to a different host. The typical scenario is to use it as cache volume. The `hostPath` field is required when this type is used.
   field.
   e.g.:
    ```yaml
        storageVolumes:
        - name: cn-cache
          storageClassName: "hostPath"
          hostPath:
            path: /storage
          mountPath: /storage   
    ```

> Note: In both cases, the `storageSize` field will be ignored.
