# Scale in FE nodes Howto

FE nodes in a StarRocks cluster are used to store metadata. Normally, users do not need to scale in the FE nodes.
Incorrect operator may cause metadata inconsistency and malfunctioning. Be sure for every scale-in op, the number of the
offline FE nodes shall not be larger than the quorum. E.g. Don't try to scale-in directly from 7->3, user should first
scale-in 7->5, and then 5->3. For each scale in op, user should carefully check all the status of the remaining FE,
connect to the remaining FE node directly and run `SHOW FRONTENDS`, make sure all the FE nodes have consistent view of
the current FE nodes.

> Note: If you scale-in FE from 3->1, it will fail for sure, because the BDBJE HAGroup can't change from HA mode to
> single node mode automatically.

## How to correctly scale in FE nodes

If users want to scale in the FE nodes, they should:

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