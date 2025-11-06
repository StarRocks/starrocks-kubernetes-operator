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
Add spec-based configuration for upgrade hooks with support for both predefined and custom scripts:

### Spec Configuration
```yaml
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      # Predefined hooks (safe, recommended)
      predefined:
        - disable-tablet-clone
        - disable-balancer

      # Custom hooks from ConfigMap (advanced)
      custom:
        configMapName: fe-upgrade-hooks
        scriptKey: hooks.sh
```

### Implementation Areas

1. **API Enhancement** - Add ComponentUpgradeHooks to component specs (FE, BE, CN)
2. **Controller Logic** - Detect image changes and execute hooks per component
3. **Hook Execution** - Support both SQL commands and shell scripts from ConfigMap
4. **State Management** - Track upgrade state per component (not cluster-wide)

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

### Example 1: Predefined Hooks (Recommended)
```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    replicas: 3
    upgradeHooks:
      predefined:
        - disable-tablet-clone
        - disable-balancer
      timeoutSeconds: 300

  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.2
    replicas: 3
    upgradeHooks:
      predefined:
        - disable-tablet-clone
```

### Example 2: Custom Hooks from ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fe-upgrade-hooks
  namespace: starrocks
data:
  hooks.sh: |
    #!/bin/bash
    pre_upgrade() {
        mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER} \
          -e 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")'
        return $?
    }
    post_upgrade() {
        mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER} \
          -e 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")'
        return $?
    }
---
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      custom:
        configMapName: fe-upgrade-hooks
        scriptKey: hooks.sh
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
