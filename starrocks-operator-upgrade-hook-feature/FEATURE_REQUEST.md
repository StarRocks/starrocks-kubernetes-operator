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

Created: 2025-09-25 08:40:59
