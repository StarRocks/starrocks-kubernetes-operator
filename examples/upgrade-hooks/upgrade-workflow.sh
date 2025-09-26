#!/bin/bash

# StarRocks Cluster Upgrade Workflow with Hooks
# This script demonstrates the complete upgrade workflow using the upgrade hooks feature

set -e

CLUSTER_NAME="starrockscluster-upgrade-example"
NAMESPACE="starrocks" 
NEW_FE_VERSION="starrocks/fe-ubuntu:3.2.2"
NEW_BE_VERSION="starrocks/be-ubuntu:3.2.2"

echo "🚀 Starting StarRocks cluster upgrade workflow"

# Function to wait for a specific phase
wait_for_phase() {
    local expected_phase=$1
    local timeout=${2:-300}
    local elapsed=0
    
    while [ $elapsed -lt $timeout ]; do
        local current_phase=$(kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus.phase}' 2>/dev/null || echo "")
        
        if [ "$current_phase" = "$expected_phase" ]; then
            echo "✅ Reached phase: $expected_phase"
            return 0
        fi
        
        echo "⏳ Current phase: $current_phase, waiting for: $expected_phase (${elapsed}s/${timeout}s)"
        sleep 10
        elapsed=$((elapsed + 10))
    done
    
    echo "❌ Timeout waiting for phase: $expected_phase"
    return 1
}

# Function to check cluster status
check_cluster_status() {
    echo "📊 Cluster Status:"
    kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.phase}' | xargs echo "  Phase:"
    
    local prep_status=$(kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus}' 2>/dev/null || echo "null")
    if [ "$prep_status" != "null" ] && [ "$prep_status" != "" ]; then
        echo "  Upgrade Preparation:"
        kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus.phase}' | xargs echo "    Phase:"
        kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus.reason}' | xargs echo "    Reason:"
        
        local completed_hooks=$(kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus.completedHooks}' 2>/dev/null || echo "[]")
        if [ "$completed_hooks" != "[]" ] && [ "$completed_hooks" != "" ]; then
            echo "    Completed Hooks: $completed_hooks"
        fi
        
        local failed_hooks=$(kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.upgradePreparationStatus.failedHooks}' 2>/dev/null || echo "[]")
        if [ "$failed_hooks" != "[]" ] && [ "$failed_hooks" != "" ]; then
            echo "    Failed Hooks: $failed_hooks"
        fi
    fi
}

# Step 1: Verify cluster exists and is ready
echo "1️⃣  Verifying cluster status..."
if ! kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE >/dev/null 2>&1; then
    echo "❌ Cluster $CLUSTER_NAME not found in namespace $NAMESPACE"
    echo "Please deploy the cluster first using:"
    echo "  kubectl apply -f starrockscluster-with-upgrade-hooks.yaml"
    exit 1
fi

check_cluster_status

# Step 2: Trigger upgrade preparation (annotation is already present, so we'll trigger by changing image)
echo -e "\n2️⃣  Triggering upgrade preparation by updating FE image..."
kubectl patch starrockscluster $CLUSTER_NAME -n $NAMESPACE \
  --type='merge' \
  -p "{\"spec\":{\"starRocksFeSpec\":{\"image\":\"$NEW_FE_VERSION\"}}}"

# Step 3: Wait for upgrade preparation to complete
echo -e "\n3️⃣  Waiting for upgrade preparation to complete..."
wait_for_phase "Completed" 600

# Step 4: Check preparation results
echo -e "\n4️⃣  Checking preparation results..."
check_cluster_status

# Step 5: Apply BE upgrade
echo -e "\n5️⃣  Updating BE image..."
kubectl patch starrockscluster $CLUSTER_NAME -n $NAMESPACE \
  --type='merge' \
  -p "{\"spec\":{\"starRocksBeSpec\":{\"image\":\"$NEW_BE_VERSION\"}}}"

# Step 6: Monitor upgrade progress
echo -e "\n6️⃣  Monitoring upgrade progress..."
echo "You can monitor the upgrade with these commands:"
echo "  kubectl get starrockscluster $CLUSTER_NAME -n $NAMESPACE -w"
echo "  kubectl get pods -n $NAMESPACE -l app.kubernetes.io/instance=$CLUSTER_NAME -w"

# Step 7: Post-upgrade cleanup (optional)
echo -e "\n7️⃣  Post-upgrade cleanup options:"
echo "To re-enable services after upgrade:"
echo "  kubectl annotate starrockscluster $CLUSTER_NAME -n $NAMESPACE starrocks.com/upgrade-hooks=enable-tablet-clone,enable-balancer"
echo ""
echo "To clean up preparation annotation:"
echo "  kubectl annotate starrockscluster $CLUSTER_NAME -n $NAMESPACE starrocks.com/prepare-upgrade-"

echo -e "\n✅ Upgrade workflow initiated successfully!"
echo "Monitor the upgrade progress and run post-upgrade cleanup when ready."

# Optional: Wait for pods to be ready
read -p "Wait for upgrade to complete? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "⏳ Waiting for FE pods to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=fe,app.kubernetes.io/instance=$CLUSTER_NAME -n $NAMESPACE --timeout=600s
    
    echo "⏳ Waiting for BE pods to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=be,app.kubernetes.io/instance=$CLUSTER_NAME -n $NAMESPACE --timeout=600s
    
    echo "🎉 Upgrade completed successfully!"
    check_cluster_status
fi
