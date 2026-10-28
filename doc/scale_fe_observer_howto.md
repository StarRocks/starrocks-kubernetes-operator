# Scale FE observer Howto

## Overview

FE observers are mainly used to increase the query concurrency of a StarRocks cluster. Unlike FE followers, observers do
not participate in leader election, so adding observers does not add leader selection pressure to the cluster.

Use `starRocksFeSpec.observerSpec` when the cluster needs more FE read capacity without changing the FE follower quorum.
Observers can be scaled with the read workload, while FE followers should still be scaled carefully because they store
metadata and participate in quorum.

## How to correctly scale out FE read workloads by observer

FE observers can be used to scale FE read workloads without changing the FE follower quorum. To use
`starRocksFeSpec.observerSpec`, the FE image configured in `starRocksFeSpec.image` must be StarRocks 4.1.0 or later.
The operator rejects FE observer deployment for older FE image versions because those images do not enable FE observer.

To scale FE observers out, enable `observerSpec` and set the desired observer count:

```yaml
starRocksFeSpec:
   image: "starrocks/fe-ubuntu:4.1.0"
   observerSpec:
      enabled: true
      observerNumber: 2
```

## How to correctly scale in FE observer

Known limitation: neither the operator nor `fe_entrypoint.sh` runs `ALTER SYSTEM DROP OBSERVER` when observers are
removed. If you decrease `starRocksFeSpec.observerSpec.observerNumber` or set
`starRocksFeSpec.observerSpec.enabled: false`, the observer Pods are removed from Kubernetes, but the removed observers
can remain in StarRocks metadata and appear as permanent `Alive=false` rows in `SHOW FRONTENDS`. Because observers are
intended for dynamic read-workload scaling, stale observer rows can accumulate after repeated scale-in operations.

For planned observer scale-in, manually drop the observers that will be removed before lowering `observerNumber`.

1. Execute `SHOW FRONTENDS` and find the `OBSERVER` rows for the observer Pods that will be removed.

2. Drop each observer from StarRocks metadata by using its FE host and `EditLogPort` from `SHOW FRONTENDS`.
   ```sql
   mysql> ALTER SYSTEM DROP OBSERVER "<observer-host>:<edit-log-port>";
   ```

   For example:
   ```sql
   mysql> ALTER SYSTEM DROP OBSERVER "kube-starrocks-fe-observer-1.kube-starrocks-fe-observer-search.default.svc.cluster.local:9010";
   ```

3. Lower `starRocksFeSpec.observerSpec.observerNumber`, or set `starRocksFeSpec.observerSpec.enabled: false` if all
   observers should be removed.

4. Execute `SHOW FRONTENDS` again and verify the removed observer rows are gone.

If the observer Pods were already removed from Kubernetes, connect to an FE node, run `SHOW FRONTENDS`, find `OBSERVER`
rows with `Alive=false`, and run `ALTER SYSTEM DROP OBSERVER "<observer-host>:<edit-log-port>"` for each stale row.
