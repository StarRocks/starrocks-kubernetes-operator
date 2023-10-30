# CHANGELOG

## [v1.8.5](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.5)

Release Notes for starrocks-kubernetes-operator v1.8.5

We are pleased to announce the release of starrocks-kubernetes-operator v1.8.5. This release includes several bug fixes
and enhancements. Here are the key updates:

### What's New

1. **[Chart] Service Account Customization**: The operator now supports to provide the ability to add custom
   annotations and labels. **The previous `operator.global.rbac.serviceAccountName` field is no longer supported.**
2. **[Operator] Explicit Protocol Selection for Istio:** To provide additional functionality such as routing and rich
   metrics, Istio needs to determine the protocol of the traffic. This is now achieved by automatically adding the
   appProtocol field. This enhancement is particularly important for "server-first" protocols like MySQL, which are
   incompatible with automatic protocol detection and may cause connection failures.

### Bug Fixes

1. Fixed an issue that occurred when `starrocks.initPassword.enabled` is true and the value
   of `starrocks.starrocksCluster.name` is set. The FE service domain name would follow the set value, while the initpwd
   pod would still use the `starrocks.nameOverride` value to compose the FE service domain name. This fix ensures
   consistent hostname usage.

### Notes for Deployed Users

If `starrocks.starrocksCluster.name` is not set, the result of helm template will remain the same as before. If
`starrocks.starrocksCluster.name` is set and is different from the value of `starrocks.nameOverride`, the old configmaps
for FE, BE, and CN will be deleted. New configmaps with the new name for FE, BE, and CN will be created. **This may
result in the restart of FE/BE/CN pods.**

We encourage you to update to this new version and benefit from these improvements. As always, your feedback is very
welcome.

## [v1.8.4](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.4)

Release Notes for starrocks-kubernetes-operator v1.8.4

We are excited to announce the release of starrocks-kubernetes-operator v1.8.4, a Kubernetes operator for StarRocks.
This release brings several new features, bug fixes, and documentation improvements.

### What's New

1. **[Feature] add golangci-lint GitHub Action:** This feature enables the golangci-lint tool to check the code quality
   and style of the operator project on every pull request.
2. **[Feature] allow you to configure the terminationGracePeriodSeconds in StarRocksCluster CRD:** This feature allows
   you to specify how long to wait before forcefully terminating a pod when deleting or updating a StarRocksCluster
   resource.
3. **[Feature] add storage fields for starrocksCnSpec in values.yaml:** This feature allows you to configure the storage
   class and size for the CN nodes in the StarRocksCluster.
4. **[Feature] integration with Prometheus by ServiceMonitor CRD:** This feature allows you to monitor the metrics of
   StarRocks cluster by using Prometheus and ServiceMonitor CRD.
5. **[Feature] support startupProbeFailureSeconds fields in StarRocksCluster CRD:** This feature allows you to configure
   the startup probe failure threshold for the pods in the StarRocksCluster resource.
6. **[Feature] facilitate the configuration of environmental variables for the operator pod:** for instance, one may
   designate KUBE_STARROCKS_UNSUPPORTED_ENVS to eliminate environments incompatible with the kubernetes cluster.

### Bug Fixes

1. [Bugfix] This bugfix solves the problem that FE Proxy cannot handle stream load requests correctly when there are
   multiple FE pods in the StarRocks cluster.

### Maintenance

1. **[Doc] Set up StarRocks locally:** This document provides a guide on how to set up a local StarRocks cluster.
2. **[Doc] Establish a comprehensive StarRocks cluster, encompassing all available features.** For additional examples
   concerning StarRocksCluster,
   please refer to: https://github.com/StarRocks/starrocks-kubernetes-operator/tree/main/examples/starrocks.
3. **[Doc] Augment the instructional material on StarRocks utilization**, encompassing topics such as '
   logging_and_related_configurations_howto.md' and 'mount_external_configmaps_or_secrets_howto.md'.
   For a holistic view of available guides,
   please refer to: https://github.com/StarRocks/starrocks-kubernetes-operator/tree/main/doc

## [v1.8.3](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.3)

This is a minor release of StarRocks Kubernetes Operator.

### What's New

1. **[operator] Set proxy_read_timeout 600s:**  We have set proxy_read_timeout to 600s in nginx.conf, because the
   default value is 60s, which may cause timeout.
2. **[chart] Add JAVA_OPTS_FOR_JDK_11 in FE config:** We have added JAVA_OPTS_FOR_JDK_11 in FE config, because the
   default value is not suitable for JDK 11. **If you used the default config of FE in values.yaml, Upgrading to v1.8.3
   will cause the pods of FE to restart.**
3. **[chart] Allow user to specify serviceAccount for operator:** By default the operator chart will create a
   serviceAccount for the operator, named starrocks. But if you want to use an existing serviceAccount, you can specify
   it in values.yaml.
4. **[operator] Support `Ports` for feProxy component:** We have supported `Ports` for feProxy component, allowing
   users to specify the nodePort of feProxy service.
5. **[operator] Add namespace flag:** It makes operator watch resources in a single namespace instead of all namespaces
   in the cluster. In most cases, you should not set this value. If your kubernetes cluster manages too many nodes, and
   operator watching all namespaces use too many memory resources, you can set this value.

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
