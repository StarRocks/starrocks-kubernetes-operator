# Automatic scaling for CN nodes

This topic describes how to configure automatic scaling for CN nodes in a StarRocks cluster.
> Note: If you are seeking for help for versions before v1.10.2, this doc is for you. If you are using v1.10.2 or later, please refer to the [HPA Dynamic Scaling](./hpa_dynamic_scaling_with_helm_howto.md) doc.

## Prerequisites

- Ensure that you have installed the Kubernetes cluster. v1.23.0+ is recommended.
- Ensure that you have installed the [Helm](https://helm.sh/) package manager. 3.0.0+ is recommended.
- Ensure that the helm chart repo for StarRocks is added.
  See [Add the Helm Chart Repo for StarRocks](./add_helm_repo_howto.md).
- Ensure that you have deployed a StarRocks cluster.
  See [Deploy StarRocks With Operator](./deploy_starrocks_with_operator_howto.md)
  or [Deploy StarRocks With Helm](./deploy_starrocks_with_helm_howto.md)
  to learn how to deploy a StarRocks cluster.
- Ensure that you have installed [metrics-server](https://github.com/kubernetes-sigs/metrics-server).

> Suppose you have installed a StarRocks cluster named starrockscluster-sample under the starrocks namespace

There are two ways to deploy StarRocks cluster:

1. Deploy StarRocks cluster with `StarRocksCluster` CR yaml.
2. Deploy StarRocks cluster with Helm chart.

Therefore, there are two ways to configure automatic scaling for CN nodes.

You can specify the resource metrics for CNs, such as average CPU utilization, average memory
usage, elastic scaling threshold, upper elastic scaling limit, and lower elastic scaling limit. The upper elastic
scaling limit and lower elastic scaling limit specify the maximum number and minimum number of CNs allowed for elastic
scaling.

## Configure automatic scaling for CN nodes by using CR yaml

Run the command `kubectl -n starrocks edit src starrockscluster-sample`` to configure the automatic scaling policy for
CN nodes.

> **NOTE**
>
> If you have configured the automatic scaling policy for the CN cluster, delete the `replicas` field from the
> `starRocksCnSpec` in the StarRocks cluster configuration file.

The following is a [template](../examples/starrocks/deploy_a_starrocks_cluster_with_cn.yaml) to help you configure
automatic scaling policies:

```YAML
  starRocksCnSpec:
    image: starrocks/cn-ubuntu:3.0-latest
    requests:
      cpu: 4
      memory: 4Gi
    autoScalingPolicy: # Automatic scaling policy of the CN cluster.
      maxReplicas: 10 # The maximum number of CNs is set to 10.
      minReplicas: 1 # The minimum number of CNs is set to 1.
      hpaPolicy:
        metrics: # Resource metrics
          - type: Resource
            resource:
              name: memory # The average memory usage of CNs is specified as a resource metric.
              target:
                averageUtilization: 30
                # The elastic scaling threshold is 30%.
                # When the average memory utilization of CNs exceeds 30%, the number of CNs increases for scale-out.
                # When the average memory utilization of CNs is below 30%, the number of CNs decreases for scale-in.
                type: Utilization
          - type: Resource
            resource:
              name: cpu # The average CPU utilization of CNs is specified as a resource metric.
              target:
                averageUtilization: 60
                # The elastic scaling threshold is 60%.
                # When the average CPU utilization of CNs exceeds 60%, the number of CNs increases for scale-out.
                # When the average CPU utilization of CNs is below 60%, the number of CNs decreases for scale-in.
                type: Utilization
        behavior: #  The scaling behavior is customized according to business scenarios, helping you achieve rapid or slow scaling or disable scaling.
          scaleUp:
            policies:
              - type: Pods
                value: 1
                periodSeconds: 10
          scaleDown:
            selectPolicy: Disabled
```

## Configure automatic scaling for CN nodes by using Helm chart

Add the following snippets to `values.yaml` to configure the automatic scaling policy for CN nodes,

```YAML
  starrocksCluster: # do not forget to set enabledCn to true to enable deployment of CNs.
    enabledCn: true

  starrocksCnSpec:
    image:
      repository: starrocks/cn-ubuntu
      tag: 3.5-latest
    resources:
      requests:
        cpu: 4
        memory: 4Gi
    autoScalingPolicy:
      minReplicas: 1
      maxReplicas: 10
      hpaPolicy:
        metrics: # Resource metrics
          - type: Resource
            resource:
              name: memory # The average memory usage of CNs is specified as a resource metric.
              target:
                averageUtilization: 30
                # The elastic scaling threshold is 30%.
                # When the average memory utilization of CNs exceeds 30%, the number of CNs increases for scale-out.
                # When the average memory utilization of CNs is below 30%, the number of CNs decreases for scale-in.
                type: Utilization
          - type: Resource
            resource:
              name: cpu # The average CPU utilization of CNs is specified as a resource metric.
              target:
                averageUtilization: 60
                # The elastic scaling threshold is 60%.
                # When the average CPU utilization of CNs exceeds 60%, the number of CNs increases for scale-out.
                # When the average CPU utilization of CNs is below 60%, the number of CNs decreases for scale-in.
                type: Utilization
        behavior: #  The scaling behavior is customized according to business scenarios, helping you achieve rapid or slow scaling or disable scaling.
          scaleUp:
            policies:
              - type: Pods
                value: 1
                periodSeconds: 10
          scaleDown:
            selectPolicy: Disabled
```

## Fields description

The following are descriptions of a few important fields:

- The upper and lower limits for elastic scaling.

  ```YAML
  maxReplicas: 10 # The maximum number of CNs is set to 10.
  minReplicas: 1 # The minimum number of CNs is set to 1.
  ```

- The threshold for elastic scaling.

  ```YAML
  # For example, the average CPU utilization of CNs is specified as a resource metric.
  # The elastic scaling threshold is 60%.
  # When the average CPU utilization of CNs exceeds 60%, the number of CNs increases for scale-out.
  # When the average CPU utilization of CNs is below 60%, the number of CNs decreases for scale-in.
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 60
        type: Utilization
  ```

- The `behavior` for elastic scaling.

  Kubernetes also supports using `behavior` to customize scaling behaviors according to business scenarios, helping you
  achieve rapid or slow scaling or disable scaling. For more information about automatic scaling policies,
  see [Horizontal Pod Scaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).
