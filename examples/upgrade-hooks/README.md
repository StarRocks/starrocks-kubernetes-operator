# StarRocks Cluster Upgrade with Hooks Example

This example demonstrates how to use the upgrade hooks feature to perform automated pre-upgrade preparation.

## Files

- `starrockscluster-with-upgrade-hooks.yaml`: Complete cluster definition with upgrade hooks
- `upgrade-workflow.sh`: Script demonstrating the upgrade workflow

## Usage

### 1. Deploy the Cluster

```bash
kubectl apply -f starrockscluster-with-upgrade-hooks.yaml
```

### 2. Wait for Cluster Ready

```bash
kubectl wait --for=condition=Ready starrockscluster/starrockscluster-upgrade-example --timeout=600s
```

### 3. Trigger Upgrade Preparation

The cluster already has the preparation annotation, so upgrade preparation will be triggered automatically when you change the image version.

### 4. Upgrade the Cluster

```bash
# Update FE image
kubectl patch starrockscluster starrockscluster-upgrade-example \
  --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:3.2.2"}}}'

# Update BE image  
kubectl patch starrockscluster starrockscluster-upgrade-example \
  --type='merge' \
  -p '{"spec":{"starRocksBeSpec":{"image":"starrocks/be-ubuntu:3.2.2"}}}'
```

### 5. Monitor Progress

```bash
# Check upgrade preparation status
kubectl get starrockscluster starrockscluster-upgrade-example -o jsonpath='{.status.upgradePreparationStatus.phase}'

# Watch detailed status
watch kubectl get starrockscluster starrockscluster-upgrade-example -o yaml | grep -A 15 upgradePreparationStatus
```

### 6. Post-Upgrade Cleanup

```bash
# Re-enable services after upgrade
kubectl annotate starrockscluster starrockscluster-upgrade-example \
  starrocks.com/upgrade-hooks=enable-tablet-clone,enable-balancer

# Clean up preparation annotation
kubectl annotate starrockscluster starrockscluster-upgrade-example \
  starrocks.com/prepare-upgrade-
```

## Configuration Options

### Annotation-Based (Current Example)
```yaml
metadata:
  annotations:
    starrocks.com/prepare-upgrade: "true"
    starrocks.com/upgrade-hooks: "disable-tablet-clone,disable-balancer"
```

### Spec-Based (Alternative)
```yaml
spec:
  upgradePreparation:
    enabled: true
    timeoutSeconds: 300
    hooks:
    - name: disable-tablet-clone
      command: 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")'
      critical: true
```

## Available Predefined Hooks

- `disable-tablet-clone`: Stops tablet replication during upgrade
- `disable-balancer`: Disables cluster balancing during upgrade  
- `enable-tablet-clone`: Re-enables tablet replication after upgrade
- `enable-balancer`: Re-enables cluster balancing after upgrade

## Troubleshooting

### Check Hook Execution Status
```bash
kubectl get starrockscluster starrockscluster-upgrade-example -o jsonpath='{.status.upgradePreparationStatus}'
```

### View Operator Logs
```bash
kubectl logs -f deployment/starrocks-operator-controller-manager -n starrocks-operator-system
```

### Manual Hook Execution (if needed)
```bash
kubectl exec -it starrockscluster-upgrade-example-fe-0 -- mysql -h127.0.0.1 -P9030 -uroot -e "ADMIN SET FRONTEND CONFIG (\"tablet_sched_max_scheduling_tablets\" = \"0\")"
```
