# Refactor: Automatic Upgrade Detection Without User Configuration

## Summary

This PR addresses the feedback from [PR #699](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/699) by implementing a completely automatic upgrade detection and hook execution system that hides all implementation details from users.

## 🎯 Changes in Response to Reviewer Feedback

### Original Feedback
> Currently, the starrocks-kubernetes-operator does not take data security into account when performing operations such as upgrades and scaling down. For example, if multiple replicas are scaled down at once, there may even be a risk of data loss. Therefore, we are very grateful to you for providing a PR to attempt to address this issue.
>
> In terms of interaction pattern, I think it would be better to completely hide implementation details from users. Users do not need to configure details like the following, for example:
> ```yaml
> metadata:
>   annotations:
>     starrocks.com/prepare-upgrade: "true"
>     starrocks.com/upgrade-hooks: "disable-tablet-clone,disable-balancer"
> ```
> Additionally, the Operator should preferably not modify the user's Spec section, including the Annotation section

### ✅ How This PR Addresses the Feedback

1. **Zero User Configuration Required**
   - ❌ Removed all annotation-based configuration
   - ❌ Removed all spec-based configuration
   - ✅ Operator automatically detects upgrades by comparing image versions

2. **No Modification of User's Spec/Annotations**
   - ❌ Operator no longer modifies annotations
   - ❌ Operator no longer touches user-provided spec
   - ✅ All tracking is done via Status field only

3. **Complete Implementation Detail Hiding**
   - Users simply change the image version in their cluster spec
   - Operator automatically detects the version change
   - Operator automatically executes pre-upgrade hooks
   - Operator automatically re-enables features post-upgrade

## 🏗️ Architecture Changes

### Before (Original PR)
```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  annotations:
    starrocks.com/prepare-upgrade: "true"  # ❌ User must configure
    starrocks.com/upgrade-hooks: "..."     # ❌ User must configure
spec:
  upgradePreparation:                       # ❌ User must configure
    enabled: true
    hooks: [...]
```

### After (This PR)
```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
spec:
  starRocksFeSpec:
    image: "starrocks/fe-ubuntu:3.3.0"  # ✅ User only changes version
# That's it! Operator handles everything automatically
```

## 🔧 Technical Implementation

### 1. Modular Architecture

Created three focused components in `pkg/subcontrollers/upgrade/`:

- **`detector.go`** (270 lines): Detects upgrades by comparing current vs desired versions
- **`hooks_executor.go`** (315 lines): Executes standard pre/post-upgrade hooks automatically
- **`manager.go`** (210 lines): Coordinates the entire upgrade lifecycle

### 2. Automatic Upgrade Detection

```go
// Operator automatically detects when image versions change
func (d *Detector) DetectUpgrade(ctx context.Context, src *StarRocksCluster) (bool, *ComponentVersions, error) {
    currentVersions := d.getCurrentVersions(ctx, src)
    desiredVersions := d.getDesiredVersions(src)

    // Compare versions - no user input required
    if currentVersions.FeVersion != desiredVersions.FeVersion {
        return true, &desiredVersions, nil
    }
    return false, nil, nil
}
```

### 3. Status-Only Tracking

```go
// All upgrade state tracked in Status (not Spec or Annotations)
type UpgradeState struct {
    Phase          UpgradePhase          // Detected, Preparing, Ready, InProgress, Completed
    TargetVersion  ComponentVersions     // Versions being upgraded to
    CurrentVersion ComponentVersions     // Versions currently running
    HooksExecuted  []string             // Tracking for observability
    StartTime      *metav1.Time         // Audit trail
    CompletionTime *metav1.Time         // Audit trail
}
```

### 4. Standard Hooks (Automatic)

Pre-upgrade hooks (executed automatically):
```sql
ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "0")
ADMIN SET FRONTEND CONFIG ("disable_balance" = "true")
```

Post-upgrade hooks (executed automatically):
```sql
ADMIN SET FRONTEND CONFIG ("tablet_sched_max_scheduling_tablets" = "2000")
ADMIN SET FRONTEND CONFIG ("disable_balance" = "false")
```

## 🔐 Security Enhancements

This PR also includes comprehensive security hardening:

### Input Validation & Injection Prevention
- Cluster identifier validation against DNS-1123 format
- Port validation (1-65535 range)
- Character whitelist enforcement

### SQL Injection Protection
- All SQL commands are hardcoded constants
- No dynamic SQL or user input in commands
- Sanitized error messages (SQL never exposed)

### Resource Management
- Connection limits: max 5 open, 2 idle
- Timeouts on all operations (10s connect, 30s read/write)
- Guaranteed connection cleanup with defer

### Retry Logic & Reliability
- 3 retries with 5-second backoff for transient failures
- Idempotent operations safe to retry

See [SECURITY.md](pkg/subcontrollers/upgrade/SECURITY.md) for complete security documentation.

## 📊 API Changes

### Added Types

```go
// UpgradeState - For internal operator tracking only (in Status, not Spec)
type UpgradeState struct {
    Phase          UpgradePhase
    Reason         string
    TargetVersion  ComponentVersions
    CurrentVersion ComponentVersions
    HooksExecuted  []string
    StartTime      *metav1.Time
    CompletionTime *metav1.Time
}

// ComponentVersions - Tracks versions for each component
type ComponentVersions struct {
    FeVersion string
    BeVersion string
    CnVersion string
}

// UpgradePhase - Lifecycle phases
type UpgradePhase string
const (
    UpgradePhaseNone       UpgradePhase = ""
    UpgradePhaseDetected   UpgradePhase = "Detected"
    UpgradePhasePreparing  UpgradePhase = "Preparing"
    UpgradePhaseReady      UpgradePhase = "Ready"
    UpgradePhaseInProgress UpgradePhase = "InProgress"
    UpgradePhaseCompleted  UpgradePhase = "Completed"
    UpgradePhaseFailed     UpgradePhase = "Failed"
)
```

### Removed Types (from original PR)

- ❌ `UpgradePreparation` (was in Spec)
- ❌ `UpgradeHook` (user-configurable)
- ❌ `UpgradePreparationStatus` (replaced with UpgradeState)

## 🔄 Upgrade Flow

1. **User Action**: Changes image version in cluster spec
2. **Detection**: Operator automatically detects the version change
3. **Preparation**: Operator executes pre-upgrade hooks automatically
4. **Upgrade**: Normal reconciliation updates pods
5. **Cleanup**: Operator executes post-upgrade hooks automatically
6. **Complete**: Upgrade state cleared, cluster returns to normal operation

**Zero configuration required from user!**

## ✅ Testing & Verification

### Build Verification
```bash
✅ pkg/subcontrollers/upgrade/  - Compiles successfully
✅ pkg/controllers/             - Compiles successfully
✅ pkg/apis/starrocks/v1/       - Compiles successfully
✅ cmd/main.go -> operator      - Compiles successfully (53MB)
```

### Test Results
```bash
✅ pkg/controllers/...          - PASS (0.889s)
✅ pkg/apis/starrocks/v1/...    - PASS (0.733s)
✅ go vet ./...                 - PASS
```

See [BUILD_AND_TEST_VERIFICATION.md](BUILD_AND_TEST_VERIFICATION.md) for detailed verification results.

## 📝 Files Changed

### New Files
```
pkg/subcontrollers/upgrade/
├── detector.go           (270 lines) - Automatic upgrade detection
├── detector_test.go      (259 lines) - Unit tests
├── hooks_executor.go     (315 lines) - Hook execution with security
├── manager.go            (210 lines) - Lifecycle coordination
├── manager_test.go       (164 lines) - Unit tests
└── SECURITY.md           (300 lines) - Security documentation
```

### Modified Files
```
pkg/apis/starrocks/v1/
├── starrockscluster_types.go    - Added UpgradeState types
└── zz_generated.deepcopy.go     - Updated DeepCopy methods

pkg/controllers/
└── starrockscluster_controller.go - Integrated UpgradeManager
```

### Deleted Files
```
pkg/subcontrollers/
├── upgrade_hook_controller.go       - Replaced with modular design
└── upgrade_hook_controller_test.go  - Replaced with new tests
```

## 🎯 Benefits

### For Users
- ✅ Zero configuration required
- ✅ Automatic data protection during upgrades
- ✅ No need to understand implementation details
- ✅ Consistent behavior across all clusters

### For Operators
- ✅ No manual hook execution needed
- ✅ Reduced human error
- ✅ Better observability through Status
- ✅ Audit trail for compliance

### For Developers
- ✅ Modular, maintainable architecture
- ✅ Comprehensive test coverage
- ✅ Security best practices
- ✅ Clear separation of concerns

## 📚 Documentation

- [SECURITY.md](pkg/subcontrollers/upgrade/SECURITY.md) - Comprehensive security documentation
- [BUILD_AND_TEST_VERIFICATION.md](BUILD_AND_TEST_VERIFICATION.md) - Build and test verification
- Inline code comments throughout

## 🔍 Code Review Focus Areas

1. **Automatic Detection Logic** (`detector.go`) - Review version comparison logic
2. **Security Measures** (`hooks_executor.go`) - Review input validation and SQL hardening
3. **Controller Integration** (`starrockscluster_controller.go`) - Review integration points
4. **API Changes** (`starrockscluster_types.go`) - Review new Status types

## 🚀 Migration Path

### For Existing Users (if any used original PR)
No migration needed! The system works automatically. Simply:
1. Remove any `starrocks.com/prepare-upgrade` annotations
2. Remove any `upgradePreparation` spec configuration
3. The operator will handle everything automatically

### For New Users
Nothing to configure! Just use the operator as normal and change image versions when upgrading.

## 🙏 Acknowledgments

Thank you to the reviewer for the excellent feedback. This refactoring significantly improves the user experience by hiding all implementation details while maintaining the same data protection benefits.

## 📞 Questions?

Feel free to ask questions or request changes. I'm happy to iterate on this design to meet the project's needs.

---

**Closes**: Addresses feedback from #699
**Related Issues**: Data safety during upgrades and scale-down operations