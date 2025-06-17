# PVC Expansion How-To Guide

This guide explains how to expand Persistent Volume Claims (PVCs) for StarRocks clusters using the Kubernetes operator.

## Overview

The StarRocks Kubernetes operator now supports expanding storage volumes for FE, BE, and CN components. This feature allows you to increase storage capacity without data loss or significant downtime.

## Prerequisites

1. **Storage Class Support**: Your storage class must support volume expansion
2. **Kubernetes Version**: Kubernetes 1.11+ (for PVC expansion support)
3. **Operator Version**: StarRocks operator v1.9.9+ (with PVC expansion feature)

### Checking Storage Class Support

Before attempting PVC expansion, verify that your storage class supports it:

```bash
kubectl get storageclass <storage-class-name> -o yaml
```

Look for `allowVolumeExpansion: true` in the output. If this field is missing or set to `false`, you'll need to use a different storage class or enable expansion on your current one.

**Important**: The StarRocks operator will automatically validate storage class expansion support before attempting any PVC expansion. If your storage class doesn't support expansion, the operator will reject the expansion request with a clear error message.

Example of an expansion-enabled storage class:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
allowVolumeExpansion: true  # This enables PVC expansion
```

## Supported Storage Types

The following storage volumes can be expanded:

### FE (Frontend) Component
- **Metadata storage**: `storageSize` in `storageSpec`
- **Log storage**: `logStorageSize` in `storageSpec`

### BE (Backend) Component  
- **Data storage**: `storageSize` in `storageSpec`
- **Log storage**: `logStorageSize` in `storageSpec`
- **Spill storage**: `spillStorageSize` in `storageSpec`

### CN (Compute Node) Component
- **Data storage**: `storageSize` in `storageSpec`
- **Log storage**: `logStorageSize` in `storageSpec`
- **Spill storage**: `spillStorageSize` in `storageSpec`

## How to Expand Storage

### Method 1: Using kubectl with StarRocksCluster CRD

1. **Get the current cluster configuration**:
   ```bash
   kubectl get starrockscluster <cluster-name> -n <namespace> -o yaml > cluster.yaml
   ```

2. **Edit the storage sizes** in `cluster.yaml`:
   ```yaml
   spec:
     starRocksFeSpec:
       storageVolumes:
       - name: fe-meta
         storageSize: 20Gi  # Increased from 10Gi
         mountPath: /opt/starrocks/fe/meta
     starRocksBeSpec:
       storageVolumes:
       - name: be-data
         storageSize: 200Gi  # Increased from 100Gi
         mountPath: /opt/starrocks/be/storage
   ```

3. **Apply the changes**:
   ```bash
   kubectl apply -f cluster.yaml
   ```

### Method 2: Using Helm

1. **Update your values file** (e.g., `values.yaml`):
   ```yaml
   starrocks:
     starrocksFESpec:
       storageSpec:
         storageSize: 20Gi      # Increased from 10Gi
         logStorageSize: 10Gi   # Increased from 5Gi
     starrocksBeSpec:
       storageSpec:
         storageSize: 200Gi     # Increased from 100Gi
         logStorageSize: 20Gi   # Increased from 10Gi
   ```

2. **Upgrade the Helm release**:
   ```bash
   helm upgrade <release-name> kube-starrocks/starrocks -f values.yaml
   ```

## Expansion Process

When you update storage sizes, the operator performs the following steps:

1. **Detection**: Identifies which PVCs need expansion
2. **Storage Class Validation**: Verifies that storage classes support volume expansion (`allowVolumeExpansion: true`)
3. **Detachment Check**: Determines if the storage class requires PVC detachment for expansion
4. **Size Validation**: Ensures no size reductions are attempted
5. **PVC Expansion**: Patches existing PVCs with new sizes (online or offline based on storage class)
6. **StatefulSet Management**: Recreates StatefulSets if needed
7. **Pod Restart**: Allows pods to restart and mount expanded volumes

## Expansion Methods

The operator supports two expansion methods based on storage class capabilities:

### Online Expansion (Preferred)
- **When**: Storage class supports online expansion (most modern CSI drivers)
- **Process**:
  - **PVC size only changed**: Expand PVCs, leave StatefulSet unchanged (zero downtime)
  - **PVC size + other changes**: Expand PVCs, then recreate StatefulSet (minimal downtime)
- **Downtime**: None for size-only changes, minimal for mixed changes
- **Examples**: GCE Persistent Disk, AWS EBS (gp3, io1, io2), most modern CSI drivers

### Offline Expansion (When Required)
- **When**: Storage class requires PVC detachment for expansion
- **Process**:
  1. **Delete StatefulSet immediately** (VolumeClaimTemplates are immutable)
  2. Wait for all pods to terminate and PVCs to detach
  3. Expand PVCs while detached
  4. Wait for expansion completion
  5. Recreate StatefulSet with updated VolumeClaimTemplates
  6. Pods restart with expanded volumes
- **Downtime**: Temporary (during expansion process)
- **Examples**: Azure Disk (certain configurations), some legacy storage systems

> **Important**: StatefulSet deletion is necessary because VolumeClaimTemplates are immutable in Kubernetes. Once a StatefulSet is created, its VolumeClaimTemplates cannot be modified. To update storage sizes, the StatefulSet must be deleted and recreated with new VolumeClaimTemplates.

### Storage Classes Supporting Online Expansion
The operator uses a **conservative approach** and only performs online expansion for known-safe storage classes:

- **Google Cloud Persistent Disk** (`kubernetes.io/gce-pd`, `pd.csi.storage.gke.io`)
- **AWS EBS** (`kubernetes.io/aws-ebs`, `ebs.csi.aws.com`)
- **DigitalOcean Block Storage** (`dobs.csi.digitalocean.com`)
- **Linode Block Storage** (`linodebs.csi.linode.com`)
- **OpenEBS Local Storage** (`openebs.io/local`)
- **Longhorn Distributed Storage** (`driver.longhorn.io`)
- **Custom CSI drivers** with `expansion-mode: online` parameter

### Storage Classes Requiring Detachment (Default)
**All other storage classes** default to offline expansion for safety, including:
- **Azure Disk CSI** (`disk.csi.azure.com`)
- **Legacy Azure Disk** (`kubernetes.io/azure-disk`)
- **vSphere CSI** (`csi.vsphere.vmware.com`)
- **Unknown or custom CSI drivers** (unless explicitly marked as `expansion-mode: online`)

## Why StatefulSet Deletion is Required

In Kubernetes, StatefulSet VolumeClaimTemplates are **immutable** after creation. This means:

1. **Cannot modify storage size**: Once a StatefulSet is created, you cannot change the storage size in its VolumeClaimTemplates
2. **Cannot add/remove volumes**: The volume configuration is fixed at creation time
3. **Cannot change storage class**: The storage class reference is immutable

Therefore, to update PVC storage sizes, the operator must:
1. **Delete the existing StatefulSet** (preserving the PVCs)
2. **Expand the PVCs directly** while they are detached
3. **Recreate the StatefulSet** with updated VolumeClaimTemplates
4. **Attach the expanded PVCs** to the new pods

This process ensures that:
- **Data is preserved**: PVCs are never deleted, only expanded
- **Configuration is updated**: New StatefulSet has correct storage sizes
- **Consistency is maintained**: All replicas use the same expanded storage

## Conservative Expansion Strategy

The StarRocks operator uses a **conservative approach** to PVC expansion:

### Default Behavior: Offline Expansion
- **Unknown storage classes**: Default to offline expansion for safety
- **Untested provisioners**: Assume detachment is required
- **Custom CSI drivers**: Require explicit `expansion-mode: online` parameter

### Rationale for Conservative Approach
1. **Data Safety**: Prevents potential data corruption from failed online expansions
2. **Reliability**: Offline expansion works with all storage classes that support expansion
3. **Predictability**: Users know exactly what to expect during expansion
4. **Error Prevention**: Avoids storage-specific expansion failures

### Opting Into Online Expansion
To use online expansion with custom storage classes, add this parameter:
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: my-custom-storage
provisioner: my.custom.csi.driver
parameters:
  expansion-mode: online  # Explicitly enable online expansion
allowVolumeExpansion: true
```

