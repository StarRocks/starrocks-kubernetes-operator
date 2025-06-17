#!/usr/bin/env bash

# Integration test script for PVC expansion feature
# This script demonstrates the PVC expansion functionality

set -e

NAMESPACE="pvc-expansion-test"
CLUSTER_NAME="test-cluster"
TIMEOUT=300

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed"
        exit 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        error "kubectl is not connected to a cluster"
        exit 1
    fi

    # Check if the StarRocks operator is running
    if ! kubectl get deployment starrocks-operator -n starrocks-operator-system &> /dev/null; then
        warn "StarRocks operator not found. Please install the operator first."
        warn "You can install it using: helm install starrocks-operator kube-starrocks/operator"
    fi

    # Check if default storage class supports expansion
    check_storage_class_expansion

    log "Prerequisites check completed"
}

# Check if storage class supports expansion
check_storage_class_expansion() {
    log "Checking storage class expansion support..."

    # Get default storage class
    local default_sc=$(kubectl get storageclass -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}')

    if [ -z "$default_sc" ]; then
        # Try beta annotation
        default_sc=$(kubectl get storageclass -o jsonpath='{.items[?(@.metadata.annotations.storageclass\.beta\.kubernetes\.io/is-default-class=="true")].metadata.name}')
    fi

    if [ -z "$default_sc" ]; then
        warn "No default storage class found. PVC expansion may not work."
        return
    fi

    local expansion_support=$(kubectl get storageclass "$default_sc" -o jsonpath='{.allowVolumeExpansion}')

    if [ "$expansion_support" != "true" ]; then
        warn "Default storage class '$default_sc' does not support volume expansion."
        warn "PVC expansion tests may fail. Consider using a storage class with allowVolumeExpansion: true"
    else
        log "Default storage class '$default_sc' supports volume expansion ✓"

        # Check expansion mode based on conservative approach
        local provisioner=$(kubectl get storageclass "$default_sc" -o jsonpath='{.provisioner}')
        local expansion_mode=$(kubectl get storageclass "$default_sc" -o jsonpath='{.parameters.expansion-mode}' 2>/dev/null || echo "")

        # Check if provisioner supports online expansion
        case "$provisioner" in
            "kubernetes.io/gce-pd"|"pd.csi.storage.gke.io"|"kubernetes.io/aws-ebs"|"ebs.csi.aws.com"|"dobs.csi.digitalocean.com"|"linodebs.csi.linode.com"|"openebs.io/local"|"driver.longhorn.io")
                log "Storage class '$default_sc' (provisioner: $provisioner) supports online expansion ✓"
                ;;
            *)
                if [ "$expansion_mode" = "online" ]; then
                    log "Storage class '$default_sc' explicitly configured for online expansion ✓"
                else
                    warn "Storage class '$default_sc' (provisioner: $provisioner) will use offline expansion (conservative default)."
                    warn "This will cause temporary downtime during expansion as StatefulSet needs to be recreated."
                    warn "To enable online expansion, add 'expansion-mode: online' parameter if your storage supports it."
                fi
                ;;
        esac
    fi
}

# Create test namespace
create_namespace() {
    log "Creating test namespace: $NAMESPACE"
    kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
}

# Create initial cluster with small storage
create_initial_cluster() {
    log "Creating initial StarRocks cluster with small storage..."

    cat <<EOF | kubectl apply -f -
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: $CLUSTER_NAME
  namespace: $NAMESPACE
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.3-latest
    replicas: 1
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 1
      memory: 2Gi
    storageVolumes:
    - name: fe-meta
      storageSize: 1Gi
      mountPath: /opt/starrocks/fe/meta
    - name: fe-log
      storageSize: 1Gi
      mountPath: /opt/starrocks/fe/log
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.3-latest
    replicas: 1
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 1
      memory: 2Gi
    storageVolumes:
    - name: be-data
      storageSize: 2Gi
      mountPath: /opt/starrocks/be/storage
    - name: be-log
      storageSize: 1Gi
      mountPath: /opt/starrocks/be/log
EOF
}

