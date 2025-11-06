# StarRocks Operator Upgrade Hooks Implementation

This document provides a technical overview of the upgrade hooks implementation in the StarRocks Kubernetes Operator.

## Architecture Overview

The upgrade hooks feature extends the StarRocks Operator to support automated pre-upgrade preparation by executing SQL commands before image updates. The implementation follows Kubernetes operator patterns and integrates seamlessly with the existing controller architecture.

### Components

1. **API Extensions** (`pkg/apis/starrocks/v1/starrockscluster_types.go`)
   - New CRD fields for upgrade preparation configuration
   - Status tracking for upgrade preparation progress

2. **Upgrade Hook Controller** (`pkg/subcontrollers/upgrade_hook_controller.go`)
   - Manages hook execution lifecycle
   - Connects to FE via MySQL protocol
   - Handles error scenarios and retries

3. **Main Controller Integration** (`pkg/controllers/starrockscluster_controller.go`)
   - Orchestrates upgrade preparation before component reconciliation
   - Manages status updates and requeue logic

## Implementation Details

### API Types

```go
// UpgradeHook defines a pre-upgrade command
type UpgradeHook struct {
    Name     string `json:"name"`
    Command  string `json:"command"`
    Critical bool   `json:"critical,omitempty"`
}

// UpgradePreparation configuration
type UpgradePreparation struct {
    Enabled        bool          `json:"enabled,omitempty"`
    Hooks          []UpgradeHook `json:"hooks,omitempty"`
    TimeoutSeconds int32         `json:"timeoutSeconds,omitempty"`
}
```

### Controller Logic Flow

1. **Detection**: Check for upgrade preparation triggers (annotations or spec)
2. **Validation**: Verify cluster readiness and hook configuration
3. **Execution**: Connect to FE and execute SQL commands sequentially
4. **Status Update**: Track progress and completion status
5. **Integration**: Allow main reconciliation to proceed when ready

### Database Connection

The controller establishes MySQL connections to the FE service:

```go
// Connection string format
dsn := fmt.Sprintf("root@tcp(%s:%s)/", host, port)

// Default FE service discovery
host := fmt.Sprintf("%s-fe-service.%s.svc.cluster.local", clusterName, namespace)
port := "9030" // Default query port
```

### Error Handling

- **Non-critical hooks**: Continue execution if they fail
- **Critical hooks**: Stop upgrade process and report failure
- **Connection errors**: Retry with backoff
- **Timeout handling**: Configurable per-hook timeout

### Status Tracking

The implementation tracks upgrade preparation through distinct phases:

- `Pending`: Preparation not started
- `Running`: Hooks are being executed
- `Completed`: All hooks executed successfully
- `Failed`: One or more critical hooks failed

## Configuration Method

Upgrade hooks are configured in the component spec using the `upgradeHooks` field:

```yaml
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      # Option 1: Predefined hooks (recommended, safe)
      predefined:
        - disable-tablet-clone
        - disable-balancer

      # Option 2: Custom hooks from ConfigMap (advanced)
      custom:
        configMapName: fe-upgrade-hooks
        scriptKey: hooks.sh

      timeoutSeconds: 300
```

## Predefined Hooks

The implementation includes common upgrade preparation hooks:

| Hook Name | SQL Command | Purpose |
|-----------|-------------|---------|
| `disable-tablet-clone` | `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")` | Stop tablet replication |
| `disable-balancer` | `ADMIN SET FRONTEND CONFIG ("disable_balance"="true")` | Disable cluster balancing |
| `enable-tablet-clone` | `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")` | Re-enable tablet replication |
| `enable-balancer` | `ADMIN SET FRONTEND CONFIG ("disable_balance"="false")` | Re-enable cluster balancing |

## Security Considerations

1. **Authentication**: Uses root user for FE connections (cluster admin context)
2. **Network**: Connections are cluster-internal via service discovery
3. **Validation**: SQL commands are validated before execution
4. **Permissions**: Requires appropriate RBAC for StarRocksCluster resource updates

## Integration Points

### Main Controller Integration
```go
// Pre-reconciliation hook execution
upgradeHookController := subcontrollers.NewUpgradeHookController(r.Client)
if upgradeHookController.ShouldExecuteUpgradeHooks(ctx, src) {
    if err = upgradeHookController.ExecuteUpgradeHooks(ctx, src); err != nil {
        return requeueIfError(err)
    }
}
```

### Requeue Logic
```go
// Requeue after successful hook execution
if src.Status.UpgradePreparationStatus.Phase == srapi.UpgradePreparationCompleted {
    return ctrl.Result{Requeue: true}, nil
}
```

## Testing Strategy

### Unit Tests
- Hook detection logic
- Predefined hook parsing
- Status management
- Error handling scenarios

### Integration Tests
- End-to-end upgrade scenarios
- Database connection handling
- Multi-hook execution
- Failure recovery

## Deployment Considerations

### Prerequisites
- MySQL driver dependency (already present in operator)
- FE service accessibility
- Appropriate RBAC permissions

### Monitoring
- Status field monitoring via kubectl
- Event generation for hook execution
- Error reporting through status conditions

## Future Enhancements

### Potential Improvements
1. **Custom Hook Validation**: Syntax checking for SQL commands
2. **Rollback Hooks**: Post-upgrade cleanup automation
3. **Conditional Hooks**: Execute hooks based on cluster state
4. **External Hooks**: Support for webhook-based preparations
5. **Batch Operations**: Parallel hook execution for performance

### Observability Enhancements
1. **Metrics**: Prometheus metrics for hook execution
2. **Tracing**: Distributed tracing for upgrade workflows
3. **Alerting**: Integration with monitoring systems

## Migration Guide

For existing deployments:

1. **API Changes**: New CRD fields are optional and backward-compatible
2. **Behavior**: No changes to existing upgrade behavior without explicit configuration
3. **Rollback**: Feature can be disabled by removing annotations/spec configuration

## Troubleshooting

### Common Issues
1. **Connection Failures**: Check FE service accessibility
2. **SQL Errors**: Verify command syntax and FE state
3. **Permission Errors**: Ensure proper RBAC configuration
4. **Timeout Issues**: Adjust timeout configuration for large clusters

### Debug Commands
```bash
# Check upgrade preparation status
kubectl get starrockscluster <name> -o jsonpath='{.status.upgradePreparationStatus}'

# Monitor operator logs
kubectl logs -f deployment/starrocks-operator -n starrocks-operator-system

# Test FE connectivity
kubectl exec -it <fe-pod> -- mysql -h127.0.0.1 -P9030 -uroot
```
