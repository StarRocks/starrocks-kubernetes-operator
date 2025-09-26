# Usage Examples for StarRocks Operator Upgrade Hooks

## Method 1: Using Annotations

Apply the annotation to trigger upgrade preparation:

```bash
kubectl annotate starrockscluster starrockscluster-sample   starrocks.com/prepare-upgrade=true   starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer
```

Then proceed with the image upgrade:

```bash
kubectl -n starrocks patch starrockscluster starrockscluster-sample   --type='merge'   -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:latest"}}}'
```

## Method 2: Using Spec Configuration

Define upgrade hooks in the StarRocksCluster spec:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
  namespace: starrocks
spec:
  upgradePreparation:
    enabled: true
    timeoutSeconds: 300
    hooks:
    - name: disable-tablet-clone
      command: 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")'
      critical: true
    - name: disable-balancer
      command: 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "0")'
      critical: true
    - name: disable-balance-flags
      command: 'ADMIN SET FRONTEND CONFIG ("disable_balance"="true")'
      critical: true
    - name: disable-colocate-balance
      command: 'ADMIN SET FRONTEND CONFIG ("disable_colocate_balance"="true")'
      critical: true
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.1
    replicas: 3
    # ... rest of FE spec
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.1
    replicas: 3
    # ... rest of BE spec
```

## Checking Upgrade Status

Monitor the upgrade preparation status:

```bash
kubectl get starrockscluster starrockscluster-sample -o yaml |   grep -A 20 upgradePreparationStatus
```

## Manual Cleanup (if needed)

If you need to manually cleanup upgrade preparation:

```bash
kubectl annotate starrockscluster starrockscluster-sample   starrocks.com/prepare-upgrade-
```

## Monitoring Events

Watch for upgrade hook events:

```bash
kubectl get events --field-selector involvedObject.name=starrockscluster-sample
```
