# HPA Dynamic Scaling for CN Nodes with Helm Charts

This document describes how to implement Horizontal Pod Autoscaler (HPA) based dynamic scaling for CN (Compute Node) nodes in StarRocks clusters using Helm Charts. This feature was introduced in v1.11.0 and addresses critical autoscaling issues including resource cleanup, version compatibility, and graceful scaling operations.

## Prerequisites

- Kubernetes cluster v1.23.0 or higher
- [Helm](https://helm.sh/) v3.0.0 or higher
- [metrics-server](https://github.com/kubernetes-sigs/metrics-server) installed in your cluster

## Overview

The HPA feature enables automatic scaling of CN nodes based on resource metrics such as CPU and memory utilization. CN nodes are designed for elastic scaling as they handle query processing without affecting data distribution.

### Key Features Resolved

Since v1.11.0, the following critical issues have been resolved:
- **Resource Cleanup**: Proper HPA resource management when autoscaling policies are removed
- **Version Compatibility**: Support for multiple HPA API versions across different Kubernetes versions
- **Graceful Scaling**: Coordinated scaling operations with proper node registration/deregistration
- **Conflict Resolution**: Eliminated conflicts between operator and HPA replica management

## Deployment Architecture

There are two ways to deploy StarRocks with HPA-enabled CN nodes:

### 1. Integrated StarRocks Cluster with CN Nodes
Deploy a complete StarRocks cluster including FE and CN nodes with HPA in a single chart.

### 2. Separate Cluster and Warehouse Deployment
Deploy the main StarRocks cluster first, then deploy separate Warehouse instances with HPA-enabled CN nodes.

## Method 1: Integrated StarRocks Cluster with CN Nodes

### Step 1: Create Values Configuration

Create a `starrocks-cluster-values.yaml` file:

```yaml
starrocks:
  # Main cluster configuration
  starrocksCluster:
    enabledCn: true  # Enable CN nodes
    enabledBe: false # Disable BE nodes for shared-data mode

  # FE (Frontend) Configuration
  starrocksFESpec:
    # Three FE pods need to be running for high availability.
    replicas: 1
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.5.0
    resources:
      requests:
        cpu: 100m
        memory: 200Mi
      limits:
        cpu: 2
        memory: 4Gi
    storageSpec:
      logStorageSize: 1Gi
      name: fe
      storageSize: 10Gi
    persistentVolumeClaimRetentionPolicy:
      whenDeleted: Delete
    config: |
      run_mode = shared_data
      cloud_native_storage_type = S3
      aws_s3_path = your-bucket/path
      aws_s3_region = us-west-2
      aws_s3_endpoint = https://s3.us-west-2.amazonaws.com
      aws_s3_access_key = YOUR_ACCESS_KEY
      aws_s3_secret_key = YOUR_SECRET_KEY
      aws_s3_use_aws_sdk_default_behavior = true

  # CN (Compute Node) Configuration with HPA
  starrocksCnSpec:
    image:
      repository: starrocks/cn-ubuntu
      tag: 3.5.0
    storageSpec:
      logStorageSize: 1Gi
      name: cn
      storageSize: 10Gi
    persistentVolumeClaimRetentionPolicy:
      whenDeleted: Delete
      whenScaled: Delete
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
      limits:
        cpu: 8
        memory: 8Gi

    # HPA Configuration
    autoScalingPolicy:
      version: v2  # HPA API version (v2, v2beta2, v2beta1)
      minReplicas: 1    # Minimum number of CN pods
      maxReplicas: 2   # Maximum number of CN pods
      hpaPolicy:
        metrics:
          - type: Resource
            resource:
              name: cpu
              target:
                averageUtilization: 60  # Scale when CPU > 60%
                type: Utilization
          - type: Resource
            resource:
              name: memory
              target:
                averageUtilization: 70  # Scale when Memory > 70%
                type: Utilization
        behavior:
          scaleUp:
            policies:
              - type: Pods
                value: 2          # Add 2 pods at a time
                periodSeconds: 60 # Wait 60s between scale-up actions
          scaleDown:
            policies:
              - type: Pods
                value: 1          # Remove 1 pod at a time
                periodSeconds: 120 # Wait 120s between scale-down actions
            # selectPolicy: Disabled  # Uncomment to disable scale-down

# Operator Configuration
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.11.0-rc5
    imagePullPolicy: IfNotPresent
```

### Step 2: Deploy StarRocks Cluster

```bash
# Add StarRocks Helm repository
$ helm repo add starrocks https://starrocks.github.io/starrocks-kubernetes-operator
$ helm repo update

# Deploy the cluster
$ helm install kube-starrocks starrocks/kube-starrocks \
  -f starrocks-cluster-values.yaml \
  --namespace starrocks \
  --create-namespace

# The number of CN pods will be automatically managed by HPA based on CPU and memory usage.
$ kubectl get hpa
NAME                           REFERENCE                         TARGETS             MINPODS   MAXPODS   REPLICAS   AGE
kube-starrocks-cn-autoscaler   StarRocksCluster/kube-starrocks   15%/60%, 421%/70%   1         2         2          2m12s

$ kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
kube-starrocks-cn-0                        1/1     Running   0          2m17s
kube-starrocks-cn-1                        1/1     Running   0          106s
kube-starrocks-fe-0                        1/1     Running   0          3m6s
kube-starrocks-operator-5fd56d547b-7fp4m   1/1     Running   0          3m22s

# The pvc will be deleted when the CN nodes are scaled down or the cluster is deleted.
$ kubectl get pvc
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
cn-data-kube-starrocks-cn-0   Bound    pvc-d14e41ec-eafb-44c5-8482-43a7f01e46ef   10Gi       RWO            standard       3m7s
cn-data-kube-starrocks-cn-1   Bound    pvc-9f98dc6a-194a-4284-9474-721226669d85   10Gi       RWO            standard       2m36s
cn-log-kube-starrocks-cn-0    Bound    pvc-236c7e6b-d79b-40d0-a414-d4a3b657ce2f   1Gi        RWO            standard       3m7s
cn-log-kube-starrocks-cn-1    Bound    pvc-cd8f2d22-1368-4830-93a0-f19f9d67ede9   1Gi        RWO            standard       2m36s
fe-log-kube-starrocks-fe-0    Bound    pvc-9ef1467b-992b-4693-af92-c18214f91188   1Gi        RWO            standard       3m56s
fe-meta-kube-starrocks-fe-0   Bound    pvc-522e40b4-8a48-4a62-a593-1ca9174a1475   10Gi       RWO            standard       3m56s
```

## Method 2: Separate Cluster and Warehouse Deployment

Note: warehouse is an enterprise feature.

### Step 1: Deploy Main StarRocks Cluster

Create `cluster-values.yaml`:

```yaml
starrocks:
  # Main cluster configuration
  starrocksCluster:
    enabledCn: false
    enabledBe: false

  # FE (Frontend) Configuration
  starrocksFESpec:
    # Three FE pods need to be running for high availability.
    replicas: 1
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.5.0
    resources:
      requests:
        cpu: 100m
        memory: 200Mi
      limits:
        cpu: 2
        memory: 4Gi
    storageSpec:
      logStorageSize: 1Gi
      name: fe
      storageSize: 10Gi
    # whenScaled: Delete is not supported for FE nodes
    persistentVolumeClaimRetentionPolicy:
      whenDeleted: Delete
    config: |
      run_mode = shared_data
      cloud_native_storage_type = S3
      aws_s3_path = your-bucket/path
      aws_s3_region = us-west-2
      aws_s3_endpoint = https://s3.us-west-2.amazonaws.com
      aws_s3_access_key = YOUR_ACCESS_KEY
      aws_s3_secret_key = YOUR_SECRET_KEY
      aws_s3_use_aws_sdk_default_behavior = true

# Operator Configuration
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.11.0-rc5
    imagePullPolicy: IfNotPresent
```

Deploy the main cluster:

```bash
$ helm install kube-starrocks starrocks/kube-starrocks \
  -f cluster-values.yaml \
  --namespace starrocks \
  --create-namespace

$ kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
kube-starrocks-fe-0                        1/1     Running   0          73s
kube-starrocks-operator-5fd56d547b-46c9h   1/1     Running   0          91s

# The FE PVCS will be deleted when the StarRocks cluster is deleted.
$ kubectl get pvc
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
fe-log-kube-starrocks-fe-0    Bound    pvc-5fce212f-0f22-494c-925d-ffd17bada556   1Gi        RWO            standard       75s
fe-meta-kube-starrocks-fe-0   Bound    pvc-74b5ea92-c298-48eb-a7bf-5682679c4d7e   10Gi       RWO            standard       75s  
```

### Step 2: Deploy Warehouse with HPA

Create `warehouse-values.yaml`:

```yaml
spec:
  # Reference to the main StarRocks cluster
  starRocksClusterName: kube-starrocks

  image:
    repository: xxx/cn-ubuntu  # an enterprise image repository
    tag: 3.5.0

  resources:
    requests:
      cpu: 100m
      memory: 100Mi
    limits:
      cpu: 8
      memory: 8Gi

  # Storage configuration
  storageSpec:
    logStorageSize: 1Gi
    storageSize: 10Gi
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Delete
    whenScaled: Delete

  # HPA Autoscaling Policy
  autoScalingPolicy:
    version: v2
    minReplicas: 1
    maxReplicas: 2
    hpaPolicy:
      metrics:
        - type: Resource
          resource:
            name: cpu
            target:
              averageUtilization: 60
              type: Utilization
        - type: Resource
          resource:
            name: memory
            target:
              averageUtilization: 70
              type: Utilization
      behavior:
        scaleUp:
          policies:
            - type: Pods
              value: 1
              periodSeconds: 60
        scaleDown:
          policies:
            - type: Pods
              value: 1
              periodSeconds: 120
          selectPolicy: Max  # Use the policy that allows the highest scaling rate
```

Deploy the warehouse:

```bash
$ helm install wh-1 starrocks/warehouse \
  -f warehouse-values.yaml \
  --namespace starrocks

$ kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
kube-starrocks-fe-0                        1/1     Running   0          4m15s
kube-starrocks-operator-5fd56d547b-46c9h   1/1     Running   0          4m33s
wh-1-warehouse-cn-0                        1/1     Running   0          20s

$ kubectl get hpa
NAME                           REFERENCE                 TARGETS                        MINPODS   MAXPODS   REPLICAS   AGE
wh-1-warehouse-cn-autoscaler   StarRocksWarehouse/wh-1   <unknown>/60%, <unknown>/70%   1         2         1          22s  

# The CN PVCs will be deleted when the Warehouse is deleted or scaled down.
$ kubectl get pvc
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
cn-data-wh-1-warehouse-cn-0   Bound    pvc-e49fe13b-f5c3-42fc-90a9-c18adccec39c   10Gi       RWO            standard       21s
cn-log-wh-1-warehouse-cn-0    Bound    pvc-686d8dc7-55d5-4667-bec1-106ac9ea32c0   1Gi        RWO            standard       21s
fe-log-kube-starrocks-fe-0    Bound    pvc-5fce212f-0f22-494c-925d-ffd17bada556   1Gi        RWO            standard       5m50s
fe-meta-kube-starrocks-fe-0   Bound    pvc-74b5ea92-c298-48eb-a7bf-5682679c4d7e   10Gi       RWO            standard       5m50s
```

## Scale Down the CN nodes

```bash
# we scale down the number of CN nodes to 1 by modifying the HPA
kubectl patch hpa <your-hpa-name> -p '{"spec":{"minReplicas":1,"maxReplicas":1}}'

# pods will be scaled down gracefully
$ kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
kube-starrocks-fe-0                        1/1     Running   0          13m
kube-starrocks-operator-5fd56d547b-46c9h   1/1     Running   0          13m
wh-1-warehouse-cn-0                        1/1     Running   0          7m44s

# PVC will be deleted when the CN node is scaled down
$ kubectl get pvc
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
cn-data-wh-1-warehouse-cn-0   Bound    pvc-e49fe13b-f5c3-42fc-90a9-c18adccec39c   10Gi       RWO            standard       7m57s
cn-log-wh-1-warehouse-cn-0    Bound    pvc-686d8dc7-55d5-4667-bec1-106ac9ea32c0   1Gi        RWO            standard       7m57s
fe-log-kube-starrocks-fe-0    Bound    pvc-5fce212f-0f22-494c-925d-ffd17bada556   1Gi        RWO            standard       13m
fe-meta-kube-starrocks-fe-0   Bound    pvc-74b5ea92-c298-48eb-a7bf-5682679c4d7e   10Gi       RWO            standard       13m

# If you execute show compute nodes in the FE, you will see that the CN node is still deregistered
mysql> show compute nodes \G;
*************************** 1. row ***************************
        ComputeNodeId: 10010
                   IP: wh-1-warehouse-cn-0.wh-1-warehouse-cn-search.default.svc.cluster.local
                   ....
```