---
title: "API reference"
description: "StarRocks operator generated API reference docs"
draft: false
images: []
menu: "operator"
weight: 211
toc: true
---
> This page is automatically generated with `gen-crd-api-reference-docs`.
<p>Packages:</p>
<ul>
<li>
<a href="#starrocks.com%2fv1alpha1">starrocks.com/v1alpha1</a>
</li>
</ul>
<h2 id="starrocks.com/v1alpha1">starrocks.com/v1alpha1</h2>
<div>
</div>
Resource Types:
<ul></ul>
<h3 id="starrocks.com/v1alpha1.AutoScalingPolicy">AutoScalingPolicy
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksCnSpec">StarRocksCnSpec</a>)
</p>
<div>
<p>AutoScalingPolicy defines the auto scale</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>hpaPolicy</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.HPAPolicy">
HPAPolicy
</a>
</em>
</td>
<td>
<p>the policy of cn autoscale. operator use autoscaling v2.</p>
</td>
</tr>
<tr>
<td>
<code>minReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>the min numbers of target.</p>
</td>
</tr>
<tr>
<td>
<code>maxReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>the max numbers of target.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.ClusterPhase">ClusterPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
</p>
<div>
</div>
<h3 id="starrocks.com/v1alpha1.ConfigMapInfo">ConfigMapInfo
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksFeSpec">StarRocksFeSpec</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>configMapName</code><br/>
<em>
string
</em>
</td>
<td>
<p>the config info for start progress.</p>
</td>
</tr>
<tr>
<td>
<code>resolveKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>the config response key in configmap.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.HPAPolicy">HPAPolicy
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.AutoScalingPolicy">AutoScalingPolicy</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metrics</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#metricspec-v2-autoscaling">
[]Kubernetes autoscaling/v2.MetricSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics specifies how to scale based on a single metric</p>
</td>
</tr>
<tr>
<td>
<code>behavior</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#horizontalpodautoscalerbehavior-v2-autoscaling">
Kubernetes autoscaling/v2.HorizontalPodAutoscalerBehavior
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HorizontalPodAutoscalerBehavior configures the scaling behavior of the target</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.MemberPhase">MemberPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksBeStatus">StarRocksBeStatus</a>, <a href="#starrocks.com/v1alpha1.StarRocksCnStatus">StarRocksCnStatus</a>, <a href="#starrocks.com/v1alpha1.StarRocksFeStatus">StarRocksFeStatus</a>)
</p>
<div>
</div>
<h3 id="starrocks.com/v1alpha1.StarRocksBeSpec">StarRocksBeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
</p>
<div>
<p>StarRocksBeSpec defines the desired state of be.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Replicas is the number of desired be Pod. the default value=3
Optional</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image for a starrocks be deployment.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<p>serviceAccount for be access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>fsGroup</code><br/>
<em>
int64
</em>
</td>
<td>
<p>A special supplemental group that applies to all containers in a pod.
Some volume types allow the Kubelet to change the ownership of that volume
to be owned by the pod:</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>name of the starrocks be cluster.</p>
</td>
</tr>
<tr>
<td>
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksService">
StarRocksService
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service defines the template for the associated Kubernetes Service object.
the service for user access be.</p>
</td>
</tr>
<tr>
<td>
<code>limits</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Limits describes the maximum amount of compute resources allowed.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>requests</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for be configMap.</p>
</td>
</tr>
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksProbe">
StarRocksProbe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Probe defines the mode probe service in container is alive.</p>
</td>
</tr>
<tr>
<td>
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for be storage data and log.</p>
</td>
</tr>
<tr>
<td>
<code>ReplicaInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ReplicaInstance is the names of replica starrocksbe cluster.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) If specified, the pod&rsquo;s nodeSelector，displayName=&ldquo;Map of nodeSelectors to match when scheduling pods on nodes&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>beEnvVars</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>beEnvVars is a slice of environment variables that are added to the pods, the default is empty.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) Tolerations for scheduling pods onto some dedicated nodes</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksBeStatus">StarRocksBeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
</p>
<div>
<p>StarRocksBeStatus represents the status of starrocks be.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<p>the name of be service for fe find be instance.</p>
</td>
</tr>
<tr>
<td>
<code>failedInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>FailedInstances deploy failed instance of be.</p>
</td>
</tr>
<tr>
<td>
<code>creatingInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>CreatingInstances represents status in creating pods of be.</p>
</td>
</tr>
<tr>
<td>
<code>runningInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>RunningInstances represents status in running pods of be.</p>
</td>
</tr>
<tr>
<td>
<code>resourceNames</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>The statefulset names of be.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.MemberPhase">
MemberPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of be status. If be have one failed pod phase=failed,
also if be have one creating pod phase=creating, also if be all running phase=running, others unknown.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reason for the phase.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksCluster">StarRocksCluster
</h3>
<div>
<p>StarRocksCluster defines a starrocks cluster deployment.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksClusterSpec">
StarRocksClusterSpec
</a>
</em>
</td>
<td>
<p>Specification of the desired state of the starrocks cluster.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specify a Service Account for starRocksCluster use k8s cluster.
Deprecated: component use serviceAccount in own&rsquo;s field.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksFeSpec">
StarRocksFeSpec
</a>
</em>
</td>
<td>
<p>StarRocksFeSpec define fe configuration for start fe service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksBeSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksBeSpec">
StarRocksBeSpec
</a>
</em>
</td>
<td>
<p>StarRocksBeSpec define be configuration for start be service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn configuration for start cn service.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksClusterStatus">
StarRocksClusterStatus
</a>
</em>
</td>
<td>
<p>Most recent observed status of the starrocks cluster</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksClusterSpec">StarRocksClusterSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksCluster">StarRocksCluster</a>)
</p>
<div>
<p>StarRocksClusterSpec defines the desired state of StarRocksCluster</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specify a Service Account for starRocksCluster use k8s cluster.
Deprecated: component use serviceAccount in own&rsquo;s field.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksFeSpec">
StarRocksFeSpec
</a>
</em>
</td>
<td>
<p>StarRocksFeSpec define fe configuration for start fe service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksBeSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksBeSpec">
StarRocksBeSpec
</a>
</em>
</td>
<td>
<p>StarRocksBeSpec define be configuration for start be service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnSpec</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn configuration for start cn service.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksClusterStatus">StarRocksClusterStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksCluster">StarRocksCluster</a>)
</p>
<div>
<p>StarRocksClusterStatus defines the observed state of StarRocksCluster.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.ClusterPhase">
ClusterPhase
</a>
</em>
</td>
<td>
<p>Represents the state of cluster. the possible value are: running, failed, pending</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeStatus</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksFeStatus">
StarRocksFeStatus
</a>
</em>
</td>
<td>
<p>Represents the status of fe. the status have running, failed and creating pods.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksBeStatus</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksBeStatus">
StarRocksBeStatus
</a>
</em>
</td>
<td>
<p>Represents the status of be. the status have running, failed and creating pods.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnStatus</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksCnStatus">
StarRocksCnStatus
</a>
</em>
</td>
<td>
<p>Represents the status of cn. the status have running, failed and creating pods.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksCnSpec">StarRocksCnSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
</p>
<div>
<p>StarRocksCnSpec defines the desired state of cn.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>name of the starrocks cn cluster.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<p>serviceAccount for cn access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>fsGroup</code><br/>
<em>
int64
</em>
</td>
<td>
<p>A special supplemental group that applies to all containers in a pod.
Some volume types allow the Kubelet to change the ownership of that volume
to be owned by the pod:</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas is the number of desired cn Pod.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image for a starrocks cn deployment.</p>
</td>
</tr>
<tr>
<td>
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksService">
StarRocksService
</a>
</em>
</td>
<td>
<p>Service defines the template for the associated Kubernetes Service object.
the service for user access cn.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for cn configMap.</p>
</td>
</tr>
<tr>
<td>
<code>limits</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Limits describes the maximum amount of compute resources allowed.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>requests</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksProbe">
StarRocksProbe
</a>
</em>
</td>
<td>
<p>Probe defines the mode probe service in container is alive.</p>
</td>
</tr>
<tr>
<td>
<code>autoScalingPolicy</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.AutoScalingPolicy">
AutoScalingPolicy
</a>
</em>
</td>
<td>
<p>AutoScalingPolicy auto scaling strategy</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) If specified, the pod&rsquo;s nodeSelector，displayName=&ldquo;Map of nodeSelectors to match when scheduling pods on nodes&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>cnEnvVars</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>cnEnvVars is a slice of environment variables that are added to the pods, the default is empty.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) Tolerations for scheduling pods onto some dedicated nodes</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksCnStatus">StarRocksCnStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<p>the name of cn service for fe find cn instance.</p>
</td>
</tr>
<tr>
<td>
<code>failedInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>FailedInstances deploy failed cn pod names.</p>
</td>
</tr>
<tr>
<td>
<code>creatingInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>CreatingInstances in creating status cn pod names.</p>
</td>
</tr>
<tr>
<td>
<code>runningInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>RunningInstances in running status be pod names.</p>
</td>
</tr>
<tr>
<td>
<code>resourceNames</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>The statefulset names of be.</p>
</td>
</tr>
<tr>
<td>
<code>HpaName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The policy name of autoScale.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.MemberPhase">
MemberPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of cn status. If cn have one failed pod phase=failed,
also if cn have one creating pod phase=creating, also if cn all running phase=running, others unknown.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reason for the phase.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksFeSpec">StarRocksFeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
</p>
<div>
<p>StarRocksFeSpec defines the desired state of fe.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>name of the starrocks be cluster.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<p>serviceAccount for fe access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>fsGroup</code><br/>
<em>
int64
</em>
</td>
<td>
<p>A special supplemental group that applies to all containers in a pod.
Some volume types allow the Kubelet to change the ownership of that volume
to be owned by the pod:</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Replicas is the number of desired fe Pod, the number is 1,3,5</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image for a starrocks fe deployment..</p>
</td>
</tr>
<tr>
<td>
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksService">
StarRocksService
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service defines the template for the associated Kubernetes Service object.</p>
</td>
</tr>
<tr>
<td>
<code>limits</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Limits describes the maximum amount of compute resources allowed.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>requests</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for fe configMap.</p>
</td>
</tr>
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksProbe">
StarRocksProbe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Probe defines the mode probe service in container is alive.</p>
</td>
</tr>
<tr>
<td>
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for fe meta storage.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) If specified, the pod&rsquo;s nodeSelector，displayName=&ldquo;Map of nodeSelectors to match when scheduling pods on nodes&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>FeEnvVars</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>feEnvVars is a slice of environment variables that are added to the pods, the default is empty.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>(Optional) Tolerations for scheduling pods onto some dedicated nodes</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksFeStatus">StarRocksFeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
</p>
<div>
<p>StarRocksFeStatus represents the status of starrocks fe.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<p>the name of fe service exposed for user.</p>
</td>
</tr>
<tr>
<td>
<code>failedInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>FailedInstances failed fe pod names.</p>
</td>
</tr>
<tr>
<td>
<code>creatingInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>CreatingInstances in creating pod names.</p>
</td>
</tr>
<tr>
<td>
<code>runningInstances</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>RunningInstances in running status pod names.</p>
</td>
</tr>
<tr>
<td>
<code>resourceNames</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>ResourceNames the statefulset names of fe in v1alpha1 version.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.MemberPhase">
MemberPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of fe status. If fe have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if fe all running phase=running, others unknown.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Reason represents the reason of not running.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksProbe">StarRocksProbe
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksFeSpec">StarRocksFeSpec</a>)
</p>
<div>
<p>StarRocksProbe defines the mode for probe be alive.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
<p>Type identifies the mode of probe main container</p>
</td>
</tr>
<tr>
<td>
<code>initialDelaySeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Number of seconds after the container has started before liveness probes are initiated.
Default to 10 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>periodSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>How often (in seconds) to perform the probe.
Default to Kubernetes default (10 seconds). Minimum value is 1.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksService">StarRocksService
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksFeSpec">StarRocksFeSpec</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name assigned to service.</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#servicetype-v1-core">
Kubernetes core/v1.ServiceType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>type of service,the possible value for the service type are : ClusterIP, NodePort, LoadBalancer,ExternalName.
More info: <a href="https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types">https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types</a></p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#starrocks.com/v1alpha1.StarRocksServicePort">
[]StarRocksServicePort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports the components exposed ports and listen ports in pod.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StarRocksServicePort">StarRocksServicePort
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksService">StarRocksService</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the map about coming port and target port</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Port the pod is exposed on service.</p>
</td>
</tr>
<tr>
<td>
<code>containerPort</code><br/>
<em>
int32
</em>
</td>
<td>
<p>ContainerPort the service listen in pod.</p>
</td>
</tr>
<tr>
<td>
<code>nodePort</code><br/>
<em>
int32
</em>
</td>
<td>
<p>The easiest way to expose fe, cn or be is to use a Service of type <code>NodePort</code>.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1alpha1.StorageVolume">StorageVolume
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1alpha1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1alpha1.StarRocksFeSpec">StarRocksFeSpec</a>)
</p>
<div>
<p>StorageVolume defines additional PVC template for StatefulSets and volumeMount for pods that mount this PVC</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>name of a storage volume.</p>
</td>
</tr>
<tr>
<td>
<code>storageClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>storageClassName is the name of the StorageClass required by the claim.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1">https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1</a></p>
</td>
</tr>
<tr>
<td>
<code>storageSize</code><br/>
<em>
string
</em>
</td>
<td>
<p>StorageSize is a valid memory size type based on powers-of-2, so 1Mi is 1024Ki.
Supported units:Mi, Gi, GiB, Ti, Ti, Pi, Ei, Ex: <code>512Mi</code>.</p>
</td>
</tr>
<tr>
<td>
<code>mountPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>MountPath specify the path of volume mount.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
