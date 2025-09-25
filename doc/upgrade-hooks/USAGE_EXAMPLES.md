# Usage Examples for StarRocks Operator Upgrade Hooks

This document provides comprehensive examples for using the upgrade hooks feature in the StarRocks Kubernetes Operator.

## Overview

The StarRocks Operator now supports pre-upgrade hooks that automatically execute SQL commands before image upgrades. This eliminates the need for manual intervention to prepare the cluster for upgrades.

## Method 1: Using Annotations

Apply the annotation to trigger upgrade preparation:

```bash
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer
```

Then proceed with the image upgrade:

```bash
kubectl -n starrocks patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:latest"}}}'
```

### Available Predefined Hooks

- `disable-tablet-clone`: Sets `tablet_sched_max_scheduling_tablets = "0"`
- `disable-balancer`: Sets `disable_balance = "true"`
- `enable-tablet-clone`: Sets `tablet_sched_max_scheduling_tablets = "2000"`
- `enable-balancer`: Sets `disable_balance = "false"`

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
    service:
      type: ClusterIP
      ports:
      - name: http-port
        port: 8030
        containerPort: 8030
      - name: rpc-port
        port: 9020
        containerPort: 9020
      - name: query-port
        port: 9030
        containerPort: 9030
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.1
    replicas: 3
    service:
      type: ClusterIP
      ports:
      - name: be-port
        port: 9060
        containerPort: 9060
      - name: webserver-port
        port: 8040
        containerPort: 8040
      - name: heartbeat-port
        port: 9050
        containerPort: 9050
```

## Complete Upgrade Workflow

### 1. Preparation Phase
```bash
# Apply upgrade preparation annotation
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer
```

### 2. Monitor Preparation Status
```bash
# Check preparation status
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.phase}'

# Watch detailed status
kubectl get starrockscluster starrockscluster-sample -o yaml | grep -A 20 upgradePreparationStatus
```

### 3. Execute Upgrade
Once preparation shows "Completed":
```bash
# Upgrade FE image
kubectl patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:3.2.2"}}}'

# Upgrade BE image
kubectl patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksBeSpec":{"image":"starrocks/be-ubuntu:3.2.2"}}}'
```

### 4. Post-Upgrade Cleanup
```bash
# Enable services after upgrade (manual or via hook)
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/upgrade-hooks=enable-tablet-clone,enable-balancer

# Clean up preparation annotation
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade-
```

## Advanced Examples

### Custom Hook Configuration
```yaml
spec:
  upgradePreparation:
    enabled: true
    timeoutSeconds: 600
    hooks:
    - name: backup-metadata
      command: 'BACKUP SNAPSHOT my_db.snapshot TO "s3://backup-bucket/snapshots/"'
      critical: false
    - name: disable-compaction
      command: 'ADMIN SET FRONTEND CONFIG ("disable_compaction" = "true")'
      critical: true
    - name: set-maintenance-mode
      command: 'ADMIN SET FRONTEND CONFIG ("maintenance_mode" = "true")'
      critical: true
```

### Multiple Hook Types
```bash
# Use both predefined and custom hooks
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer
```

## Monitoring and Troubleshooting

### Check Upgrade Status
```bash
# Get current phase
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.phase}'

# Get detailed status
kubectl get starrockscluster starrockscluster-sample -o yaml | yq e '.status.upgradePreparationStatus' -
```

### Monitor Events
```bash
# Watch for upgrade hook events
kubectl get events --field-selector involvedObject.name=starrockscluster-sample --sort-by='.lastTimestamp'

# Follow logs (if operator has verbose logging)
kubectl logs -f deployment/starrocks-operator-controller-manager -n starrocks-operator-system
```

### Troubleshooting Failed Hooks

If a hook fails:
```bash
# Check failed hooks
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.failedHooks}'

# Check failure reason
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.reason}'

# Manual intervention might be needed
kubectl exec -it starrockscluster-sample-fe-0 -n starrocks -- mysql -h127.0.0.1 -P9030 -uroot
```

### Reset Upgrade Preparation
```bash
# Clear all upgrade preparation state
kubectl patch starrockscluster starrockscluster-sample \
  --type='json' \
  -p='[{"op": "remove", "path": "/status/upgradePreparationStatus"}]'

# Remove annotations
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade- \
  starrocks.com/upgrade-hooks-
```

## Best Practices

1. **Test in Development**: Always test upgrade procedures in a development environment first
2. **Monitor Status**: Always check preparation status before proceeding with upgrades
3. **Critical Hooks**: Mark essential hooks as `critical: true` to prevent upgrades if they fail
4. **Timeout Configuration**: Set appropriate timeouts based on cluster size and expected execution time
5. **Backup**: Consider adding backup hooks before major upgrades
6. **Documentation**: Document custom hooks for your team

## Integration with CI/CD

### GitOps Example
```yaml
# upgrade-preparation.yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-prod
  annotations:
    starrocks.com/prepare-upgrade: "true"
    starrocks.com/upgrade-hooks: "disable-tablet-clone,disable-balancer"
spec:
  # ... existing spec
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2  # Updated version
```

### Automation Script
```bash
#!/bin/bash
CLUSTER_NAME="starrockscluster-sample"
NEW_VERSION="3.2.2"

# Prepare upgrade
kubectl annotate starrockscluster $CLUSTER_NAME \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer

# Wait for preparation
while [ "$(kubectl get starrockscluster $CLUSTER_NAME -o jsonpath='{.status.upgradePreparationStatus.phase}')" != "Completed" ]; do
  echo "Waiting for upgrade preparation..."
  sleep 30
done

# Execute upgrade
kubectl patch starrockscluster $CLUSTER_NAME \
  --type='merge' \
  -p "{\"spec\":{\"starRocksFeSpec\":{\"image\":\"starrocks/fe-ubuntu:$NEW_VERSION\"}}}"

echo "Upgrade initiated for version $NEW_VERSION"
```
