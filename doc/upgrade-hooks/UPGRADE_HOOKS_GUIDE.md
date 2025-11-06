# Upgrade Hooks Guide

## Overview

The StarRocks Kubernetes Operator supports **upgrade hooks** to execute custom logic before and after component upgrades. This allows you to:

- Prepare the cluster for upgrades (disable balancing, tablet scheduling, etc.)
- Perform health checks before proceeding
- Execute cleanup tasks after upgrades complete
- Implement custom upgrade workflows specific to your environment

## Feature Status

✅ **Available in**: v1.12.0+
⚠️ **Security Level**: HIGH - Requires administrator privileges
📚 **Maturity**: Beta

---

## Quick Start

### Option 1: Predefined Hooks (Recommended)

The safest way to use upgrade hooks is with predefined, reviewed commands:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: my-cluster
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      predefined:
        - disable-tablet-clone
        - disable-balancer
```

### Option 2: Custom Hooks (Advanced)

For complex scenarios, you can provide custom shell scripts:

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: my-cluster
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      custom:
        configMapName: fe-upgrade-hooks
        scriptKey: hooks.sh
```

See [SAFE_EXAMPLES.md](SAFE_EXAMPLES.md) for complete examples.

---

## How It Works

### Upgrade Flow

```
1. Operator detects image change (FE: 3.2.1 → 3.2.2)
   ↓
2. UpgradeState set to "Detected"
   ↓
3. Execute pre_upgrade() hook
   - Disable tablet scheduling
   - Disable balancing
   - Custom logic...
   ↓
4. UpgradeState set to "Ready"
   ↓
5. Operator updates StatefulSet with new image
   ↓
6. Pods roll out with new version
   ↓
7. Operator detects upgrade completion
   ↓
8. Execute post_upgrade() hook
   - Re-enable tablet scheduling
   - Re-enable balancing
   - Custom cleanup...
   ↓
9. UpgradeState set to "Completed"
```

### Component-Level Upgrades

Each component (FE, BE, CN) is upgraded **independently** with its own hooks:

```yaml
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      predefined: [disable-tablet-clone]

  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.2  # Different version OK!
    upgradeHooks:
      predefined: [disable-balancer]

  starRocksCnSpec:
    image: starrocks/cn-ubuntu:3.2.1  # Different version OK!
    upgradeHooks:
      custom:
        configMapName: cn-hooks
```

---

## Configuration Reference

### ComponentUpgradeHooks

```yaml
upgradeHooks:
  # List of predefined hook names (safe, recommended)
  predefined:
    - disable-tablet-clone
    - disable-balancer
    - enable-tablet-clone  # For post-upgrade
    - enable-balancer      # For post-upgrade

  # Custom script from ConfigMap (advanced, requires security review)
  custom:
    configMapName: my-hooks  # Required
    scriptKey: hooks.sh      # Optional, default: "hooks.sh"

  # Maximum time for hook execution (seconds)
  timeoutSeconds: 300  # Optional, default: 300, max: 3600
```

### Predefined Hooks

| Hook Name | Phase | SQL Command | Purpose |
|-----------|-------|-------------|---------|
| `disable-tablet-clone` | pre | `tablet_sched_max_scheduling_tablets = 0` | Stop tablet replication |
| `disable-balancer` | pre | `disable_balance = true` | Stop cluster balancing |
| `enable-tablet-clone` | post | `tablet_sched_max_scheduling_tablets = 2000` | Resume replication |
| `enable-balancer` | post | `disable_balance = false` | Resume balancing |

### Custom Hook Script Format

Your shell script must define these functions:

```bash
#!/bin/bash

# Called BEFORE upgrade starts
# Return 0 for success, non-zero for failure
pre_upgrade() {
    echo "INFO: Pre-upgrade logic here"
    # Your code...
    return 0
}

# Called AFTER upgrade completes
# Return 0 for success, non-zero for failure (non-critical)
post_upgrade() {
    echo "INFO: Post-upgrade logic here"
    # Your code...
    return 0
}
```