# Wait for cluster to be ready
wait_for_cluster_ready() {
    log "Waiting for cluster to be ready..."

    local timeout=$TIMEOUT
    while [ $timeout -gt 0 ]; do
        local fe_ready=$(kubectl get statefulset ${CLUSTER_NAME}-fe -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
        local be_ready=$(kubectl get statefulset ${CLUSTER_NAME}-be -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")

        if [ "$fe_ready" = "1" ] && [ "$be_ready" = "1" ]; then
            log "Cluster is ready!"
            return 0
        fi

        log "Waiting for cluster... FE ready: $fe_ready/1, BE ready: $be_ready/1"
        sleep 10
        timeout=$((timeout - 10))
    done

    error "Timeout waiting for cluster to be ready"
    return 1
}

# Check initial PVC sizes
check_initial_pvc_sizes() {
    log "Checking initial PVC sizes..."

    local fe_meta_size=$(kubectl get pvc fe-meta-${CLUSTER_NAME}-fe-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
    local fe_log_size=$(kubectl get pvc fe-log-${CLUSTER_NAME}-fe-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
    local be_data_size=$(kubectl get pvc be-data-${CLUSTER_NAME}-be-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
    local be_log_size=$(kubectl get pvc be-log-${CLUSTER_NAME}-be-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')

    log "Initial PVC sizes:"
    log "  FE Meta: $fe_meta_size"
    log "  FE Log: $fe_log_size"
    log "  BE Data: $be_data_size"
    log "  BE Log: $be_log_size"
}

# Expand storage sizes
expand_storage() {
    log "Expanding storage sizes..."

    cat <<EOF | kubectl apply -f -
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: $CLUSTER_NAME
  namespace: $NAMESPACE
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.3-latest
    replicas: 1
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 1
      memory: 2Gi
    storageVolumes:
    - name: fe-meta
      storageSize: 2Gi  # Expanded from 1Gi
      mountPath: /opt/starrocks/fe/meta
    - name: fe-log
      storageSize: 2Gi  # Expanded from 1Gi
      mountPath: /opt/starrocks/fe/log
  starRocksBeSpec:
    image: starrocks/be-ubuntu:3.3-latest
    replicas: 1
    limits:
      cpu: 2
      memory: 4Gi
    requests:
      cpu: 1
      memory: 2Gi
    storageVolumes:
    - name: be-data
      storageSize: 4Gi  # Expanded from 2Gi
      mountPath: /opt/starrocks/be/storage
    - name: be-log
      storageSize: 2Gi  # Expanded from 1Gi
      mountPath: /opt/starrocks/be/log
EOF
}

# Wait for PVC expansion to complete
wait_for_pvc_expansion() {
    log "Waiting for PVC expansion to complete..."

    local timeout=$TIMEOUT
    while [ $timeout -gt 0 ]; do
        local fe_meta_size=$(kubectl get pvc fe-meta-${CLUSTER_NAME}-fe-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
        local fe_log_size=$(kubectl get pvc fe-log-${CLUSTER_NAME}-fe-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
        local be_data_size=$(kubectl get pvc be-data-${CLUSTER_NAME}-be-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')
        local be_log_size=$(kubectl get pvc be-log-${CLUSTER_NAME}-be-0 -n $NAMESPACE -o jsonpath='{.spec.resources.requests.storage}')

        if [ "$fe_meta_size" = "2Gi" ] && [ "$fe_log_size" = "2Gi" ] && [ "$be_data_size" = "4Gi" ] && [ "$be_log_size" = "2Gi" ]; then
            log "PVC expansion completed successfully!"
            log "New PVC sizes:"
            log "  FE Meta: $fe_meta_size"
            log "  FE Log: $fe_log_size"
            log "  BE Data: $be_data_size"
            log "  BE Log: $be_log_size"
            return 0
        fi

        log "Waiting for PVC expansion... Current sizes: FE Meta: $fe_meta_size, FE Log: $fe_log_size, BE Data: $be_data_size, BE Log: $be_log_size"
        sleep 10
        timeout=$((timeout - 10))
    done

    error "Timeout waiting for PVC expansion to complete"
    return 1
}

# Verify cluster is still healthy after expansion
verify_cluster_health() {
    log "Verifying cluster health after expansion..."

    local fe_ready=$(kubectl get statefulset ${CLUSTER_NAME}-fe -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
    local be_ready=$(kubectl get statefulset ${CLUSTER_NAME}-be -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')

    if [ "$fe_ready" = "1" ] && [ "$be_ready" = "1" ]; then
        log "Cluster is healthy after expansion!"
        return 0
    else
        error "Cluster is not healthy after expansion. FE ready: $fe_ready/1, BE ready: $be_ready/1"
        return 1
    fi
}

# Test storage size reduction (should fail)
test_storage_reduction() {
    log "Testing storage size reduction (should fail)..."

    cat <<EOF > /tmp/reduction_test.yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: $CLUSTER_NAME
  namespace: $NAMESPACE
spec:
  starRocksFeSpec:
    image: starrocks/fe-ubuntu:3.3-latest
    replicas: 1
    storageVolumes:
    - name: fe-meta
      storageSize: 1Gi  # Attempting to reduce from 2Gi
      mountPath: /opt/starrocks/fe/meta
EOF

    if kubectl apply -f /tmp/reduction_test.yaml 2>&1 | grep -q "validation failed\|not allowed"; then
        log "Storage size reduction correctly rejected!"
        rm -f /tmp/reduction_test.yaml
        return 0
    else
        warn "Storage size reduction was not rejected as expected"
        rm -f /tmp/reduction_test.yaml
        return 1
    fi
}

# Cleanup
cleanup() {
    log "Cleaning up test resources..."
    kubectl delete starrockscluster $CLUSTER_NAME -n $NAMESPACE --ignore-not-found=true
    kubectl delete namespace $NAMESPACE --ignore-not-found=true
    rm -f /tmp/reduction_test.yaml
}

# Main test execution
main() {
    log "Starting PVC expansion integration test..."

    # Set up trap for cleanup
    trap cleanup EXIT

    check_prerequisites
    create_namespace
    create_initial_cluster
    wait_for_cluster_ready
    check_initial_pvc_sizes
    expand_storage
    wait_for_pvc_expansion
    verify_cluster_health
    test_storage_reduction

    log "PVC expansion integration test completed successfully!"
}

# Run the test
main "$@"
