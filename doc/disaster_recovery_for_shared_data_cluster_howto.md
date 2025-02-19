# Disaster Recovery for shared-data mode Cluster

From StarRocks 3.4.1, the shared-data mode cluster supports disaster recovery. The Operator can configure the disaster
recovery for the shared-data mode cluster to ensure the data security and high availability.

The following describes:

1. How the Operator supports disaster recovery.
2. Give an example of how to configure the disaster recovery.

## How does Operator support disaster recovery?

Operator adds the following fields:

```yaml
spec:
  disasterRecovery:
    generation: 1
    enabled: true

status:
  disasterRecoveryStatus:
    phase: todo/doing/done
    reason: ""
    observedGeneration: 1
    startTimestamp: xxx   # unix timestamp
    endTimestamp: yyy     # unix timestamp  
```

### When does the DR(disaster recovery) operation trigger?

1. The `enabled` field must be true.
2. The `generation` is a monotonically increasing integer that represents the expected state (spec) change version of
   the resource object. `observedGeneration` represents the version of the DR operation that has been executed. The
   Operator compares the values of `generation` and `observedGeneration`: when `disasterRecoveryStatus` is empty,
   or `observedGeneration < generation`, a new DR operation is triggered.
   If `generation == observedGeneration` && `disasterRecovery.Phase` != `v1.DRPhaseDone`, it means that after the last
   reconcile, the StarRocks cluster has entered disaster recovery mode.
3. Check the FE configuration file to ensure that the StarRocks cluster is started in shared-data mode.

If all the above conditions are met, the Operator will enter the DR mode.

### How does the Operator implement the DR process?

For BE component, the reconcile is paused.

For CN component, the reconcile is paused.

The reconcile process for FE is as follows:

1. Traverse `spec.starrocksFESpec.ConfigMaps` to confirm that `cluster_snapshot.yaml` has been mounted. Currently,
   this check is relatively simple, mainly to check whether the `SubPath` field is equal to `cluster_snapshot.yaml`.
2. Modify the FE Statefulset, including:
    1. Start a single-replica FE.
    2. Inject the `RESTORE_CLUSTER_GENERATION` and `RESTORE_CLUSTER_SNAPSHOT` environment variables. The former is
       used to determine the Generation to which the Pod belongs, and the latter is an environment variable passed to
       the FE module to trigger the disaster recovery operation of the FE Pod.
    3. Delete the startup/liveness configuration because the DR operation will take a long time.
    4. Modify the Readiness configuration. The configuration for normal cluster is to send an HTTP request to FE 8030;
       the new configuration is used to detect whether the FE 9030 port is connected. Once connected, it means that the
       FE Pod disaster recovery operation is complete.
3. After the FE Pod disaster recovery is complete, according to the configuration of the StarRocksCluster, StarRocks
   is started normally.

### How does Operator update the disaster recovery phase?

What is the phase of disaster recovery? In the status, `disasterRecoveryStatus.phase` represents the phase of the
disaster recovery, including `todo`, `doing`, `done`.

The status update logic is as follows:

1. The Operator detects that the disaster recovery mode is first entered (disasterRecoveryStatus is empty) or
   `observedGeneration < generation`. Then the disaster recovery mode enters the `todo` phase.
2. After the modified Statefulset is applied, update `disasterRecoveryStatus.phase` to the `doing` state. The duration
   of this state depends on the time it takes to complete the disaster recovery.
3. The Operator periodically checks the status of the FE Pod. First, confirm the `generation` to which it belongs;
   second, confirm whether the Pod is Ready.
4. After the FE Pod is Ready, update `disasterRecoveryStatus.phase` to the `done` mode.

## Example

Inorder to keep it simple and easy for users to follow this document, we use the `kube-starrocks` Helm Chart to deploy.
Please note:

