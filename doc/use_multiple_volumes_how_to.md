# Deploy StarRocks with Multiple Volumes

This document describes how to use multiple volumes to store StarRocks data.
> Note: After installation, you are not allowed to modify the volume related fields no matter in the CRD or Helm Chart.

# Deploy StarRocks with Multiple Volumes By Helm Chart

Based on the `storageSpec` field
in [values.yaml](https://github.com/StarRocks/starrocks-kubernetes-operator/blob/main/helm-charts/charts/kube-starrocks/values.yaml),
we will give an example of how to use multiple volumes to store StarRocks data.

```yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.9.8
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        cpu: 1m
        memory: 20Mi
starrocks:
  starrocksBeSpec:
    beEnvVars:
    # add storage_root_path in StarRocks config
    config: |
      be_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
      sys_log_level = INFO
      default_rowset_type = beta
      storage_root_path = /opt/starrocks/be/storage0;/opt/starrocks/be/storage1
    image:
      repository: starrocks/be-ubuntu
      tag: 3.2.2
    replicas: 1
    resources:
      limits:
        cpu: 8
        memory: 8Gi
      requests:
        cpu: 1m
        memory: 10Mi
    storageSpec:
      logStorageSize: 1Gi
      name: be-storage
      storageCount: 2   # specify the number of volumes
      storageSize: 10Gi # the size of each volume
  starrocksFESpec:
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.2.2
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 1m
        memory: 20Mi
    storageSpec:
      logStorageSize: 1Gi
      name: fe
      storageSize: 10Gi
```

Note:
1. add `storage_root_path` field in StarRocks config.
2. use `storageCount` to specify the number of volumes.
3. the `storage` directory still exists in the container, but will not be used to store data.

## What if I want to use different storageClass or storageSize for each volume?

This feature is not supported in Helm Chart. The following is a workaround:

```bash
# starrocks-community is a helm chart repository, you can show yours by `helm repo list`
# kube-starrocks is the name of the helm chart
helm template starrocks starrocks-community/kube-starrocks -f ./values.yaml >./sr.yaml

# From the sr.yaml, there will a Custom Resource Definition (CRD) named StarRocksCluster.
# You can modify the CRD to use different storageClass or storageSize for each volume.
storageVolumes:
- name: be0-data
storageClassName: "standard" # you can change the storageClassName
storageSize: "10Gi"          # you can change the storageSize
mountPath: /opt/starrocks/be/storage0
- name: be1-data
storageClassName: "standard"
storageSize: "10Gi"
mountPath: /opt/starrocks/be/storage1
- name: be-log
storageClassName:
storageSize: "1Gi"
mountPath: /opt/starrocks/be/log
- name: be-spill
storageClassName:
storageSize: "0Gi"
mountPath: /opt/starrocks/be/spill

# After modifying the CRD, you can apply it to your Kubernetes cluster.
kubectl apply -f sr.yaml
```