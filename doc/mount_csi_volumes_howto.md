# Mount CSI Ephemeral Volumes

Some Kubernetes drivers publish a resource to a pod through a **CSI ephemeral inline volume**
rather than through a PersistentVolumeClaim. The SPIFFE CSI driver is the common example: it
mounts the SPIRE Agent's Workload API socket into the pod so the workload can obtain its
identity document (SVID).

A PVC cannot carry that kind of volume:

- Dynamic provisioning goes through the CSI Controller Service (`CreateVolume`). Drivers like
  `csi.spiffe.io` implement only the Node Service and declare `volumeLifecycleModes: [Ephemeral]`,
  so a PVC against them stays `Pending` forever.
- A bound PersistentVolume carries node affinity, which would stop the pod from ever being
  rescheduled to another node. The socket, however, is node-local and exists on every node.
- A PVC is meant to outlive its pod. This kind of volume must disappear with the pod.

The operator therefore accepts a `csi` volume source alongside `emptyDir` and `hostPath`.

> **Prerequisite:** the CSI driver must already be installed in the cluster. The operator only
> references it by name; it does not install anything.

## 1. By StarRocksCluster CRD

Set `csi` on an entry of `storageVolumes`. The `csi` field takes a Kubernetes
[CSIVolumeSource](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/volume/#csi).

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: starrockscluster-sample
spec:
  starRocksFeSpec:
    replicas: 3
    image: starrocks/fe-ubuntu:latest
    storageVolumes:
      - name: spiffe-workload-api
        storageClassName: csi
        mountPath: /spiffe-workload-api
        readOnly: true
        csi:
          driver: csi.spiffe.io
          readOnly: true
```

`storageClassName: csi` may be omitted — setting the `csi` field alone is enough. Spelling it out
makes the intent obvious next to neighbouring PVC-backed volumes.

Supported on `starRocksFeSpec`, `starRocksBeSpec`, `starRocksCnSpec`, and `starRocksFeProxySpec`.

The operator rejects configurations that would otherwise be silently ignored:

| Configuration | Error |
| --- | --- |
| `storageClassName: csi` without a `csi` block, or with an empty `csi.driver` | `csi is required if storageClassName is csi, and csi.driver must not be empty` |
| `csi` together with `hostPath` on the same volume | `csi and hostPath can not be set at the same time` |
| `csi` together with any other `storageClassName` (`gp3`, `emptyDir`, ...) | `if csi is set, storageClassName must be empty or "csi"` |

## 2. By Helm chart

Each component has a `csiVolumes` list, next to `emptyDirs` and `hostPaths`:

```yaml
starrocksFESpec:
  csiVolumes:
    - name: spiffe-workload-api
      mountPath: /spiffe-workload-api
      readOnly: true
      csi:
        driver: csi.spiffe.io
        readOnly: true
```

When using the parent `kube-starrocks` chart, nest this under the `starrocks:` key.

## 3. Example: mounting the SPIFFE workload API socket

Install the SPIFFE CSI driver first, following the
[SPIRE documentation](https://github.com/spiffe/spiffe-csi). Confirm it registered:

```bash
kubectl get csidriver csi.spiffe.io
```

Then deploy the cluster:

```bash
helm install starrocks starrocks/kube-starrocks -f values.yaml
```

with `values.yaml`:

```yaml
starrocks:
  starrocksFESpec:
    csiVolumes:
      - name: spiffe-workload-api
        mountPath: /spiffe-workload-api
        readOnly: true
        csi:
          driver: csi.spiffe.io
          readOnly: true
```

Verify the socket reached the pod:

```bash
kubectl exec starrockscluster-sample-fe-0 -- ls -l /spiffe-workload-api
```

Expected: a `spire-agent.sock` entry.
