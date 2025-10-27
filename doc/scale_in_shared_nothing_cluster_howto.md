# Scale in Shared Nothing Cluster Howto

Now, if users want to scale in the shared-nothing cluster, they will adjust the `replicas` field. For example,

```yaml
spec:
  StarRocksBeSpec:
    replicas: 3  # 6->3
```

Unfortunately, the current implementation of StarRocks Operator does not
follow [the standard operation](https://docs.starrocks.io/docs/administration/management/Scale_up_down/) defined by
StarRocks. This document introduces:

- How does StarRocks Operator scale in BE nodes for the shared-nothing cluster
- How to fix the issue if users incorrectly scale in BE nodes for the shared-nothing cluster
- How to correctly scale in BE nodes for the shared-nothing cluster
- How to correctly scale in FE nodes for the shared-nothing cluster

## 1. How does StarRocks Operator scale in BE nodes for the shared-nothing cluster

When users adjust the `replicas` field to a smaller number, StarRocks Operator will **just modify the replicas field**
of the statefulset object.

For example, a user initially has 6 BE nodes:

```yaml
# because the statefulset name is kube-starrocks-be, the pod names are:
kube-starrocks-be-0
kube-starrocks-be-1
kube-starrocks-be-2
kube-starrocks-be-3
kube-starrocks-be-4
kube-starrocks-be-5
```

When the user scale in the cluster to 3 BE nodes, `kube-starrocks-be-5`, `kube-starrocks-be-4`, and
`kube-starrocks-be-3` pods will be deleted directly.
> The same to FE nodes.

## 2. How to fix the issue if users incorrectly scale in BE nodes for the shared-nothing cluster

Because StarRocks Operator does not follow the standard operation defined by StarRocks, if users scale in the
shared-nothing cluster, e.g. 6-->3, the data in the deleted BE nodes will be lost.

Because Operator did not delete the persistent volume claims (PVCs) of the deleted BE nodes, users can
recover the data by resetting the replicas field to the original number, e.g. 3-->6.

## 3. How to correctly scale in BE nodes for the shared-nothing cluster

To scale in the shared-nothing cluster correctly, users should follow the standard operation defined by StarRocks. For
example, if users want to scale in the BE nodes from 6 to 3, they should:
> Note: Scaling in the BE nodes one by one.

1. Execute the `show backends` command to get the BE nodes information, and must choose the
   `kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local` node with the highest ordinal to be removed
   first.
   ```sql
   mysql
   > show backends;
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   | BackendId | IP                                                                     | HeartbeatPort | BePort | HttpPort | BrpcPort | LastStartTime       | LastHeartbeat       | Alive | SystemDecommissioned | ClusterDecommissioned | TabletNum | DataUsedCapacity | AvailCapacity | TotalCapacity | UsedPct | MaxDiskUsedPct | ErrMsg | Version        | Status                                                 | DataTotalCapacity | DataUsedPct | CpuCores | MemLimit | NumRunningQueries | MemUsedPct | CpuUsedPct | DataCacheMetrics                                  | Location |
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   | 10003     | kube-starrocks-be-0.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:13:34 | true  | false                | false                 | 41        | 0.000 B          | 76.209 GB     | 182.280 GB    | 58.19 % | 58.19 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:12:47"} | 76.209 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.86 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10002     | kube-starrocks-be-1.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:13:34 | true  | false                | false                 | 42        | 0.000 B          | 76.209 GB     | 182.280 GB    | 58.19 % | 58.19 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:12:47"} | 76.209 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.89 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10001     | kube-starrocks-be-2.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:13:34 | true  | false                | false                 | 41        | 0.000 B          | 76.209 GB     | 182.280 GB    | 58.19 % | 58.19 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:12:47"} | 76.209 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.87 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10004     | kube-starrocks-be-3.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:13:34 | true  | false                | false                 | 42        | 0.000 B          | 76.209 GB     | 182.280 GB    | 58.19 % | 58.19 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:12:47"} | 76.209 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.88 %     | 0.1 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10005     | kube-starrocks-be-4.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:13:34 | true  | false                | false                 | 41        | 0.000 B          | 76.209 GB     | 182.280 GB    | 58.19 % | 58.19 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:12:47"} | 76.209 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.88 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10312     | kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 04:13:24 | 2025-10-17 04:13:34 | true  | false                | false                 | 0         | 0.000 B          | 75.922 GB     | 182.280 GB    | 58.35 % | 58.35 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:13:25"} | 75.922 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.81 %     | 0.4 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   6 rows in set (0.02 sec)
   ```

2. Set the `drop_backend_after_decommission` configuration to `false` to avoid automatic deletion of the backend after
   decommissioning.
   ```yaml
   ADMIN SET FRONTEND CONFIG ("drop_backend_after_decommission" = "false");
   ```

3. Execute the `decommission` command to decommission the chosen BE node.
   ```sql
   ALTER SYSTEM DECOMMISSION BACKEND "kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local:9050"
   ```

4. Execute the `show backends` command to check the decommission status of the BE node(
   kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local). If the value of `SystemDecommissioned` field is
   true and `TabletNum` is 0, the BE node is decommissioned successfully.
   ```sql
   mysql> SHOW BACKENDS;
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   | BackendId | IP                                                                     | HeartbeatPort | BePort | HttpPort | BrpcPort | LastStartTime       | LastHeartbeat       | Alive | SystemDecommissioned | ClusterDecommissioned | TabletNum | DataUsedCapacity | AvailCapacity | TotalCapacity | UsedPct | MaxDiskUsedPct | ErrMsg | Version        | Status                                                 | DataTotalCapacity | DataUsedPct | CpuCores | MemLimit | NumRunningQueries | MemUsedPct | CpuUsedPct | DataCacheMetrics                                  | Location |
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   | 10003     | kube-starrocks-be-0.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:24:39 | true  | false                | false                 | 41        | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:23:48"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.87 %     | 0.1 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10002     | kube-starrocks-be-1.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:24:39 | true  | false                | false                 | 41        | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:23:48"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.89 %     | 0.0 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10001     | kube-starrocks-be-2.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:24:39 | true  | false                | false                 | 42        | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:23:48"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.88 %     | 0.1 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10004     | kube-starrocks-be-3.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:24:39 | true  | false                | false                 | 42        | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:23:48"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.89 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10005     | kube-starrocks-be-4.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 03:05:43 | 2025-10-17 04:24:39 | true  | false                | false                 | 41        | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:23:48"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.89 %     | 0.1 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   | 10312     | kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local | 9050          | 9060   | 8040     | 8060     | 2025-10-17 04:13:24 | 2025-10-17 04:24:39 | true  | true                 | false                 | 0         | 0.000 B          | 75.850 GB     | 182.280 GB    | 58.39 % | 58.39 %        |        | 3.3.10-227b0b3 | {"lastSuccessReportTabletsTime":"2025-10-17 04:24:26"} | 75.850 GB         | 0.00 %      | 8        | 6.207GB  | 0                 | 2.85 %     | 0.2 %      | Status: Normal, DiskUsage: 0B/0B, MemUsage: 0B/0B |          |
   +-----------+------------------------------------------------------------------------+---------------+--------+----------+----------+---------------------+---------------------+-------+----------------------+-----------------------+-----------+------------------+---------------+---------------+---------+----------------+--------+----------------+--------------------------------------------------------+-------------------+-------------+----------+----------+-------------------+------------+------------+---------------------------------------------------+----------+
   ```
5. Execute the `ALTER SYSTEM DROP BACKEND` command to drop the decommissioned BE node from the StarRocks cluster.
   ```sql
   ALTER SYSTEM DROP BACKEND "kube-starrocks-be-5.kube-starrocks-be-search.default.svc.cluster.local:9050"
   ```
6. Adjust the `replicas` field to a smaller number, e.g. 6-->5.
7. Repeat the above steps to remove other BE nodes until the desired number of BE nodes is reached.

## 4. How to correctly scale in FE nodes for the shared-nothing cluster

For FE nodes, normally, users do not need to scale in the FE nodes. Scale-in FE with CAUTIONS, incorrect operator may
cause metadata inconsistency and malfunctioning. Be sure for every scale-in op, the number of the offline FE nodes shall
not be larger than the quorum. E.g. Don't try to scale-in directly from 7->3, user should first scale-in 7->5, and then
5->3. For each scale in op, user should carefully check all the status of the remaining FE, connect to the remaining FE
node directly and run SHOW FRONTENDS, make sure all the FE nodes have consistent view of the current FE nodes.

> Note: If you scale-in FE from 3->1, it will fail for sure, because the BDBJE HAGroup can't change from HA mode to
> single node mode automatically.

If users really want to scale in the FE nodes, they should:

1. Execute the `show frontends` command to get the FE nodes information, and must choose the
   `kube-starrocks-fe-4.kube-starrocks-fe-search.default.svc.cluster.local` node with the highest ordinal to be removed
   first.
   ```sql
   mysql
   > show frontends;
   +-------------------------------------------------------------------------------------------+------------------------------------------------------------------------+-------------+----------+-----------+---------+----------+------------+------+-------+-------------------+---------------------+----------+--------+---------------------+----------------+
   | Name                                                                                      | IP                                                                     | EditLogPort | HttpPort | QueryPort | RpcPort | Role     | ClusterId  | Join | Alive | ReplayedJournalId | LastHeartbeat       | IsHelper | ErrMsg | StartTime           | Version        |
   +-------------------------------------------------------------------------------------------+------------------------------------------------------------------------+-------------+----------+-----------+---------+----------+------------+------+-------+-------------------+---------------------+----------+--------+---------------------+----------------+
   | kube-starrocks-fe-1.kube-starrocks-fe-search.default.svc.cluster.local_9010_1760648075561 | kube-starrocks-fe-1.kube-starrocks-fe-search.default.svc.cluster.local | 9010        | 8030     | 9030      | 9020    | FOLLOWER | 1931503630 | true | true  | 1646              | 2025-10-17 04:55:55 | true     |        | 2025-10-17 04:54:47 | 3.3.10-227b0b3 |
   | kube-starrocks-fe-0.kube-starrocks-fe-search.default.svc.cluster.local_9010_1760641496296 | kube-starrocks-fe-0.kube-starrocks-fe-search.default.svc.cluster.local | 9010        | 8030     | 9030      | 9020    | LEADER   | 1931503630 | true | true  | 1647              | 2025-10-17 04:55:55 | true     |        | 2025-10-17 03:05:04 | 3.3.10-227b0b3 |
   | kube-starrocks-fe-2.kube-starrocks-fe-search.default.svc.cluster.local_9010_1760648073373 | kube-starrocks-fe-2.kube-starrocks-fe-search.default.svc.cluster.local | 9010        | 8030     | 9030      | 9020    | FOLLOWER | 1931503630 | true | true  | 1646              | 2025-10-17 04:55:55 | true     |        | 2025-10-17 04:54:46 | 3.3.10-227b0b3 |
   | kube-starrocks-fe-3.kube-starrocks-fe-search.default.svc.cluster.local_9010_1760648073373 | kube-starrocks-fe-3.kube-starrocks-fe-search.default.svc.cluster.local | 9010        | 8030     | 9030      | 9020    | FOLLOWER | 1931503630 | true | true  | 1646              | 2025-10-17 04:55:55 | true     |        | 2025-10-17 04:54:46 | 3.3.10-227b0b3 |
   | kube-starrocks-fe-4.kube-starrocks-fe-search.default.svc.cluster.local_9010_1760648073373 | kube-starrocks-fe-4.kube-starrocks-fe-search.default.svc.cluster.local | 9010        | 8030     | 9030      | 9020    | FOLLOWER | 1931503630 | true | true  | 1646              | 2025-10-17 04:55:55 | true     |        | 2025-10-17 04:54:46 | 3.3.10-227b0b3 |
   +-------------------------------------------------------------------------------------------+------------------------------------------------------------------------+-------------+----------+-----------+---------+----------+------------+------+-------+-------------------+---------------------+----------+--------+---------------------+----------------+
   ```

2. Drop the FE node from the StarRocks cluster.
   ```sql
   mysql> ALTER SYSTEM DROP FOLLOWER "kube-starrocks-fe-4.kube-starrocks-fe-search.default.svc.cluster.local:9010";
   Query OK, 0 rows affected (0.22 sec)
   ```

3. Adjust the `replicas` field to a smaller number, e.g. 5-->4.


4. Repeat the above steps to remove other FE nodes until the desired number of FE nodes is reached. 4-->3
   Note: You are not allowed to scale in the FE nodes to 1.