**Environment Variables Provided:**
- `SR_FE_HOST` - FE service hostname
- `SR_FE_PORT` - FE query port (default: 9030)
- `SR_FE_USER` - FE user (default: root)
- `SR_CLUSTER_NAME` - StarRocksCluster name
- `SR_NAMESPACE` - StarRocksCluster namespace

---

## Security Considerations

⚠️ **CRITICAL**: Custom hooks execute with operator privileges and can:
- Execute arbitrary SQL commands against StarRocks
- Run arbitrary shell commands in the operator container
- Access the operator's Kubernetes service account
- Read/write the operator's filesystem

### Threat Model

| Threat | Impact | Mitigation |
|--------|--------|------------|
| SQL Injection | Database compromise | ✅ Hardcode all SQL, never use variables |
| Command Injection | Container escape | ✅ RBAC protection, no kubectl access |
| Resource Exhaustion | DoS | ✅ Timeouts, resource limits |
| Privilege Escalation | Cluster compromise | ✅ Restricted service account |
| Credential Theft | Unauthorized access | ✅ Audit logging, monitoring |

### Security Requirements

1. **RBAC Protection**: Only cluster admins can create/modify ConfigMaps
2. **Code Review**: ALL custom scripts must be reviewed by security team
3. **Audit Logging**: All hook executions are logged with script hashes
4. **Timeouts**: Hooks are terminated if they exceed timeout
5. **No External Access**: Hooks should not call external services

### Recommended Controls

- [ ] Store hook ConfigMaps in protected namespace
- [ ] Implement admission webhooks to validate hook content
- [ ] Enable audit logging for ConfigMap changes
- [ ] Monitor operator logs for suspicious hook behavior
- [ ] Use Network Policies to restrict operator egress
- [ ] Implement Pod Security Standards (restricted profile)

For complete security analysis, see [SECURITY_ANALYSIS.md](SECURITY_ANALYSIS.md).

---

## Monitoring and Troubleshooting

### Check Upgrade Status

```bash
# View upgrade state for all components
kubectl get starrockscluster my-cluster -o yaml | grep -A 10 upgradeState

# View FE upgrade state
kubectl get starrockscluster my-cluster \
  -o jsonpath='{.status.starRocksFeStatus.upgradeState}'

# View BE upgrade state
kubectl get starrockscluster my-cluster \
  -o jsonpath='{.status.starRocksBeStatus.upgradeState}'
```

### Monitor Hook Execution

```bash
# Watch operator logs for hook execution
kubectl logs -f deployment/starrocks-operator \
  -n starrocks-operator-system | grep -E "(CustomHookExecutor|UpgradeManager)"

# Check for security audit logs
kubectl logs deployment/starrocks-operator \
  -n starrocks-operator-system | grep "SECURITY AUDIT"
```

### Common Issues

#### Issue: Hook Timeout

**Symptom**: `Hook execution failed: context deadline exceeded`

**Solution**:
```yaml
upgradeHooks:
  timeoutSeconds: 600  # Increase timeout
  custom:
    configMapName: my-hooks
```

#### Issue: ConfigMap Not Found

**Symptom**: `failed to get ConfigMap: configmaps "my-hooks" not found`

**Solution**:
```bash
# Verify ConfigMap exists in same namespace
kubectl get configmap my-hooks -n starrocks

# Check ConfigMap has correct key
kubectl get configmap my-hooks -n starrocks -o yaml
```

#### Issue: Script Function Not Found

**Symptom**: `script must define pre_upgrade() function`

**Solution**:
```bash
# Ensure script defines required functions
cat <<'EOF' | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-hooks
data:
  hooks.sh: |
    #!/bin/bash
    pre_upgrade() {
        # Your logic here
        return 0
    }
    post_upgrade() {
        # Your logic here
        return 0
    }
EOF
```

