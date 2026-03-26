# Disaster Recovery Cluster State Preservation Fix

## Problem Description

After disaster recovery completes (`DRPhaseDone`), when FE pods restart, they create new database paths with random UUIDs instead of using the original recovered paths. This causes 404 errors because the system looks for files in the new UUID-based path but data files remain in the original location.

### Root Cause

The issue occurs when `ShouldEnterDisasterRecoveryMode()` returns `false` (after disaster recovery completion). At this point:

1. The operator creates a "clean" StatefulSet without disaster recovery environment variables
2. When FE pods restart, they lack the `RESTORE_CLUSTER_SNAPSHOT=true` environment variable
3. FE initializes as a "new" cluster and generates a new UUID for cluster/database paths
4. This creates a mismatch where data exists in the original path but FE looks in the new UUID path

### Sequence of Events Leading to the Bug

```
1. All three FE pods start
2. All three FE pods terminate
3. FE-0 starts with RESTORE_CLUSTER_SNAPSHOT=true (recovery mode)
4. Other FEs and CNs start, recovery completes (DRPhaseDone)
5. FE-0 terminates
6. FE-0 starts without recovery env vars → Creates new UUID path! 🚫
```

## Solution Overview

The fix introduces **cluster state preservation** after disaster recovery completes. Instead of completely removing disaster recovery environment variables, the operator now:

1. **Extracts cluster UUID** from the recovery snapshot during DR completion
2. **Stores cluster state** in the `DisasterRecoveryStatus`
3. **Preserves cluster identity** through post-DR environment variables
4. **Maintains consistency** across pod restarts after recovery

### New Environment Variables for Post-DR State

- `RECOVERED_CLUSTER_UUID=<cluster-uuid>` - Preserves the recovered cluster identity
- `USE_RECOVERED_CLUSTER_STATE=true` - Signals FE to use recovered state

## Implementation Details

### 1. API Changes (`component_type.go`)

Extended `DisasterRecoveryStatus` with new fields:

```go
type DisasterRecoveryStatus struct {
    // ... existing fields ...

    // ClusterUUID stores the recovered cluster UUID to ensure consistency
    ClusterUUID string `json:"clusterUUID,omitempty"`

    // RecoveredClusterSnapshot stores the snapshot path used for recovery
    RecoveredClusterSnapshot string `json:"recoveredClusterSnapshot,omitempty"`
}
```

### 2. New Functions (`fe_disaster_recovery.go`)

#### `ShouldPreserveClusterState()`
Determines if cluster state should be preserved in StatefulSet:
- Returns `true` during active disaster recovery (existing behavior)
- Returns `true` after DR completion if cluster UUID is available
- Ensures consistent behavior across the DR lifecycle

#### `ExtractClusterUUIDFromSnapshot()`
Extracts cluster UUID from snapshot paths like:
```
s3://bucket/path/<cluster-uuid>/meta/image/snapshot_name
```

#### `extractClusterStateFromConfigMaps()`
Reads cluster_snapshot.yaml ConfigMap to extract cluster UUID during DR completion.

#### `RewriteStatefulSetForClusterStatePreservation()`
Modifies StatefulSet to preserve cluster state:
- **During DR**: Uses existing `RESTORE_CLUSTER_GENERATION` + `RESTORE_CLUSTER_SNAPSHOT`
- **After DR**: Uses new `RECOVERED_CLUSTER_UUID` + `USE_RECOVERED_CLUSTER_STATE`

### 3. Controller Logic Updates (`fe_controller.go`)

Enhanced the main sync logic to handle both active DR and post-DR scenarios:

```go
shouldEnterDRMode, queryPort := ShouldEnterDisasterRecoveryMode(drSpec, drStatus, feConfig)
shouldPreserveState, stateQueryPort := ShouldPreserveClusterState(drSpec, drStatus, feConfig)

if shouldEnterDRMode {
    // Active disaster recovery
    EnterDisasterRecoveryMode(ctx, fc.Client, src, &expectSts, queryPort)
} else if shouldPreserveState {
    // Post-DR cluster state preservation
    RewriteStatefulSetForClusterStatePreservation(&expectSts, drSpec, drStatus, stateQueryPort)
}
```

### 4. Cluster State Extraction

During DR completion (`DRPhaseDone` transition), the operator:

1. **Reads** the cluster_snapshot.yaml ConfigMap
2. **Parses** the `cluster_snapshot_path` to extract cluster UUID
3. **Stores** the UUID in `DisasterRecoveryStatus.ClusterUUID`
4. **Preserves** this information for future pod restarts

## Testing

### New Test Coverage

Added comprehensive tests for the new functionality:

- `TestShouldPreserveClusterState()` - Tests cluster state preservation logic
- `TestExtractClusterUUIDFromSnapshot()` - Tests UUID extraction from S3 paths
- `TestExtractClusterUUIDFromSnapshotYaml()` - Tests YAML parsing logic

### Backward Compatibility

- ✅ All existing disaster recovery tests pass
- ✅ Normal (non-DR) clusters are unaffected
- ✅ Active disaster recovery behavior unchanged
- ✅ New fields are optional with proper defaults

## Benefits

1. **Fixes Path Mapping Issue**: FE pods use consistent cluster UUIDs after recovery
2. **Maintains Data Integrity**: No more 404 errors from UUID mismatches
3. **Preserves Existing Behavior**: Active DR functionality unchanged
4. **Backward Compatible**: Works with existing DR configurations
5. **Minimal Impact**: Focused fix without architectural changes

## Usage

The fix works automatically with existing disaster recovery configurations. No changes required to:
- `cluster_snapshot.yaml` format
- Disaster recovery CRD specifications
- Existing recovery procedures

The operator will automatically extract and preserve cluster state information during recovery completion.

## Migration

For clusters that experienced this issue:
1. Apply the updated operator
2. The next disaster recovery operation will work correctly
3. No manual intervention required for the cluster state preservation

## Related Files Changed

- `pkg/apis/starrocks/v1/component_type.go` - Extended DR status structure
- `pkg/subcontrollers/fe/fe_disaster_recovery.go` - Added cluster state preservation logic
- `pkg/subcontrollers/fe/fe_controller.go` - Updated main controller sync logic
- `pkg/subcontrollers/fe/fe_disaster_recovery_test.go` - Added comprehensive test coverage