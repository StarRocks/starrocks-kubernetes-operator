# CHANGELOG

## [v1.8.2](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.2)

This is a minor release of StarRocks Kubernetes Operator, a project that aims to provide a Kubernetes-native way to
deploy and manage StarRocks clusters on Kubernetes.

### What's New

1. **[operator.yaml] Change the resource of operator:** We have changed the default resource limit and request of
   operator pod in case of OOM.
2. **[operator] Remove the watch for the HPA resource:** We have removed the watch for the HPA resource, because
   operator can not make sure what version of HPA is used in the cluster.

## [v1.8.1](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.1)

This is a minor release of StarRocks Kubernetes Operator, a project that aims to provide a Kubernetes-native way to
deploy and manage StarRocks clusters on Kubernetes.

### What's New

1. **[operator] Watch HPA:** The operator will watch the HPA resources and once the HPA is missing, the operator will
   try to create it again.
2. **[operator] Select the appropriate HPA version:** The operator will select the appropriate HPA version based on the
   Kubernetes version, ensuring compatibility and stability.
3. **[operator] Reduce incorrect modification of replicas:** The operator will remove the replicas field of the CN
   statefulset when HPA is enabled, avoiding conflicts between HPA and statefulset controller.
4. **[operator] Support subpath:** The operator will support subpath for configmaps and secrets, allowing users to mount
   specific files or directories from these resources.
5. **[operator] Remove unnecessary resources:** The operator will remove the related Kubernetes resources when the
   BeSpec or CnSpec of the StarRocks cluster is deleted, ensuring a clean and consistent state of the cluster.
6. **[operator] Support ports fields:** The operator will support ports field for the StarRocks cluster, allowing users
   to customize the ports of the services.

## [v1.8.0](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.0)

We are excited to announce the release of StarRocks Kubernetes Operator v1.8.0. This release brings some new features
and bug fixes to improve the user experience of deploying and managing StarRocks clusters on Kubernetes.

### What's New

- **[operator] CRD changes:** We have modified the CRD definition in v1.8.0. Before you use helm upgrade to upgrade from
  a previous version, you need to manually update the CRD definition.
- **[operator] Non-root mode:** We have added a runAsNonRoot field to support running StarRocks pods as non-root users.
  This can enhance the security of your StarRocks clusters.
- **[chart] Subchart support:** We have split the kube-starrocks chart into two subcharts: operator and starrocks.
  Installing kube-starrocks is equivalent to installing both operator and starrocks subcharts, and uninstalling
  kube-starrocks is equivalent to uninstalling both operator and starrocks subcharts. If you want more flexibility in
  managing your StarRocks clusters, you can install operator and starrocks subcharts separately.
- **[chart] Values migration tool:** We have provided a tool to migrate your values.yaml file from a previous version to
  v1.8.0. You can use the migrate-chart-value command to upgrade your values.yaml file.
- **[chart] Multiple cluster support:** We have enabled support for **deploying multiple StarRocks clusters in different
  namespaces in one Kubernetes cluster.** You just need to install the starrocks chart in different namespaces.
- **[operator] FE proxy feature:** We have added the FE proxy feature to allow external clients and data import tools to
  access StarRocks clusters in Kubernetes. For example, you can use the STREAM LOAD syntax to import data into StarRocks
  clusters.
- **[chart] Datadog integration:** We have integrated with Datadog to provide metrics and logs for StarRocks clusters.
  You can enable this feature by setting the datadog related fields in your values.yaml file.
- **[chart] Cluster password initialization:** We have added the ability to initialize the password of root in your
  StarRocks cluster during installation. Note that this only works for helm install, can't use it in helm upgrade

### How to Upgrade

To upgrade from a previous version of StarRocks Kubernetes Operator, please follow these steps:

1. Update the CRD definition by
   running: `kubectl apply -f https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.8.0/starrocks.com_starrocksclusters.yaml`
2. Migrate your values.yaml file by
   running: `migrate-chart-value --input values.yaml --target-version v1.8.0 --output ./values-v1.8.0.yaml`
3. Upgrade your Helm release by
   running: `helm repo update; helm upgrade <release-name> starrocks-community/kube-starrocks -f values-v1.8.0.yaml`

### How to Install

To install StarRocks Kubernetes Operator for the first time, please follow these steps:

1. Add the StarRocks Helm repository by
   running: `helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator`, then
   execute `helm repo update`.
2. Install the kube-starrocks chart by
   running: `helm install <release-name> starrocks-community/kube-starrocks -f values.yaml`
3. Alternatively, you can install the operator and starrocks subcharts separately by running. Install
   operator: `helm install <release-name> starrocks-community/operator -f operator-values.yaml`. Install
   starrocks: `helm install <starrocks-release-name> starrocks-community/starrocks -f starrocks-values.yaml`

## [v1.7.1](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.7.1)

This is a minor release of StarRocks Kubernetes Operator, a project that aims to provide a Kubernetes-native way to
deploy and manage StarRocks clusters on Kubernetes.

### What's New

1. **[operator] Reconcile when service type changed:** The operator will reconcile when the service type of the
   StarRocks cluster is updated.
2. **[chart] Init password when install chart:** will only be executed when helm install is performed.
3. **[chart] Support artifacthub.io: ** The project has added some files and configurations to support artifacthub.io, a
   website that provides a centralized place to discover and distribute Kubernetes packages.

## [v1.7.0](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.7.0)

We are excited to announce the release of StarRocks Kubernetes Operator 1.7.0! This is a continous improvement of
operator open source edition, with various bug fixes, documentation, and improvements.

### What's New

1. **[operator] Update probes:** PODS Probe is switched from TCP/9030 to HTTP/8030, eliminating annoying error logs in
   FE's log.
2. **[operator] Support to config toleration, node selector and annotation for operator pod:** It allows users to
   customize the scheduling and metadata of the operator pod.
3. **[operator] Support to add annotations on Kubernetes service:** It allows users to add some metadata to the service.
4. **[operator] Support customized pod scheduler:** It allows users to specify a scheduler name for the pods.
5. **[chart] Mounting additional ConfigMaps and Secrets:** It allows mounting additional ConfigMaps and Secrets into
   StarRocks PODs.
6. **[chart] Support init password:** It Supports specifying init password for root user in helm chart.
7. **[chart] Support to ensure pods restart when configmap changes:** It allows users to apply the configuration
   changes without manual intervention.