1. Be sure to use at least v1.10.0 version of Operator and CRD.
2. Be sure to use at least 3.4.1 version of the StarRocks image to do disaster recovery.
3. Sensitive information is replaced by xxx, please set it to a reasonable value.

### 1. Create a normal working cluster

Prepare the `./starrocks-values.yaml` file:
> Note: we set `automated_cluster_snapshot_interval_seconds` to configure every minute to take a snapshot.

```yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.10.0
    imagePullPolicy: IfNotPresent
    replicaCount: 1
    resources:
      requests:
        cpu: 1m
        memory: 20Mi
starrocks:
  starrocksCluster:
    enabledBe: false
    enabledCn: true
  starrocksCnSpec:
    config: |
      sys_log_level = INFO
      # ports for admin, web, heartbeat service
      thrift_port = 9060
      webserver_port = 8040
      heartbeat_service_port = 9050
      brpc_port = 8060
    image:
      repository: starrocks/cn-ubuntu
      tag: 3.4.1
    replicas: 1
    resources:
      limits:
        cpu: 8
        memory: 8Gi
      requests:
        cpu: 1m
        memory: 10Mi
    storageSpec:
      name: cn
      logStorageSize: 1Gi
      storageSize: 10Gi
  starrocksFESpec:
    feEnvVars:
      - name: LOG_CONSOLE
        value: "1"
    config: |
      LOG_DIR = ${STARROCKS_HOME}/log
      DATE = "$(date +%Y%m%d-%H%M%S)"
      JAVA_OPTS="-Dlog4j2.formatMsgNoLookups=true -Xmx8192m -XX:+UseG1GC -Xlog:gc*:${LOG_DIR}/fe.gc.log.$DATE:time -XX:ErrorFile=${LOG_DIR}/hs_err_pid%p.log -Djava.security.policy=${STARROCKS_HOME}/conf/udf_security.policy"
      http_port = 8030
      rpc_port = 9020
      query_port = 9030
      edit_log_port = 9010
      mysql_service_nio_enabled = true
      sys_log_level = INFO
      run_mode = shared_data
      cloud_native_meta_port = 6090
      enable_load_volume_from_conf = true
      cloud_native_storage_type = S3
      aws_s3_path = xxx
      aws_s3_region = xxx
      aws_s3_endpoint = xxx
      aws_s3_access_key = xxx
      aws_s3_secret_key = xxx
      # we add this configuration because we want to get cluster snapshot quickly
      automated_cluster_snapshot_interval_seconds = 60
    replicas: 3
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.4.1
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 1m
        memory: 20Mi
    storageSpec:
      logStorageSize: 1Gi
      name: fe-storage
      storageSize: 10Gi
```

Create the cluster using Helm:

```bash
helm install -f ./starrocks-values.yaml starrocks starrocks-community/kube-starrocks

# make sure the cluster has been successfully deployed
kubectl get pods
NAME READY STATUS RESTARTS AGE
kube-starrocks-cn-0 1/1 Running 0 23s
kube-starrocks-fe-0 1/1 Running 0 79s
kube-starrocks-fe-1 1/1 Running 0 79s
kube-starrocks-fe-2 1/1 Running 0 79s
```

### 2. Create a table and insert data

Connect to the FE Pod:

```bash
# enter FE pod
kubectl exec -it kube-starrocks-fe-0 bash

# use mysql client to login
mysql -h 127.0.0.1 -P9030 -uroot
...
mysql>    
```

Execute the following SQL statement:

