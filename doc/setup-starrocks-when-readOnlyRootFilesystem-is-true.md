# Background

StarRocks has three components: Frontend (FE), Backend (BE), and Compute Node(CN). When the `readOnlyRootFilesystem` is
set to `true`, the components of StarRocks cannot start normally. This is because the components of StarRocks write data
to the disk, and the `readOnlyRootFilesystem` setting prevents the components from writing data to the disk.

For the FE component, FE writes data to the following directories:

```bash
# in fe directory
drwxr-xr-x 2 root      root      4.0K Nov 19 11:27 plugins
drwxr-xr-x 4 root      root      4.0K Nov 19 11:27 temp_dir

# in fe/bin directory
-rw-r--r-- 1 root      root         2 Nov 19 11:27 fe.pid

# in fe/conf directory
lrwxrwxrwx 1 root      root        30 Nov 19 11:27 fe.conf -> /etc/starrocks/fe/conf/fe.conf
```

For the BE component, BE writes data to the following directories:

```bash
# in be directory
drwxr-xr-x 2 root      root      4.0K Nov 19 11:27 spill

# in be/conf directory
lrwxrwxrwx 1 root      root        30 Nov 19 11:27 be.conf -> /etc/starrocks/be/conf/be.conf

# in be/bin directory
-rw-r----- 1 root      root         3 Nov 19 11:27 be.pid

# in be/lib directory
drwxr-xr-x   2 root      root      4.0K Nov 19 11:27 jdbc_drivers
drwxr-xr-x   2 root      root      4.0K Nov 19 11:27 small_file
drwxr-xr-x 130 root      root      4.0K Nov 19 11:27 udf
drwxr-xr-x   2 root      root      4.0K Nov 19 11:27 udf-runtime
```

This document describes how to set up StarRocks when the `readOnlyRootFilesystem` field is set to `true`.

# How

We create and mount a volume, and in the entrypoint script, we will copy everything from the original directory to the
mounted volume. This way, the components of StarRocks can write data to the mounted volume.

> Note: you should use the operator version `v1.9.9` or later.

# Steps

There are two ways to deploy StarRocks cluster:

1. Deploy StarRocks cluster with `StarRocksCluster` CR yaml.
2. Deploy StarRocks cluster with Helm chart.

Therefore, there are two ways to set up StarRocks when the `readOnlyRootFilesystem` field is set to `true`.

## By using StarRocksCluster CR yaml

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: kube-starrocks
  namespace: starrocks
