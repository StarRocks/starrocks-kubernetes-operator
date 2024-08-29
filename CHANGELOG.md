# CHANGELOG

## [v1.9.8](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.8)

Release Notes for starrocks-kubernetes-operator v1.9.8

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.8. This version has the following
enhancements:

1. [Helm Chart] Support multiple data volumes on helm chart.
   [#578](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/578).
2. [Helm Chart] Use a reasonable volume size for logging for BE/CN. Because the logStorageSize field of the Helm Chart
   is modified. If the user uses the default value, the Operator will try to update the PVC part of the component's
   Statefulset. Statefulset does not allow this part of the configuration to be updated, so there is a case of tuning
   failure. The solution is to set it to the original default value, 1GB.
   [#579](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/579).
3. [Helm Chart] Support config override.
   [#580](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/580).
4. [Operator] Add update permission for starrocksclusters/finalizers resource for OpenShift.
   [#484](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/484)
5. [Operator] Support to disable probe.
   [#569](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/569)

We thank our community for their contributions and feedback. For a detailed list of changes and updates, please refer to
the GitHub repository. Happy deploying!

## [v1.9.7](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.7)

Release Notes for starrocks-kubernetes-operator v1.9.7

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.7. This version has the following
enhancements and bug fixes:

1. [Operator] Fix the problem that FE proxy will fail when FE external service is
   recreated [#557](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/557).
2. [Operator] Add loadBalancerSourceRanges field to support setting the source IP range for the load
   balancer [#551](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/551).
3. [Operator] Add topologySpreadConstraints field to support setting the topology spread constraints for
   pods [#546](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/546).
4. [Operator] Add CRD version information to CRD
   annotations [#544](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/544)
5. [Operator] Make it easy to configure the hostPath
   volume [#552](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/552)
6. [Chart] Add spill storage volume for BE and
   CN [#547](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/547)
7. [Chart] Remove JAVA_OPTS_FOR_JDK_9 and JAVA_OPTS_FOR_JDK_11 env from fe config, **this will cause the pods of FE to
   restart if you are using the default config of FE in values.yaml**
   [#542](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/542)
8. [Documentation] We have added or updated some documents to help users deploy and manage StarRocks clusters on
   Kubernetes. [#524](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/524) [#525](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/525) [#527](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/527) [#530](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/530) [#531](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/531) [#532](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/532) [#536](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/536) [#538](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/538).

We thank our community for their contributions and feedback. For a detailed list of changes and updates, please refer to
the GitHub repository. Happy deploying!

## [v1.9.6](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.6)

Release Notes for starrocks-kubernetes-operator v1.9.6

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.6. This version has the following
enhancements:

1. Support command and args[#516](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/516) in
   StarRocksCluster spec. You can easily customize the command and args if you are
   using helm chart.
   Below is a code snippet from values.yaml for your reference:
   ```yaml
    entrypoint:
      script: |
        #! /bin/bash
        echo "do something before start cn"
        exec /opt/starrocks/cn_entrypoint.sh $FE_SERVICE_NAME
   ```
2. Support ImagePullPolicy in StarRocksCluster
   spec[#514](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/514).
3. Support to use namespaced scope permission to deploy
   warehouses[#513](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/513).

## [v1.9.5](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.5)

Release Notes for starrocks-kubernetes-operator v1.9.5

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.5. This version brings a mix of
features and enhancements to further improve the deployment and management of StarRocks clusters on Kubernetes
environments.

### Feature

- [Feature] Support init Containers. User can add a k8s pod format initContainers.
  PR [#499](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/499) [#508](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/508)
- [Feature] Support sidecars. User can add a k8s pod format sidecars.
  PR [#461](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/461)
- [Feature] Support deploy multiple StarRocks clusters in one namespace.
  PR [#493](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/493) [#509](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/509)

### Enhancements

- [Enhancement] Support users to apply their own resources.
  PR [#496](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/496)
- [Enhancement] Support to customize securityContext for Operator Chart.
  PR [#495](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/495)
- [Enhancement] Support warehouse to be deployed in different namespace with StarRocks cluster.
  PR [#505](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/505)

### Maintenance

[Chore] This PR aims to standardize the format of YAML files in templates directory.
PR [#501](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/501)
[Chore] Upgrade Golang version to 1.22, and `sigs.k8s.io/controller-runtime` to v0.14.0.
PR [#497](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/497)

## [v1.9.4](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.4)

Release Notes for starrocks-kubernetes-operator v1.9.4

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.4. This version brings a mix of
enhancements and bug fixes to further improve the deployment and management of StarRocks clusters on Kubernetes
environments.

### Enhancements

- [Operator] We've refined the logic for detecting Kubernetes versions, enhancing compatibility.
  PR [#469](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/469)
- [Operator] Operator: Added checks for volume name and mount path when a default emptyDir volume is incorporated.
  PR [#464](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/464)
- [Chart] Helm chart allows the exposure of the DataDog `config.mode`.
  PR [#482](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/482)

### Bug Fixes

- [Operator] Addressed an issue by attempting to use a different version for deleting HPA (Horizontal Pod Autoscaler)
  again, aiming for a more reliable deletion process.
  PR [#468](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/468)
- [Chart] Make sure xx-config-hash is a valid string.
  PR [#480](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/480)
- [Chart] Introduced support for tolerations in the init-password job.
  PR [#463](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/463)

## [v1.9.3](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.3)

Release Notes for starrocks-kubernetes-operator v1.9.3

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.3. This version brings a mix of
enhancements and bug fixes to further improve the deployment and management of StarRocks clusters on Kubernetes
environments.

### Enhancements

- [Operator] Enhanced Lifecycle Management: Users can now define postStart and preStop lifecycle hooks for
  StarRocksCluster, enabling more control over the cluster's lifecycle events.
  PR [#456](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/456)
- [Operator] HostPath Support: This version introduces support for hostPath, allowing users to mount local directories
  to the StarRocksCluster. PR [#451](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/451)
- [Chart] Scoped Permissions: The operator's permissions can now be restricted to a single namespace, enhancing security
  by limiting the operator's scope to deploying StarRocks within a designated namespace.
  PR [#446](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/446)
- [Chart] Init-Password Job Enhancements: Support for annotations and specifying the image field for the init-password
  job has been added, allowing for greater customization.
  PR [#454](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/454)
- [Operator] Configuration Flexibility: The operator now supports be_http_port and be_port in the configuration.
  PR [#450](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/450)

### Bug Fixes

- [Operator] HPA Deletion Issue Resolved: A bug where the Horizontal Pod Autoscaler (HPA) was not properly deleted when
  the autoScalingPolicy field was removed has been fixed, ensuring clean and accurate scaling operations.
  PR [#444](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/444)
- [Chart] Init-Password Job Port Configuration: The query port can now be configured in the init-password job,
  addressing previous limitations and enhancing setup flexibility.
  PR [#455](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/455)

## [v1.9.2](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.2)

Release Notes for starrocks-kubernetes-operator v1.9.2

We are excited to announce the release of StarRocks Kubernetes Operator v1.9.2. This version brings a mix of new
features, enhancements, documentation updates, and bug fixes to further improve the deployment and management of
StarRocks clusters on Kubernetes environments.

### What's New

- **[chart]** Add Datadog profiling to SR Helm to enhance monitoring
  capabilities. [#437](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/437)
- **[chart]** Add `starrocksCluster.componentValues` to define some values uniformly to streamline
  configurations. [#425](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/425)
- **[operator]** Add k8s event information during the Operator reconciliation for better debugging and operational
  insights. [#391](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/391)

### Enhancements

- **[operator]** Prefer to use containerPort to export service node
  port. [#421](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/421)
- **[chart]** Upgrade ServiceMonitor
    1. In order to
       support [StarRocks Grafana Dashboard](https://github.com/StarRocks/starrocks/blob/main/extra/grafana/kubernetes/StarRocks-Overview-kubernetes-3.0.json),
       `app_starrocks_ownerreference_name` or `app_kubernetes_io_component labels` were added to `up` metrics by
       ServiceMonitor. [#433](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/433)
    2. Users can now add labels on ServiceMonitor for more flexible monitoring
       configuration. [#432](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/432)
- **[chart]** Allow users to specify mount paths, providing more customization options for
  deployments. [#428](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/428)
- **[chart]** Eliminate Helm warnings caused by differing types of feEnvVars, beEnvVars, and cnEnvVars, improving
  the deployment
  experience. [#434](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/434) [#396](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/396)
- **[chart]** Complete imagePullSecrets and affinity fields for helm charts to enhance deployment flexibility and
  scheduling. [#417](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/417)

### Bug Fixes

- **[chart]** Allowed to remove some resources limit, like CPU, providing more flexibility in resource
  management. [#426](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/426)
- **[chart]** Upgraded the version of Golang and libraries to fix some vulnerabilities, improving
  security. [#415](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/415)

### Documentation

- Added `doc/least_permission_to_deploy_starrocks_howto.md` to help users deploy with minimal
  permissions. [#431](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/431)
- Added more details about Prometheus integration for better monitoring
  setup. [#427](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/427)
- Removed the 1.7.0 and 1.6.x version in
  index.yaml. [#435](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/435)
- Documentation improvements including moving and adding pull request templates for better contribution
  guidelines. [#410](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/410) [#409](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/409)

We thank our community for their contributions and feedback. For a detailed list of changes and updates, please refer to
the GitHub repository. Happy deploying!

## [v1.9.1](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.1)

Release Notes for starrocks-kubernetes-operator v1.9.1

We are thrilled to announce the release of StarRocks Kubernetes Operator v1.9.1. This release introduces several
enhancements and bug fixes to improve the user experience of deploying and managing StarRocks
clusters on Kubernetes.

### What's New

There are enhancements in this release.

1. **[chart]** When you set logStorageSize to 0, the operator will not create PVC for log
   storage [#398](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/398).
2. **[operator]** Checked volumes and mount paths to avoid duplication
   error [#388](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/388)
3. **[operator]** Disabled FE scale to 1 [#394](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/394)
4. **[operator]** Supported use of map to define feEnvVars/beEnvVars/cnEnvVars to merge on multiple values
   files [#396](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/396)
5. **[operator]** exposed `spec.containers.securityContext.capabilities` for user to customize the capabilities of
   containers. [#404](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/404)

### BugFix

1. **[operator]** Supported to update service annotations
   fields [#402](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/402) [#399](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/399)
2. **[operator]** Switched to using patch method instead of update method to modify statefulset and
   deployment [#397](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/397).
   This solves the problem that when CN + HPA is enabled, upgrading CN will cause all CN pods to be terminated and
   restarted.
3. **[operator]** Switched to using Patch method instead of Update to modify service
   object [#387](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/387). This solves the
   problem that when cloud provider is used, the service object will be modified by cloud provider, and the operator
   will overwrite the modification.
4. **[operator]** Considered Cn Replicas when calculating component
   hash [#392](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/392)
5. **[chart]** Corrected typo in storageSpec [#385](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/385)

## [v1.9.0](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.9.0)

Release Notes for starrocks-kubernetes-operator v1.9.0

We are thrilled to announce the release of StarRocks Kubernetes Operator v1.9.0. This release introduces several
enhancements, bug fixes, and documentation updates to improve the user experience of deploying and managing StarRocks
clusters on Kubernetes.

### What's New

1. [Feature] Add StarRocksWarehouse CRD to support StarRocks Warehouse
   Feature [#323](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/323). Note: warehouse is an
   enterprise feature for StarRocks.
2. [Enhancement] Use StarRocksCluster State to log errors when subController apply
   failed [#359](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/359): We have improved error
   logging by using the StarRocksCluster status.State when the subController apply operation fails.
3. [Enhancement] Support to mount emptyDir in
   storageVolumes [#324](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/324): We have added support to
   mount emptyDir in
   storageVolumes.

### BugFix

1. [BugFix] We have fixed an issue where the cluster status phase was not in sync with the
   component.[#380](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/380)
2. [BugFix] We have fixed an issue where the HPA was not removed when the autoScalingPolicy was
   removed.[#379](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/379)
3. [BugFix] We have fixed an issue where the HPA resource was not removed when the CN spec was
   removed.[#357](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/357)

### Maintenance

1. We have improved the setup of the Kubernetes environment for unit tests by using
   setup-envtest. [#347](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/347)
2. We have added unit tests and ensured that the code coverage is at least
   65%. [#354](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/354)
3. We have updated the script to generate the API reference
   documentation.[#350](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/350)
4. We have switched to using zap as the logger for better logging
   capabilities.[#341](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/341)

We encourage you to update to this new version and benefit from these improvements. As always, your feedback is very
welcome.

## [v1.8.8](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.8)

Release Notes for starrocks-kubernetes-operator v1.8.8

We are pleased to announce the release of starrocks-kubernetes-operator v1.8.8.

### BugFix

1. When add annotations by `spec.StarRocksFeSpec/StarRocksBeSpec/StarRocksCnSpec.service` field, Operator should not
   annotate on search(internal) service.

We encourage you to update to this new version and benefit from these improvements. As always, your feedback is very
welcome.

## [v1.8.7](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.7)

Release Notes for starrocks-kubernetes-operator v1.8.7

We are pleased to announce the release of starrocks-kubernetes-operator v1.8.7. This release includes several updates
and improvements that enhance the functionality and usability of the StarRocks Kubernetes Operator.

### Enhancement

1. Added failure seconds for liveness and readiness. When StarRocks is under heavy load, the default Liveness Probe and
   Readiness Probe may fail, causing the container to restart. This update mitigates this issue by adding failure
   seconds for liveness and readiness.

### Maintenance

1. [Documentation] Updated README.md and README_ZH-CN.md. This update provides more accurate and comprehensive
   information about StarRocks Operator.
2. [Documentation]
   Added [local_installation_how_to.md](https://github.com/StarRocks/starrocks-kubernetes-operator/blob/main/doc/local_installation_how_to.md)
   for users. This guide provides step-by-step instructions for local installation of the Operator and StarRocks
   Cluster. And it also provides a script to help users install the Operator and StarRocks Cluster locally.

We encourage you to update to this new version and benefit from these improvements. As always, your feedback is very
welcome.

## [v1.8.6](https://github.com/StarRocks/starrocks-kubernetes-operator/releases/tag/v1.8.6)

Release Notes for starrocks-kubernetes-operator v1.8.6

We are pleased to announce the release of starrocks-kubernetes-operator v1.8.6. This release includes several bug fixes
and enhancements. Here are the key updates:

### Bug Fixes

1. Fix the problem of nginx sending the request body to FE when redirecting the stream load request, which may cause
   the stream load to fail. [#303](https://github.com/StarRocks/starrocks-kubernetes-operator/pull/303)

### Maintenance

1. [Documentation] add doc/load_data_using_stream_load.md. This document introduces how to load data from outside the
   k8s network to StarRocks through fe-proxy.
2. [Documentation] update change_root_password_howto.md. This document adds the steps of how to update the
   root password through Helm Chart.
3. [Chore] Add GitHub Actions to add label on issue and PR. This chore improves the project quality by adding necessary
   labels to issues and PRs.

We encourage you to update to this new version and benefit from these improvements. As always, your feedback is very
welcome.

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
   running: `helm repo update; helm upgrade <release-name> starrocks/kube-starrocks -f values-v1.8.0.yaml`

### How to Install

To install StarRocks Kubernetes Operator for the first time, please follow these steps:

1. Add the StarRocks Helm repository by
   running: `helm repo add starrocks https://starrocks.github.io/starrocks-kubernetes-operator`, then
   execute `helm repo update`.
2. Install the kube-starrocks chart by
   running: `helm install <release-name> starrocks/kube-starrocks -f values.yaml`
3. Alternatively, you can install the operator and starrocks subcharts separately by running. Install
   operator: `helm install <release-name> starrocks/operator -f operator-values.yaml`. Install
   starrocks: `helm install <starrocks-release-name> starrocks/starrocks -f starrocks-values.yaml`

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
