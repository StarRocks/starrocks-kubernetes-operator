# Leader-aware FE upgrade Howto

By default the FE component is deployed as a StatefulSet with the `RollingUpdate` strategy. The StatefulSet
controller replaces pods from the highest ordinal down and knows nothing about BDBJE roles. If the FE leader
is not the pod with the lowest ordinal, the election triggered by its restart can hand leadership to a pod
that is itself replaced a moment later, so a single upgrade causes two leader failovers.

The operator can order the replacement itself so that every upgrade causes exactly one leader election.

## How it works

When the operator runs with `--enable-fe-leader-aware-upgrade=true`:

1. The FE StatefulSet is switched to the `OnDelete` update strategy. Any `updateStrategy` set on
   `spec.starRocksFeSpec` is ignored while the flag is on. With `OnDelete` the StatefulSet controller only
   recreates a pod on the new template after that pod has been deleted.
2. Whenever the FE pod template changes, the operator compares each pod's revision with the StatefulSet's
   update revision. For the pods that are behind, it runs `SHOW FRONTENDS` through the FE service and
   deletes **one** pod per reconcile:
   - followers first, highest ordinal first;
   - the FE reported as `LEADER` only once it is the last pod left behind.
3. It waits for every FE pod to be ready before deleting the next one, and re-reads the leader on every
   pass, so a leadership change in the middle of a rollout still leaves the then-current leader for the end.
4. BE and CN wait for the FE rollout to finish, as they already do today.

The operator refuses to delete anything, and retries on the next reconcile, when:

- `SHOW FRONTENDS` cannot be run (FE unreachable, wrong root password);
- any FE reports `Alive = false`;
- no FE, or more than one FE, reports as `LEADER`;
- the leader cannot be matched to an FE pod.

Each deletion is recorded as a `FeLeaderAwareUpgrade` event on the StarRocksCluster.

## Enable it

With Helm, set the value on the operator chart and upgrade the release:

```yaml
starrocksOperator:
  enableFeLeaderAwareUpgrade: true
```

With the plain manifest, set `--enable-fe-leader-aware-upgrade=true` in the operator container args of
`deploy/operator.yaml`.

Both forms also grant the operator the `delete` verb on pods, which the feature needs.

## Root password

The operator connects to FE as `root`. If the FE pods carry a `MYSQL_PWD` environment variable (see
[change_root_password_howto.md](./change_root_password_howto.md)), that value is used, including when it
comes from a secret. Otherwise an empty password is used. The connection honours `--fe-ssl-mode`, see
[connect_to_ssl_enabled_fe_howto.md](./connect_to_ssl_enabled_fe_howto.md).

## Disable it

Set the flag back to `false` and restart the operator. The FE StatefulSet goes back to the strategy from the
CR (or the default `RollingUpdate`) on the next reconcile.
