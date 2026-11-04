# Scale in FE nodes Howto

FE nodes in a StarRocks cluster are used to store metadata. Normally, users do not need to scale in the FE nodes.
Incorrect operator may cause metadata inconsistency and malfunction. Be sure for every scale-in op, the number of the
offline FE nodes shall not be larger than the quorum. E.g. Don't try to scale-in directly from 7->3, user should first
scale-in 7->5, and then 5->3. For each scale in op, user should carefully check all the status of the remaining FE,
connect to the remaining FE node directly and run `SHOW FRONTENDS`, make sure all the FE nodes have consistent view of
the current FE nodes.

> Note: If you scale-in FE from 3->1, it will fail for sure, because the BDBJE HAGroup can't change from HA mode to
> single node mode automatically.

## What the operator does on an FE scale-in

When `starRocksFeSpec.replicas` is lowered, the StatefulSet removes the pods with the highest ordinals. Once the
scale-in has settled (the StatefulSet runs exactly `replicas` pods, all ready and on the current revision), the operator
connects to FE and, for every FE that belonged to a removed pod and is still listed in `SHOW FRONTENDS`:

1. deletes the PVC of that pod's FE meta volume (e.g. `fe-meta-kube-starrocks-fe-2`), because a dropped FE can not
   rejoin the cluster with its old metadata;
2. runs `ALTER SYSTEM DROP FOLLOWER` (or `DROP OBSERVER`) for it.

It only touches an FE whose host is the FQDN of a removed pod of this StatefulSet, that FE reports as not alive, and that
is not the leader. Without this step the removed FE stays registered as a dead member of the BDBJE group, which stops
BDBJE from cleaning its old journal files, so `fe-meta` keeps growing until FE runs out of disk. If the FE pod is added
back later by a scale-out, it starts with an empty meta volume and the FE entrypoint registers it again.

You still have to scale in step by step as described above, and check `SHOW FRONTENDS` after each step.

## How to scale in FE nodes manually

You can also drop the FE nodes yourself before lowering `replicas`. In that case the operator has nothing to drop. To
do so:

1. Execute the `SHOW FRONTENDS` command to get the FE nodes information, and must choose the
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

4. Repeat the above steps to remove other FE nodes until the desired number of FE nodes is reached. 4-->3.