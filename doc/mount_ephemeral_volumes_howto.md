# Mount Generic Ephemeral Volumes

A `storageVolumes` entry with a real storage class becomes a `volumeClaimTemplate` of the FE, BE or
CN StatefulSet. Such a claim outlives its pod on purpose: when `cn-0` is recreated it binds to the
same PersistentVolumeClaim, and therefore to the same PersistentVolume.

For a volume that only holds a cache or spill data, that guarantee costs more than it gives:

- A PersistentVolume backed by a local disk carries node affinity. Drain or replace the node it was
  provisioned on and the replacement pod stays `Pending`, even though the cluster has capacity
  elsewhere.
- The claim is deleted only by a `persistentVolumeClaimRetentionPolicy` set to `Delete`, which
  defaults to `Retain`, and even then only when the StatefulSet is deleted or scaled down. A
  replacement pod never deletes it, so a volume holding nothing but a cache stays allocated for the
  life of the StatefulSet.
- An `emptyDir` avoids both, but it gives up the storage class and the size request, and it
  competes for the node's ephemeral storage instead of getting a volume of its own.

Setting `ephemeral: true` keeps the storage class and the size, and moves the claim into the pod
spec as a
[generic ephemeral volume](https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes).
Kubernetes creates the PersistentVolumeClaim when the pod is created, names it
`<pod-name>-<volume-name>`, sets its owner reference to the pod, and deletes it with the pod. The
next pod gets a new volume of its own, provisioned where the scheduler places it when the storage
class binds with `WaitForFirstConsumer`.

Use it only where the data can be rebuilt: a CN data cache, which is refilled from object storage
on the next read, or a spill directory, which belongs to a query that is already gone.

> **Prerequisite:** generic ephemeral volumes are generally available since Kubernetes 1.23.

## 1. By StarRocksCluster CRD

Set `ephemeral: true` on an entry of `storageVolumes`.

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
spec:
  starRocksFeSpec:
    replicas: 1
    image: starrocks/fe-ubuntu:latest
  starRocksCnSpec:
    replicas: 3
    image: starrocks/cn-ubuntu:latest
    storageVolumes:
      - name: cn-data
        storageClassName: local-nvme
        storageSize: 500Gi
        mountPath: /opt/starrocks/cn/storage
        ephemeral: true
      - name: cn-log
        storageClassName: gp3
        storageSize: 20Gi
        mountPath: /opt/starrocks/cn/log
```

Here `cn-log` stays in `volumeClaimTemplates` and `cn-data` does not.

Supported on `starRocksFeSpec`, `starRocksBeSpec`, `starRocksCnSpec`, and `starRocksFeProxySpec`.
FE metadata and, in a shared-nothing cluster, BE tablet replicas are not caches: making them
ephemeral means every replacement pod starts from an empty disk.

`emptyDir`, `hostPath` and `csi` volumes are not backed by a PersistentVolumeClaim, so `ephemeral`
does not apply to them. Combining them, through `storageClassName` or through a `hostPath` or `csi`
block, is rejected with `ephemeral can not be used together with emptyDir, hostPath or csi`.

## 2. By Helm chart

`storageSpec` exposes `storageEphemeral` for the data volumes and `spillEphemeral` for the spill
volume, on `starrocksCnSpec` and `starrocksBeSpec`:

```yaml
starrocksCnSpec:
  storageSpec:
    name: cn
    storageClassName: local-nvme
    storageSize: 500Gi
    storageEphemeral: true
    spillStorageSize: 200Gi
    spillEphemeral: true
    logStorageClassName: gp3
```

`logStorageClassName` falls back to `storageClassName` when it is empty. Left on the local class,
the log volume stays a `volumeClaimTemplate` with node affinity and keeps the pod bound to one
node, so give it a class of its own, or set `logStorageSize: 0Gi` to mount an `emptyDir` and
provision nothing.

When using the parent `kube-starrocks` chart, nest this under the `starrocks:` key.

## 3. Verifying

Deploy the cluster, then look at where each claim came from:

```bash
kubectl get statefulset starrockscluster-sample-cn -o jsonpath='{.spec.volumeClaimTemplates[*].metadata.name}'
```

Expected: `cn-log` only.

```bash
kubectl get pvc -o custom-columns=NAME:.metadata.name,OWNER:.metadata.ownerReferences[*].kind
```

Expected, for `cn-0`, and the same pair for every other pod:

```
NAME                                   OWNER
cn-log-starrockscluster-sample-cn-0    <none>
starrockscluster-sample-cn-0-cn-data   Pod
```

The second claim is owned by the pod, so deleting the pod deletes it, while the `cn-log` claim
stays. The replacement pod gets a new claim of its own.

## 4. Changing a volume after it is deployed

`storageSize` of an ephemeral volume can be changed. The claim template lives in the pod template,
which the operator is free to update, so the change rolls the pods and every new pod gets a claim
at the new size. The same change on a `volumeClaimTemplate` is rejected by Kubernetes, and the
operator reports it in `status.reason`, which ends with:

```
Forbidden: updates to statefulset spec for fields other than 'replicas', 'ordinals', 'template', 'updateStrategy', 'persistentVolumeClaimRetentionPolicy' and 'minReadySeconds' are forbidden
```

Turning `ephemeral` on or off for a volume that is already deployed hits that same error, because
it adds or removes a `volumeClaimTemplate`. Apply the change to the StarRocksCluster first, then
delete the StatefulSet with `--cascade=orphan`: the operator finds none on the next reconcile,
creates it from the current spec, adopts the running pods and rolls them onto the new template.
In the other order the operator recreates the StatefulSet from the old spec before there is
anything new to apply. The PersistentVolumeClaims the old template created survive that delete,
so remove them once the pods no longer use them.
