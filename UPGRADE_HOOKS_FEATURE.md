# StarRocks Kubernetes Operator - Upgrade Hooks Feature

This document describes the new **Upgrade Hooks** feature that has been added to the StarRocks Kubernetes Operator. This feature enables automated execution of SQL commands before StarRocks cluster upgrades, eliminating the need for manual intervention during upgrade processes.

## Overview

The StarRocks Operator now supports **Level 3 Operator** capabilities with automated upgrade preparation through:

- **Pre-upgrade SQL hooks**: Execute ADMIN SET commands before image upgrades
- **Annotation-based triggers**: Simple annotation-driven upgrade preparation
- **Spec-based configuration**: Declarative hook configuration in the StarRocksCluster spec
- **Status tracking**: Monitor upgrade preparation progress through status fields
- **Automatic cleanup**: Post-upgrade cleanup of preparation state

## Why This Feature is Important

StarRocks requires specific preparation steps before safely upgrading:

1. **Disable tablet scheduling**: Prevent tablet movement during upgrade
2. **Disable cluster balancing**: Stop rebalancing operations
3. **Set appropriate timeouts**: Extend timeouts for upgrade operations
4. **Custom preparations**: Execute cluster-specific preparation commands

Previously, these steps required manual execution via `mysql` client before each upgrade. This feature automates the entire process.

## Architecture Changes

### API Extensions

New types added to `pkg/apis/starrocks/v1/starrockscluster_types.go`:

```go
// UpgradeHook defines pre-upgrade commands to execute
type UpgradeHook struct {
    Name     string `json:"name"`
    Command  string `json:"command"`
    Critical bool   `json:"critical,omitempty"`
}

// UpgradePreparation defines upgrade preparation configuration
type UpgradePreparation struct {
    Enabled        bool          `json:"enabled,omitempty"`
    Hooks          []UpgradeHook `json:"hooks,omitempty"`
    TimeoutSeconds int32         `json:"timeoutSeconds,omitempty"`
}

// UpgradePreparationStatus tracks upgrade preparation state
type UpgradePreparationStatus struct {
    Phase          UpgradePreparationPhase `json:"phase,omitempty"`
    Reason         string                  `json:"reason,omitempty"`
    CompletedHooks []string               `json:"completedHooks,omitempty"`
    FailedHooks    []string               `json:"failedHooks,omitempty"`
    StartTime      *metav1.Time           `json:"startTime,omitempty"`
    CompletionTime *metav1.Time           `json:"completionTime,omitempty"`
}
```

### New Controller

- **UpgradeHookController** (`pkg/subcontrollers/upgrade_hook_controller.go`): Handles hook execution
- **MySQL Integration**: Uses go-sql-driver/mysql for FE connectivity
- **Predefined Hooks**: Common upgrade preparation patterns
- **Error Handling**: Robust error handling and status reporting

### Integration Points

- **StarRocksClusterReconciler**: Integrated hook execution before subcontroller sync
- **Automatic Detection**: Detects upgrade triggers and executes hooks accordingly
- **Status Management**: Updates cluster status with preparation progress

## Usage Methods

### Method 1: Annotations (Quick & Simple)

```bash
# Apply upgrade preparation annotations
kubectl annotate starrockscluster starrockscluster-sample \
  starrocks.com/prepare-upgrade=true \
  starrocks.com/upgrade-hooks=disable-tablet-clone,disable-balancer

# Perform the upgrade
kubectl patch starrockscluster starrockscluster-sample --type='merge' \
  -p '{"spec":{"starRocksFeSpec":{"image":"starrocks/fe-ubuntu:3.3.0"}}}'
```

### Method 2: Spec Configuration (Declarative)

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
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.3.0
  # ... rest of spec
```

## Predefined Hooks

The operator includes these predefined hooks:

| Hook Name | Command | Purpose |
|-----------|---------|---------|
| `disable-tablet-clone` | `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")` | Disable tablet scheduling |
| `disable-balancer` | `ADMIN SET FRONTEND CONFIG ("disable_balance"="true")` | Disable cluster balancing |
| `enable-tablet-clone` | `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")` | Re-enable tablet scheduling |
| `enable-balancer` | `ADMIN SET FRONTEND CONFIG ("disable_balance"="false")` | Re-enable cluster balancing |

## Status Monitoring

Monitor upgrade preparation progress:

```bash
# Check preparation status
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus}'

# Watch for completion
kubectl get starrockscluster starrockscluster-sample -o jsonpath='{.status.upgradePreparationStatus.phase}'
```

Status phases:
- `Pending`: Preparation is queued
- `Running`: Hooks are being executed
- `Completed`: All hooks completed successfully
- `Failed`: One or more critical hooks failed

## Implementation Files

The feature implementation includes:

```
pkg/
├── apis/starrocks/v1/
│   ├── starrockscluster_types.go        # API extensions
│   └── zz_generated.deepcopy.go         # Generated DeepCopy methods
├── controllers/
│   └── starrockscluster_controller.go   # Main controller integration
├── subcontrollers/
│   ├── upgrade_hook_controller.go       # Hook execution logic
│   └── upgrade_hook_controller_test.go  # Unit tests
doc/
└── upgrade_hooks_howto.md               # Detailed usage guide
examples/starrocks/
└── starrockscluster-with-upgrade-hooks.yaml  # Complete example
```

## Testing

Run the included tests:

```bash
cd pkg/subcontrollers
go test -v -run TestUpgradeHookController
```

## Dependencies

The feature uses existing dependencies:

- `github.com/go-sql-driver/mysql` (already in go.mod)
- Standard Kubernetes controller-runtime libraries
- No additional external dependencies required

## Backwards Compatibility

The feature is fully backwards compatible:

- Existing StarRocksCluster resources continue to work unchanged
- No upgrade hooks are executed unless explicitly configured
- API changes are additive with optional fields
- No breaking changes to existing functionality

## Future Enhancements

Potential future improvements:

1. **Hook Templates**: Library of common upgrade preparation patterns
2. **Rollback Hooks**: Automatic rollback on upgrade failure  
3. **Multi-Stage Hooks**: Pre/during/post upgrade hook stages
4. **Custom Validation**: Validate cluster state before proceeding
5. **Webhook Integration**: External webhook support for complex preparations

## Contributing

To contribute to this feature:

1. Review the implementation in `pkg/subcontrollers/upgrade_hook_controller.go`
2. Add tests for new functionality in `upgrade_hook_controller_test.go`
3. Update documentation in `doc/upgrade_hooks_howto.md`
4. Test with real StarRocks clusters

## Migration Guide

For existing deployments:

1. **No immediate action required**: Feature is opt-in
2. **Gradual adoption**: Start with annotation-based hooks for testing
3. **Move to spec-based**: Migrate to declarative configuration over time
4. **Monitor and validate**: Ensure hooks execute correctly in your environment

This feature significantly improves the StarRocks Operator's automation capabilities, moving it from a Level 2 to Level 3 operator with sophisticated lifecycle management capabilities.