```sql
CREATE
DATABASE IF NOT EXISTS quickstart;

USE
quickstart;

-- create table
CREATE TABLE source_wiki_edit
(
    event_time     DATETIME,
    channel        VARCHAR(32)  DEFAULT '',
    user           VARCHAR(128) DEFAULT '',
    is_anonymous   TINYINT      DEFAULT '0',
    is_minor       TINYINT      DEFAULT '0',
    is_new         TINYINT      DEFAULT '0',
    is_robot       TINYINT      DEFAULT '0',
    is_unpatrolled TINYINT      DEFAULT '0',
    delta          INT          DEFAULT '0',
    added          INT          DEFAULT '0',
    deleted        INT          DEFAULT '0'
) DUPLICATE KEY(
   event_time,
   channel,user,
   is_anonymous,
   is_minor,
   is_new,
   is_robot,
   is_unpatrolled
)
PARTITION BY RANGE(event_time)(
PARTITION p06 VALUES LESS THAN ('2015-09-12 06:00:00'),
PARTITION p12 VALUES LESS THAN ('2015-09-12 12:00:00'),
PARTITION p18 VALUES LESS THAN ('2015-09-12 18:00:00'),
PARTITION p24 VALUES LESS THAN ('2015-09-13 00:00:00')
)
DISTRIBUTED BY HASH(user);

-- insert data
INSERT INTO source_wiki_edit
VALUES ("2015-09-12 00:00:00", "#en.wikipedia", "AustinFF", 0, 0, 0, 0, 0, 21, 5, 0),
       ("2015-09-12 00:00:00", "#ca.wikipedia", "helloSR", 0, 1, 0, 1, 0, 3, 23, 0),
       ("2015-09-12 08:00:00", "#ca.wikipedia", "helloSR", 0, 1, 0, 1, 0, 3, 23, 0);

-- select data
select *
from source_wiki_edit;
```

### 3. Generate Cluster Snapshot

Begin backup:

```sql
mysql
> ADMIN SET AUTOMATED CLUSTER SNAPSHOT ON STORAGE VOLUME builtin_storage_volume;
Query
OK, 0 rows affected (0.10 sec)
```

Wait for the backup to complete:

```sql
SELECT *
FROM INFORMATION_SCHEMA.CLUSTER_SNAPSHOT_JOBS \ G;
SNAPSHOT_NAME
: automated_cluster_snapshot_1739864377140
       JOB_ID: 13018
 CREATED_TIME: 2025-02-18 15:39:37
FINISHED_TIME: 2025-02-18 15:40:27
        STATE: FINISHED
  DETAIL_INFO:
ERROR_MESSAGE:

mysql>
SELECT *
FROM INFORMATION_SCHEMA.CLUSTER_SNAPSHOTS \ G;
*
************************** 1. row ***************************
     SNAPSHOT_NAME: automated_cluster_snapshot_1739864488333
     SNAPSHOT_TYPE: AUTOMATED
      CREATED_TIME: 2025-02-18 15:41:28
     FE_JOURNAL_ID: 1776
STARMGR_JOURNAL_ID: 126
        PROPERTIES:
    STORAGE_VOLUME: builtin_storage_volume
      STORAGE_PATH: s3://xxx/data/7351ce6a-f4a4-4937-a876-cb8801085aea/meta/image/automated_cluster_snapshot_1739864488333
1 row in set (0.03 sec)
```

Please note: because we set the backup interval to 1 minute, the backup path may be different from the above. The final
result can be viewed through s3:

```bash
s3cmd ls s3://xxx/data/7351ce6a-f4a4-4937-a876-cb8801085aea/meta/image/

sDIR s3://xxx/data/7351ce6a-f4a4-4937-a876-cb8801085aea/meta/image/automated_cluster_snapshot_1739858235830/
```

### 4. Delete the created cluster

```bash
helm uninstall starrocks

# delete pvcs
kubectl get pvc | awk '{if (NR>1){print $1}}' | xargs kubectl delete pvc
persistentvolumeclaim "cn-data-kube-starrocks-cn-0" deleted
persistentvolumeclaim "cn-log-kube-starrocks-cn-0" deleted
persistentvolumeclaim "fe-storage-log-kube-starrocks-fe-0" deleted
persistentvolumeclaim "fe-storage-log-kube-starrocks-fe-1" deleted
persistentvolumeclaim "fe-storage-log-kube-starrocks-fe-2" deleted
persistentvolumeclaim "fe-storage-meta-kube-starrocks-fe-0" deleted
persistentvolumeclaim "fe-storage-meta-kube-starrocks-fe-1" deleted
persistentvolumeclaim "fe-storage-meta-kube-starrocks-fe-2" deleted
```

