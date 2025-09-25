# Usage Examples for StarRocks Operator Upgrade Hooks

This document provides examples of how to use the upgrade hooks feature to automate StarRocks cluster upgrades.

## Overview

The StarRocks Operator upgrade hooks feature allows you to execute SQL commands automatically before performing image upgrades. This is essential for safely upgrading StarRocks clusters by disabling tablet scheduling and balancing operations.

## Method 1: Using Annotations

The quickest way to trigger upgrade preparation is using annotations:

```bash
# Apply the upgrade preparation annotation
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer
```

Then proceed with the image upgrade:

```bash
# Upgrade FE image
kubectl -n starrocks patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:latest"}}}'

# Upgrade BE image
kubectl -n starrocks patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksBeSpec":{"image":"starrocks/be-ubuntu:latest"}}}'
```

### Available Predefined Hooks

- `disable-tablet-clone`: Disables tablet cloning/scheduling
- `disable-balancer`: Disables cluster balancing
- `enable-tablet-clone`: Re-enables tablet cloning (post-upgrade)
- `enable-balancer`: Re-enables cluster balancing (post-upgrade)

## Method 2: Using Spec Configuration

For more control, define upgrade hooks directly in the StarRocksCluster spec:

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
      command: 'ADMIN SET FRONTEND CONFIG ("disable_balance"="true")'
      critical: true
    - name: custom-preparation
      command: 'ADMIN SET FRONTEND CONFIG ("max_clone_task_timeout_second" = "7200")'
      critical: false
  
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.1
    replicas: 3
    service:
      type: ClusterIP
      ports:
      - name: http
        port: 8030
        containerPort: 8030
      - name: rpc
        port: 9020
        containerPort: 9020
      - name: query
        port: 9030
        containerPort: 9030
  
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.1
    replicas: 3
```

## Method 3: Combined Approach

You can combine both methods for maximum flexibility:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
  namespace: starrocks
  annotations:
    starrocks.com/prepare-upgrade: "true"
    starrocks.com/upgrade-hooks: "disable-tablet-clone"
spec:
  upgradePreparation:
    enabled: true
    hooks:
    - name: custom-hook
      command: 'ADMIN SET FRONTEND CONFIG ("enable_strict_storage_medium_check" = "false")'
      critical: false
  # ... rest of spec
```

## Monitoring Upgrade Progress

### Check Upgrade Preparation Status

```bash
# Check the upgrade preparation status
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus}' | jq

# Monitor the phase
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.phase}'
```

### Watch Events

```bash
# Watch for upgrade-related events
kubectl get events --field-selector involvedObject.name=starrockscluster-sample -w

# Or use describe to see recent events
kubectl describe starrockscluster starrockscluster-sample
```

### Monitor Pod Changes

```bash
# Watch for pod changes during upgrade
kubectl get pods -l app.kubernetes.io/name=starrockscluster-sample -w
```

## Troubleshooting

### Manual Cleanup

If upgrade preparation gets stuck, you can manually clean up:

```bash
# Remove upgrade preparation annotations
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade- \
  starrocks.com/upgrade-hooks-

# Reset upgrade preparation status (requires direct status patch)
kubectl patch starrockscluster starrockscluster-sample \
  --subresource=status \
  --type='merge' \
  -p '{"status":{"upgradePreparationStatus":null}}'
```

### Check FE Connectivity

```bash
# Test connection to FE service
kubectl exec -it starrockscluster-sample-fe-0 -- mysql -h starrockscluster-sample-fe-service -P 9030 -u root -e "SELECT 1"
```

### View Hook Execution Logs

```bash
# Check operator logs for hook execution details
kubectl logs -n starrocks-operator-system deployment/starrocks-operator-controller-manager -f
```

## Post-Upgrade Cleanup

After successful upgrade, you may want to re-enable disabled features:

```bash
# Re-enable tablet cloning and balancing
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/upgrade-hooks=enable-tablet-clone,enable-balancer \
  starrocks.com/prepare-upgrade=true
```

## Best Practices

1. **Always test in staging first**: Test upgrade procedures in a non-production environment
2. **Monitor cluster health**: Check cluster status before and after upgrades
3. **Use critical flags wisely**: Mark essential hooks as critical to prevent upgrades if they fail
4. **Set appropriate timeouts**: Configure reasonable timeout values for hook execution
5. **Plan for rollback**: Have a rollback plan in case the upgrade fails

## Complete Upgrade Workflow Example

Here's a complete workflow for upgrading a StarRocks cluster:

```bash
# 1. Check current cluster status
kubectl get starrockscluster starrockscluster-sample

# 2. Apply upgrade preparation
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer

# 3. Wait for preparation to complete
kubectl wait --for=condition=ready starrockscluster starrockscluster-sample --timeout=300s

# 4. Upgrade images
kubectl patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{
    "spec": {
      "starRocksFeSpec": {"image": "starrocks/fe-ubuntu:3.3.0"},
      "starRocksBeSpec": {"image": "starrocks/be-ubuntu:3.3.0"}
    }
  }'

# 5. Monitor upgrade progress
kubectl get pods -l app.kubernetes.io/name=starrockscluster-sample -w

# 6. Verify upgrade completion
kubectl get starrockscluster starrockscluster-sample

# 7. Re-enable services (optional)
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/upgrade-hooks=enable-tablet-clone,enable-balancer \
  starrocks.com/prepare-upgrade=true
```