## StatefulSet Immutability Optimization

The operator intelligently handles StatefulSet updates to minimize disruption:

### Scenario 1: Only PVC Size Changes
```yaml
# Before
spec:
  starRocksBeSpec:
    storageVolumes:
    - name: be-data
      storageSize: 100Gi  # Only this changes

# After
spec:
  starRocksBeSpec:
    storageVolumes:
    - name: be-data
      storageSize: 200Gi  # Only this changes
```

**Operator Behavior**:
1. Expand PVCs directly (online or offline based on storage class)
2. **Do NOT update StatefulSet** (avoids VolumeClaimTemplate immutability error)
3. **Zero downtime** for online expansion, temporary downtime for offline expansion

### Scenario 2: PVC Size + Other Changes
```yaml
# Before
spec:
  starRocksBeSpec:
    replicas: 3
    storageVolumes:
    - name: be-data
      storageSize: 100Gi

# After
spec:
  starRocksBeSpec:
    replicas: 5          # Replica count changed
    storageVolumes:
    - name: be-data
      storageSize: 200Gi  # Size also changed
```

**Operator Behavior**:
1. **Delete StatefulSet first** to ensure clean PVC detachment
2. **Wait for all pods to terminate** and PVCs to detach
3. **Expand PVCs while detached** (safer and more reliable)
4. **Recreate StatefulSet** with updated VolumeClaimTemplates and other changes
5. **Necessary downtime** (StatefulSet recreation required anyway)

