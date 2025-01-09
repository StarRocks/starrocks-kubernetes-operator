# StarRocks Logging and Related Configurations

StarRocks is composed of three components: FE, BE, and CN. Among them, FE is essential. This document describes the
logging and associated configurations for StarRocks components, including:

1. Location of log storage;
2. Default storage volume;
3. Persisting logs;
4. Logging to the console;
5. Collecting logs into Datadog.

## 1. Location of Log Storage

1. The FE component's logs are located in: `/opt/starrocks/fe/log`, with key logs
   including: `fe.out`, `fe.log`, `fe.warn.log`.
2. The BE component's logs are located in: `/opt/starrocks/be/log`, with key logs
   including: `be.out`, `be.INFO`, `be.WARNING`.
3. The CN component's logs are located in: `/opt/starrocks/cn/log`, with key logs
   including: `cn.out`, `cn.INFO`, `cn.WARNING`.

## 2. Default Storage Volume

By default, all components use the `emptyDir` storage volume. One inherent problem is that once a Pod restarts, logs
before the restart will not be accessible anymore, which obviously complicates troubleshooting. To address this, one of
two approaches can be adopted:

1. Persist the logs so that logs from prior to the Pod restart remain available.
2. Log to the console and view logs from prior to the restart using `kubectl logs my-pod -p`.

## 3. Persisting Logs

All component Spec definitions have a `storageVolumes` field, allowing users to customize the storage volume. Taking FE
as an example:

```yaml
spec:
  starRocksFeSpec:
    storageVolumes:
      - mountPath: /opt/starrocks/fe/log
        name: fe-log
        storageSize: 10Gi
        # storageClassName: ""  # If storageClassName is not set, Kubernetes will use the default storage class.
```

If `storageClassName` is left blank, the default storage class will be used. You can view available storage classes in
the Kubernetes cluster with `kubectl get storageclass`. **Note: selecting an appropriate storage class is crucial as it
dictates the type of storage volume**. See https://kubernetes.io/docs/concepts/storage/persistent-volumes/ for more
information.

> Attention: The Operator will create PVC resources for the StarRocks cluster. The storage class controller will then
> automatically generate the specific storage volume.

### 3.1 Helm Chart Supports Persisting Logs

If you deployed the StarRocks cluster using Helm Chart, you can modify the `values.yaml` content to persist logs. Here's
an example for the FE component:

For the kube-starrocks Helm Chart, you can configure as:

```yaml
starrocks:
  starrocksFESpec:
    storageSpec:
      name: "fe"
      storageSize: 10Gi
      logStorageSize: 10Gi
      # storageClassName: ""  # If storageClassName is not set, Kubernetes will use the default storage class.
```

For the starrocks Helm Chart, configure as:

```yaml
starrocksFESpec:
  storageSpec:
    name: "fe"
    storageSize: 10Gi
    logStorageSize: 10Gi
    # storageClassName: ""  # If storageClassName is not set, Kubernetes will use the default storage class.
```

> Note:
> 1. In FE, `storageSize` specifies the size of the storage volume for metadata, while `logStorageSize` designates the
     size of the storage volume for logs.
> 2. Fe container stop running if the storage volume free space which the fe meta residents, is less than 5Gi. Set it to
     at least 10GB or more.

## 4. Logging to the Console

By setting the environment variable `LOG_CONSOLE = 1`, you can direct component logs to the console. Here's an example
for FE:

```yaml
spec:
  starRocksFeSpec:
    feEnvVars:
      - name: LOG_CONSOLE
        value: "1"
```

### 4.1 Helm Chart Supports Environment Variable Settings

If you've deployed the StarRocks cluster using Helm Chart, you can modify the `values.yaml` content to set environment
variables. An example for the FE component:

For the kube-starrocks Helm Chart, configure as:

```yaml
starrocks:
  starrocksFESpec:
    feEnvVars:
      - name: LOG_CONSOLE
        value: "1"
```

For the starrocks Helm Chart, configure as:

```yaml
starrocksFESpec:
  feEnvVars:
    - name: LOG_CONSOLE
      value: "1"
```

## 5. Collecting Logs into Datadog

Refer to: [Datadog](./integration/integration-with-datadog.md).
