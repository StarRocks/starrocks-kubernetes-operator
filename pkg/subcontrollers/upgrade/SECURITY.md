# Security Considerations for Upgrade Hook System

## Overview

The upgrade hook system executes SQL commands on the StarRocks FE (Frontend) to prepare the cluster for safe upgrades. This document outlines the security measures implemented to protect against potential vulnerabilities.

## Security Measures Implemented

### 1. **Input Validation and Sanitization**

#### Cluster Identifier Validation
- **What**: Validates cluster name and namespace before constructing connection strings
- **Why**: Prevents DNS injection and ensures only valid Kubernetes names are used
- **Where**: `validateClusterIdentifiers()` in `hooks_executor.go`
- **Details**:
  - Cluster names must follow DNS-1123 subdomain format
  - Maximum length: 253 characters for names, 63 for namespaces
  - Only alphanumeric characters, hyphens, and dots allowed
  - Rejects any suspicious characters that could indicate injection attempts

#### Port Validation
- **What**: Validates port numbers are within valid range (1-65535)
- **Why**: Prevents invalid port configurations
- **Where**: `GetFEConnection()` in `hooks_executor.go`

### 2. **Network Security**

#### In-Cluster Only Connections
- **What**: Connections are restricted to `.svc.cluster.local` services
- **Why**: Ensures the operator only connects to services within the Kubernetes cluster
- **Where**: DSN construction in `GetFEConnection()`
- **Details**:
  ```go
  host := fmt.Sprintf("%s.%s.svc.cluster.local", feServiceName, src.Namespace)
  ```

#### Connection Timeouts
- **What**: All database operations have strict timeouts
- **Why**: Prevents resource exhaustion from hanging connections
- **Timeouts**:
  - Connection: 10 seconds
  - Read: 30 seconds
  - Write: 30 seconds
  - Hook execution: 300 seconds (configurable)

### 3. **SQL Injection Prevention**

#### Hardcoded Commands
- **What**: All SQL commands are hardcoded in the source code
- **Why**: No user input is incorporated into SQL commands
- **Where**: `getStandardPreUpgradeHooks()` and `getStandardPostUpgradeHooks()`
- **Commands**:
  ```go
  // Pre-upgrade
  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")`
  `ADMIN SET FRONTEND CONFIG ("disable_balance" = "true")`

  // Post-upgrade
  `ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")`
  `ADMIN SET FRONTEND CONFIG ("disable_balance" = "false")`
  ```

#### No Dynamic SQL
- **What**: No string concatenation or interpolation of user-controlled values
- **Why**: Eliminates SQL injection vectors entirely
- **Note**: Commands are constants defined in code, reviewed during code review

### 4. **Resource Management**

#### Connection Pooling Limits
- **What**: Conservative limits on database connections
- **Why**: Prevents resource exhaustion attacks
- **Limits**:
  - Max open connections: 5
  - Max idle connections: 2
  - Connection max lifetime: 5 minutes

#### Automatic Resource Cleanup
- **What**: Database connections are always closed, even on error
- **Why**: Prevents connection leaks
- **Implementation**: `defer` statements with error handling
  ```go
  defer func() {
      if closeErr := db.Close(); closeErr != nil {
          logger.Error(closeErr, "Failed to close database connection")
      }
  }()
  ```

### 5. **Error Handling and Information Disclosure**

#### Sanitized Error Messages
- **What**: Error messages don't expose SQL commands or sensitive details
- **Why**: Prevents information leakage that could aid attackers
- **Example**:
  ```go
  // Bad: return fmt.Errorf("failed to execute: %s", hook.Command)
  // Good: return fmt.Errorf("failed to execute hook: %w", err)
  ```

#### Limited Logging
- **What**: SQL commands are NOT logged in production
- **Why**: Prevents sensitive command details from appearing in logs
- **Implementation**: Only hook names are logged, not command content

### 6. **Retry Logic with Backoff**

#### Configurable Retries
- **What**: Hooks can retry on transient failures
- **Why**: Improves reliability without compromising security
- **Configuration**:
  - Max retries: 3
  - Retry delay: 5 seconds
  - Prevents rapid retry loops

### 7. **Least Privilege Principle**

#### Database User
- **What**: Uses `root` user for admin commands
- **Why**: Required for ADMIN SET FRONTEND CONFIG commands
- **Note**: StarRocks RBAC should be configured to restrict root access

#### Read-Only Validation
- **What**: Validation query (`SELECT 1`) has no side effects
- **Why**: Safe to execute without modifying cluster state
- **Where**: `ValidateFEReady()`

## Threat Model

### Threats Mitigated

1. **SQL Injection**: ✅ Eliminated by using hardcoded commands only
2. **DNS Injection**: ✅ Prevented by cluster identifier validation
3. **Connection String Injection**: ✅ Prevented by strict validation
4. **Resource Exhaustion**: ✅ Mitigated by connection limits and timeouts
5. **Information Disclosure**: ✅ Prevented by sanitized errors and limited logging
6. **Man-in-the-Middle**: ⚠️ Relies on Kubernetes network policies (in-cluster traffic)

### Threats Not Addressed

1. **StarRocks FE Compromise**: If the FE itself is compromised, hooks will execute on compromised system
2. **Kubernetes API Server Compromise**: If API server is compromised, attacker can modify cluster objects
3. **Network Sniffing**: Connection uses MySQL protocol without TLS (relies on cluster network security)

## Security Best Practices for Operators

### For Cluster Administrators

1. **Network Policies**: Implement Kubernetes NetworkPolicies to restrict traffic
2. **RBAC**: Configure StarRocks RBAC to limit root user capabilities
3. **Monitoring**: Monitor operator logs for unusual patterns
4. **Audit**: Enable audit logging for StarRocks configuration changes

### For Operator Developers

1. **Code Review**: All changes to hook commands must be reviewed
2. **Static Analysis**: Run security scanners on code
3. **Testing**: Test with malicious inputs (fuzzing)
4. **Dependencies**: Keep dependencies updated for security patches

## Compliance Considerations

### Audit Trail
- All hook executions are logged with:
  - Timestamp
  - Hook name
  - Success/failure status
  - Error details (sanitized)

### Access Control
- Operator requires RBAC permissions:
  - Get/List/Watch StarRocksClusters
  - Update StarRocksCluster status
  - Network access to FE services

## Security Testing

### Recommended Tests

1. **Input Validation Testing**
   ```go
   // Test with invalid cluster names
   cluster.Name = "../../../etc/passwd"
   cluster.Name = "test; DROP TABLE users;"
   ```

2. **Connection Timeout Testing**
   - Simulate slow network
   - Verify timeouts work as expected

3. **Resource Limit Testing**
   - Create many connections
   - Verify limits are enforced

4. **Error Handling Testing**
   - Simulate database errors
   - Verify no sensitive info in errors

## Incident Response

### If SQL Injection is Suspected

1. Review recent code changes to hook commands
2. Check operator logs for unusual SQL patterns
3. Verify cluster names in all StarRocksCluster objects
4. Audit StarRocks FE logs

### If Connection Leaks are Detected

1. Check operator metrics for connection pool usage
2. Review defer statements in code
3. Monitor database connection counts
4. Restart operator if necessary

## Contact

For security issues or questions:
- File an issue: https://github.com/StarRocks/starrocks-kubernetes-operator/issues
- Email: security@starrocks.com

## Version

- Document Version: 1.0
- Last Updated: 2025-09-30
- Applies to: Operator v1.11.2+