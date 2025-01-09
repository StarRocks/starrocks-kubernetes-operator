# Deploy StarRocks with Multiple Volumes

This document describes how to use multiple volumes to store StarRocks data.

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
  starrocksCluster:
    enabledBe: true
    enabledCn: true
  starrocksCnSpec:
    cnEnvVars:
    # add storage_root_path in StarRocks config
    config: |
      sys_log_level = INFO
      # ports for admin, web, heartbeat service
      thrift_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
      storage_root_path = /opt/starrocks/cn/storage0;/opt/starrocks/cn/storage1;/opt/starrocks/cn/storage2
    image:
      repository: starrocks/cn-ubuntu
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
      name: cn
      storageCount: 3   # specify the number of volumes
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