### Why This Optimization Matters

Without this optimization, you would get this error on the next StatefulSet update:
```
StatefulSet.apps "celostar-be" is invalid: spec: Forbidden: updates to statefulset spec for fields other than 'replicas'
```

The operator prevents this by:
- **Detecting change scope**: Distinguishing between size-only and mixed changes
- **Minimizing disruption**: Avoiding unnecessary StatefulSet recreation
- **Maintaining consistency**: Ensuring PVCs and StatefulSet specs stay aligned

## 🔄 Updated Decision Matrix

The operator intelligently chooses the optimal expansion strategy based on the scope of changes and storage class capabilities:

| Change Type | Storage Class | Process | Downtime | StatefulSet Action |
|-------------|---------------|---------|----------|-------------------|
| **Size only** | Online (GCE, AWS) | Expand PVCs only | None | No update (template mismatch OK) |
| **Size only** | Offline (Azure) | Delete → Expand → Recreate | Minimal | Recreate (required for detachment) |
| **Size + replicas** | Any | Delete → Expand → Recreate | Necessary | Recreate (required for changes) |
| **Size + resources** | Any | Delete → Expand → Recreate | Necessary | Recreate (required for changes) |
| **Size + new volumes** | Any | Delete → Expand → Recreate | Necessary | Recreate (required for changes) |

### Key Decision Factors

1. **Change Scope**: Size-only vs. mixed changes
2. **Storage Class**: Online vs. offline expansion capability
3. **Disruption Minimization**: Avoid unnecessary StatefulSet recreation
4. **Safety**: Ensure PVCs are detached when required

### Optimized Process Flow

#### **Size-Only Changes (Zero Downtime)**:
```
1. Detect PVC expansion needed
2. Analyze scope of changes → Only storage sizes changed
3. Expand PVCs (online or offline based on storage class)
4. Skip StatefulSet update (avoid immutability error)
5. Complete with minimal/no downtime
```