#### Issue: Hook Fails but Returns 0

**Symptom**: Hook executes but doesn't have desired effect

**Solution**: Add explicit error checking:
```bash
pre_upgrade() {
    $MYSQL_CMD -e 'SQL COMMAND' || {
        echo "ERROR: Command failed"
        return 1  # Must return non-zero on error!
    }
}
```

---

## Examples

### Example 1: Basic Upgrade with Predefined Hooks

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: basic-cluster
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    replicas: 3
    upgradeHooks:
      predefined:
        - disable-tablet-clone
        - disable-balancer
```

### Example 2: Multi-Component Upgrade

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: multi-cluster
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    upgradeHooks:
      predefined: [disable-tablet-clone, disable-balancer]

  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.2
    upgradeHooks:
      predefined: [disable-tablet-clone]

  starRocksCnSpec:
    image: starrocks/cn-ubuntu:3.2.2
    # CN upgrades typically don't need hooks
```

### Example 3: Custom Hook with Health Checks

See [SAFE_EXAMPLES.md](SAFE_EXAMPLES.md) for complete, production-ready examples.

---

## Best Practices

### DO ✅

1. **Use Predefined Hooks When Possible**
   - Safer, well-tested, no security review needed

2. **Test in Development First**
   - Always test hooks in non-production environment

3. **Keep Hooks Simple**
   - One responsibility per hook
   - Easy to understand and review

4. **Add Comprehensive Logging**
   - Log all actions for audit trail
   - Include timestamps and operation details

5. **Handle Errors Gracefully**
   - Check return codes
   - Fail fast on critical errors
   - Continue on non-critical errors

### DON'T ❌

1. **Never Use User Input in SQL**
   - SQL injection vulnerability

2. **Never Call External Services**
   - Data exfiltration risk

3. **Never Use kubectl in Hooks**
   - Privilege escalation risk

4. **Never Disable Error Handling**
   - Can mask failures

5. **Never Store Secrets in ConfigMaps**
   - Use Kubernetes Secrets instead

---

## Migration from Annotations

If you were using the annotation-based approach (now deprecated):

**Old (Deprecated):**
```yaml
metadata:
  annotations:
    starrocks.com/prepare-upgrade: "true"
    starrocks.com/upgrade-hooks: "disable-tablet-clone,disable-balancer"
```

**New (Current):**
```yaml
spec:
  starRocksFeSpec:
    upgradeHooks:
      predefined:
        - disable-tablet-clone
        - disable-balancer
```

---

## FAQ

### Q: Are hooks required for upgrades?

**A**: No. If no hooks are configured, the operator will execute default pre-upgrade hooks (disable tablet scheduling and balancing). You only need to configure hooks if you want custom behavior.

### Q: Can I use both predefined and custom hooks?

**A**: Yes! Both can be configured and will execute in order (predefined first, then custom).

### Q: What happens if a hook fails?

**A**: Pre-upgrade hooks are critical - if they fail, the upgrade is blocked. Post-upgrade hooks are non-critical - failures are logged but don't block the upgrade.

### Q: Can different components have different hooks?

**A**: Yes! Each component (FE, BE, CN) can have its own hooks configuration.

### Q: How do I debug hook execution?

**A**: Check operator logs for detailed execution traces:
```bash
kubectl logs deployment/starrocks-operator -n starrocks-operator-system --tail=100
```

### Q: Are hooks supported for rollbacks?

**A**: Not currently. Hooks only execute during forward upgrades (version increase detection).

---

## References

- [Security Analysis](SECURITY_ANALYSIS.md) - Comprehensive threat model and mitigations
- [Safe Examples](SAFE_EXAMPLES.md) - Production-ready hook examples
- [Feature Request](FEATURE_REQUEST.md) - Original feature proposal
- [Implementation Details](README.md) - Technical architecture

---

**Version**: 1.0
**Last Updated**: 2025-01-06
**Status**: Beta
