# ComputeNode

ComputeNode(cn) is a compute node in starrocks, without local storage data, can execute queries except OlapScanNode and OlapTableSink. 
starrocks-operator is a cloud native technology, implement the control of ComputeNode. 

## Kubernetes Resources 

1. ComputeNodeGroup CR
2. cn-operator Deployment
3. CN Deployment (CN、RegisterSidecar)
4. Offline CronJob
5. HPA
6. ClusterRoleBinding
7. ServiceAccount

## The process of controller 

### ComputeNodeGroup Creation
* Create kubernetes resources(deployment, cronJob, hpa, clusterRoleBinding, serviceAccount) according to the ComputeNodeGroup CR.
* Set finalizers to the ComputeNodeGroup CR.

### ComputeNodeGroup Deletion
* Set the replicas of the Deployment to 0, after all the pods are deleted, 
drop ComputeNodes on FE which alive == false.
* Clean finalizers

### ComputeNodeGroup Scale Up

Set the replicas of the Deployment to the desired number, 
each cn-pod contains a register-sidecar for registering cn-pod's ip to fe.

### ComputeNodeGroup Scale Down

Set the replicas of the Deployment to the desired number,
obsoleted CN on FE will be clean by the offline-job

### auto-scaling

Use HPA to scale up and down the ComputeNodeGroup.

## TODO

- FE ComputeNode-Table add label column to distinguish different cr
- multiple auto-scaling strategy