#### **Mixed Changes (Necessary Downtime)**:
```
1. Detect PVC expansion needed
2. Analyze scope of changes → Storage sizes + other changes
3. Delete StatefulSet immediately
4. Wait for all pods to terminate and PVCs to detach
5. Expand PVCs while safely detached
6. Recreate StatefulSet with all updates (storage + other changes)
7. Complete with necessary downtime (unavoidable due to other changes)
```

## Monitoring Expansion Progress

### Check PVC Status
```bash
kubectl get pvc -n <namespace>
```

Look for the `STATUS` column. During expansion, you might see:
- `Bound` - Normal state
- `FileSystemResizePending` - Expansion in progress

### Check PVC Details
```bash
kubectl describe pvc <pvc-name> -n <namespace>
```

Look for events related to volume expansion.

### Check StatefulSet Status
```bash
kubectl get statefulset -n <namespace>
kubectl describe statefulset <statefulset-name> -n <namespace>
```

### Monitor Operator Logs
```bash
kubectl logs -n starrocks-operator-system deployment/starrocks-operator -f
```

Look for messages about PVC expansion and StatefulSet recreation.

### Check Cluster Status
```bash
kubectl get starrockscluster <cluster-name> -n <namespace> -o yaml
```

Check the `status` section for component phases and any error messages.

## Important Limitations and Notes

### Size Restrictions
- **Only increases allowed**: Storage sizes can only be increased, never decreased
- **Minimum increments**: Some storage providers have minimum increment requirements
- **Maximum limits**: Check your storage provider's maximum volume size limits

### Downtime Considerations
- **PVC expansion**: Usually non-disruptive
- **StatefulSet recreation**: May cause temporary pod restarts
- **Data preservation**: All data is preserved during the expansion process

### Storage Provider Specific Notes

#### AWS EBS
- Supports online expansion for most volume types
- GP3, GP2, IO1, IO2 volumes support expansion
- Expansion may take several minutes for large volumes

#### Google Cloud Persistent Disks
- Supports online expansion
- Both SSD and standard persistent disks can be expanded
- Automatic filesystem resize for most cases

#### Azure Disks
- Premium SSD and Standard SSD support expansion
- May require pod restart for filesystem resize

## Troubleshooting

### Common Issues

1. **Storage class doesn't support expansion**
   ```
   Error: Storage class validation failed for volume fe-meta: storage class standard does not support volume expansion (allowVolumeExpansion is not set to true)
   ```
   **Solution**: Use a storage class with `allowVolumeExpansion: true` or enable expansion on your current storage class

2. **Attempting to reduce storage size**
   ```
   Error: storage validation failed: Storage size reduction is not allowed
   ```
   **Solution**: Only increase storage sizes, never decrease them

3. **PVC stuck in FileSystemResizePending**
   ```bash
   kubectl describe pvc <pvc-name>
   ```
   **Solution**: Check if the pod needs to be restarted to complete filesystem resize

4. **StatefulSet recreation timeout**
   **Solution**: Check operator logs and ensure pods can be safely terminated

### Recovery Steps

If expansion fails:

1. **Check operator logs** for detailed error messages
2. **Verify storage class** supports expansion
3. **Ensure sufficient quota** in your cloud provider
4. **Check node capacity** for volume attachment limits
5. **Restart operator** if needed:
   ```bash
   kubectl rollout restart deployment/starrocks-operator -n starrocks-operator-system
   ```

## Best Practices

1. **Plan capacity**: Monitor storage usage and plan expansions in advance
2. **Test in staging**: Always test expansion procedures in non-production environments
3. **Backup data**: Although expansion preserves data, maintain regular backups
4. **Monitor performance**: Large volumes may have different performance characteristics
5. **Gradual expansion**: For very large increases, consider expanding in smaller increments

## Examples

See the complete example in `examples/starrocks/pvc_expansion_example.yaml` for detailed configuration examples.

## Support

For issues related to PVC expansion:
1. Check the troubleshooting section above
2. Review operator logs for detailed error messages
3. Consult your storage provider's documentation
4. Open an issue in the StarRocks Kubernetes operator repository