spec:
  starRocksFeSpec:
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    configMapInfo:
      configMapName: kube-starrocks-fe-cm
      resolveKey: fe.conf
    storageVolumes:
    - mountPath: /opt/starrocks-artifacts
      name: fe-artifacts
      storageClassName: emptyDir
      storageSize: 20Gi
    - mountPath: /opt/starrocks-meta
      name: fe-meta
      storageSize: 10Gi
    - mountPath: /opt/starrocks-log
      name: fe-log
      storageSize: 10Gi
    command: ["bash", "-c"]
    args:
      - cp -r /opt/starrocks/* /opt/starrocks-artifacts && exec /opt/starrocks-artifacts/fe_entrypoint.sh $FE_SERVICE_NAME
    feEnvVars:
    - name: STARROCKS_ROOT
      value: /opt/starrocks-artifacts
    image: starrocks/fe-ubuntu:3.2.2
    imagePullPolicy: IfNotPresent
    replicas: 1
    requests:
      cpu: 1m
      memory: 22Mi
  starRocksBeSpec:
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    configMapInfo:
      configMapName: kube-starrocks-be-cm
      resolveKey: be.conf
    storageVolumes:
    - mountPath: /opt/starrocks-artifacts
      name: be-artifacts
      storageClassName: emptyDir
      storageSize: 20Gi
    - mountPath: /opt/starrocks-storage
      name: be-storage  # the name must be this
      storageSize: 10Gi
    - mountPath: /opt/starrocks-log
      name: be-log  # the name must be this
      storageSize: 10Gi
    command: ["bash", "-c"]
    args:
      - cp -r /opt/starrocks/* /opt/starrocks-artifacts && exec /opt/starrocks-artifacts/be_entrypoint.sh $FE_SERVICE_NAME
    beEnvVars:
    - name: STARROCKS_ROOT
      value: /opt/starrocks-artifacts
    image: starrocks/be-ubuntu:3.2.2
    imagePullPolicy: IfNotPresent
    replicas: 2
    requests:
      cpu: 1m
      memory: 10Mi

---

apiVersion: v1
data:
  fe.conf: |
    LOG_DIR = ${STARROCKS_HOME}/log
    DATE = "$(date +%Y%m%d-%H%M%S)"
    JAVA_OPTS="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseG1GC -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time"
    http_port = 8030
    rpc_port = 9020
    query_port = 9030
    edit_log_port = 9010
    mysql_service_nio_enabled = true
    sys_log_level = INFO
    
    # config for meta and log
    meta_dir = /opt/starrocks-meta
    dump_log_dir = /opt/starrocks-log
    sys_log_dir = /opt/starrocks-log
    audit_log_dir = /opt/starrocks-log
kind: ConfigMap
metadata:
  name: kube-starrocks-fe-cm
  namespace: starrocks

---

apiVersion: v1
data:
  be.conf: |
    be_port = 9060
    webserver_port = 8040
    heartbeat_service_port = 9050
    brpc_port = 8060
    sys_log_level = INFO
    default_rowset_type = beta

    # config for storage and log
    storage_root_path = /opt/starrocks-storage
    sys_log_dir = /opt/starrocks-log
kind: ConfigMap
metadata:
  name: kube-starrocks-be-cm
  namespace: starrocks
```

## By using Helm Chart

If you are using the `kube-starrocks` Helm chart, add the following snippets to `values.yaml`.

> Note: you should use the chart version `v1.9.9` or later.

```yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.9.9-rc1
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        cpu: 1m
        memory: 20Mi
starrocks:
  starrocksFESpec:
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    storageSpec:
      name: fe
      storageSize: 10Gi
      storageMountPath: /opt/starrocks-meta
      logStorageSize: 10Gi
      logMountPath: /opt/starrocks-log
    emptyDirs:
    - name: fe-artifacts
      mountPath: /opt/starrocks-artifacts
    entrypoint:
      script: |
        #! /bin/bash
        cp -r /opt/starrocks/* /opt/starrocks-artifacts
        exec /opt/starrocks/fe_entrypoint.sh $FE_SERVICE_NAME
    feEnvVars:
    - name: STARROCKS_ROOT
      value: /opt/starrocks-artifacts
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
    config: |
      LOG_DIR = ${STARROCKS_HOME}/log
      DATE = "$(date +%Y%m%d-%H%M%S)"
      JAVA_OPTS="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseG1GC -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time"
      http_port = 8030
      rpc_port = 9020
      query_port = 9030
      edit_log_port = 9010
      mysql_service_nio_enabled = true
      sys_log_level = INFO
      # config for meta and log
      meta_dir = /opt/starrocks-meta
      dump_log_dir = /opt/starrocks-log
      sys_log_dir = /opt/starrocks-log
      audit_log_dir = /opt/starrocks-log
  starrocksBeSpec:
    readOnlyRootFilesystem: true
    runAsNonRoot: true
    storageSpec:
      name: be
      storageSize: 10Gi
      storageMountPath: /opt/starrocks-storage
      logStorageSize: 10Gi
      logMountPath: /opt/starrocks-log
      spillStorageSize: 10Gi
      spillMountPath: /opt/starrocks-spill
    emptyDirs:
    - name: be-artifacts
      mountPath: /opt/starrocks-artifacts
    entrypoint:
      script: |
        #! /bin/bash
        cp -r /opt/starrocks/* /opt/starrocks-artifacts 
        exec /opt/starrocks-artifacts/be_entrypoint.sh $FE_SERVICE_NAME
    beEnvVars:
    - name: STARROCKS_ROOT
      value: /opt/starrocks-artifacts
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
    config: |
      be_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
      sys_log_level = INFO
      default_rowset_type = beta
      # config for storage and log
      storage_root_path = /opt/starrocks-storage
      sys_log_dir = /opt/starrocks-log
      spill_local_storage_dir = /opt/starrocks-spill
```