### 5. Create a new cluster for disaster recovery

We will reuse the previous `starrocks-values.yaml` file, so be sure to ensure the security of this configuration file.
Prepare a new file named `override.yaml`, which contains the configuration required for disaster recovery.

```yaml
starrocks:
  starrocksCluster: # enable disaster recovery
    disasterRecovery:
      enabled: true
      generation: 1
  starrocksFESpec: # mount the cluster_snapshot.yaml
    configMaps:
      - name: cluster-snapshot
        mountPath: /opt/starrocks/fe/conf/cluster_snapshot.yaml
        subPath: cluster_snapshot.yaml

  configMaps:
    - name: cluster-snapshot
      data:
        cluster_snapshot.yaml: |
          # information about the cluster snapshot to be downloaded and restored
          cluster_snapshot:
              cluster_snapshot_path: s3://xxx/data/7351ce6a-f4a4-4937-a876-cb8801085aea/meta/image/automated_cluster_snapshot_1739858235830
              storage_volume_name: builtin_storage_volume

          # Operator will add the other FE followers automatically
          # just leave it blank
          frontends:

          # Operator will add the CN nodes automatically
          # just leave it blank
          compute_nodes:

          # used for restoring a cloned snapshot
          storage_volumes:
            - name: builtin_storage_volume
              type: S3
              location: s3://xxx/data
              comment: my s3 volume
              properties:
                - key: aws.s3.region
                  value: xxx
                - key: aws.s3.endpoint
                  value: xxx
                - key: aws.s3.access_key
                  value: xxx
                - key: aws.s3.secret_key
                  value: xxx
```

This time, we will deploy the cluster with the following command:

```bash
# Note:
# 1. make sure you are using the at least v1.10.0 version of Operator and CRD
# 2. the command to deploy the cluster is different from the first time. We specify two files, and the override.yaml
#    is used for the recovery configuration.
helm install -f ./starrocks-values.yaml -f override.yaml starrocks starrocks-community/kube-starrocks --version 1.10.0

```

The detailed process of disaster recovery is as follows:

1. The Operator will start a FE Pod and start the disaster recovery.
   ```
   kubectl get pods
   NAME                  READY   STATUS    RESTARTS        AGE
   kube-starrocks-fe-0   1/1     Running   0               4m37s
   ```
   If you check the status of the StarRocksCluster at this time, you will see:
   ```
   kubectl get src kube-starrocks -oyaml | less
   status:
     phase: running
     disasterRecovery:
       observedGeneration: 1
       phase: doing
       reason: disaster recovery is in progress
       startTimestamp: "1739860263"
   ```

2. After the disaster recovery is complete, the Operator will automatically start other Pods.
   ```
   kubectl get pods
   NAME                  READY   STATUS    RESTARTS        AGE
   kube-starrocks-cn-0   1/1     Running   0               7m54s
   kube-starrocks-fe-0   1/1     Running   0               7m1s
   kube-starrocks-fe-1   1/1     Running   0               7m54s
   kube-starrocks-fe-2   1/1     Running   0               7m54s
   ```
   The cluster status is as follows:
   ```
   status:
     phase: running
     disasterRecoveryStatus:
       endTimestamp: 1739861262
       observedGeneration: 1
       phase: done
       reason: disaster recovery is done
       startTimestamp: 1739860263
   ```

### Verify the disaster recovery

```shell
# enter the pod
kubectl exec -it kube-starrocks-fe-0 bash

# connect mysql
mysql -h 127.0.0.1 -P9030 -uroot

# get the data

mysql >USE quickstart
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql >select * from source_wiki_edit
```