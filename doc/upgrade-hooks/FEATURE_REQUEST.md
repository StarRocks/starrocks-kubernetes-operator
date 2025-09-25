# Feature Request: Pre-Upgrade Hook Support for StarRocks Operator

## Summary
Add support for pre-upgrade hooks to automatically execute SQL commands (like ADMIN SET FRONTEND CONFIG) before image upgrades to handle tablet clone disabling and other preparatory tasks.

## Background
Currently, the StarRocks Kubernetes Operator only handles image upgrades by patching StatefulSets, but doesn't execute the recommended pre-upgrade steps documented in StarRocks upgrade procedures. Manual intervention is required to:

1. Disable tablet clone: `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0");`
2. Disable balancer: `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_balancing_tablets" = "0");`
3. Set balance flags: `ADMIN SET FRONTEND CONFIG ("disable_balance"="true");`
4. Set colocate balance: `ADMIN SET FRONTEND CONFIG ("disable_colocate_balance"="true");`

## Proposed Solution
Add annotation-based trigger mechanism for pre-upgrade hooks:

### Annotation Trigger
```yaml
metadata:
  annotations:
    starrocks.com/prepare-upgrade: "true"
    starrocks.com/upgrade-hooks: "disable-tablet-clone,disable-balancer"
```

### Implementation Areas

1. **API Enhancement** - Add new fields to StarRocksCluster CRD
2. **Controller Logic** - Detect annotation changes and execute SQL commands
3. **SQL Execution** - Connect to FE and run ADMIN SET commands
4. **State Management** - Track upgrade preparation status

## Benefits
- Automated upgrade preparation
- Reduced manual intervention
- Consistent upgrade procedures
- Better operational reliability

## Implementation Details
See attached code modifications for detailed implementation.

## Current Implementation Status

✅ **API Modifications**: Added new types to `pkg/apis/starrocks/v1/starrockscluster_types.go`:
- `UpgradeHook` type for defining upgrade hooks
- `UpgradePreparation` type for configuration
- `UpgradePreparationStatus` type for status tracking
- Added fields to `StarRocksClusterSpec` and `StarRocksClusterStatus`

✅ **Controller Implementation**: Created `pkg/subcontrollers/upgrade_hook_controller.go`:
- Hook execution logic with MySQL connection to FE
- Support for both annotation-based and spec-based configuration
- Predefined hooks for common upgrade scenarios
- Error handling and status tracking

✅ **Main Controller Integration**: Modified `pkg/controllers/starrockscluster_controller.go`:
- Added upgrade hook execution before subcontroller reconciliation
- Proper status updates and error handling
- Requeue logic for multi-step upgrade process

✅ **Test Coverage**: Created `pkg/subcontrollers/upgrade_hook_controller_test.go`:
- Unit tests for hook detection and execution logic
- Test cases for various configuration scenarios

## Usage Examples

### Method 1: Using Annotations
```bash
# Trigger upgrade preparation
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer

# Proceed with image upgrade
kubectl patch starrockscluster starrockscluster-sample \
  --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:latest"}}}'
```

### Method 2: Using Spec Configuration
```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
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
  # ... rest of cluster spec
```

## Monitoring

Check upgrade preparation status:
```bash
kubectl get starrockscluster starrockscluster-sample -o yaml | grep -A 20 upgradePreparationStatus
```

## Next Steps

1. **Testing**: Test the implementation in a development environment
2. **Documentation**: Update main README and API documentation
3. **CRD Generation**: Run `make manifests` to regenerate CRDs with new fields
4. **Deep Copy**: Run `make generate` to update deep copy methods
5. **Examples**: Add complete examples to `examples/` directory

Created: 2025-09-25
Updated: 2025-09-25 (Implementation completed)
