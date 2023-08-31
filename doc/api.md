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
<a href="#starrocks.com%2fv1">starrocks.com/v1</a>
</li>
</ul>
<h2 id="starrocks.com/v1">starrocks.com/v1</h2>
<div>
</div>
Resource Types:
<ul></ul>
<h3 id="starrocks.com/v1.AnnotationOperationValue">AnnotationOperationValue
(<code>string</code> alias)</h3>
<div>
<p>AnnotationOperationValue present the operation for fe, cn, be.</p>
</div>
<h3 id="starrocks.com/v1.AutoScalerVersion">AutoScalerVersion
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.AutoScalingPolicy">AutoScalingPolicy</a>, <a href="#starrocks.com/v1.HorizontalScaler">HorizontalScaler</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;v1&#34;</p></td>
<td><p>AutoScalerV1 the cn service use v1 autoscaler. Reference to <a href="https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/">https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/</a></p>
</td>
</tr><tr><td><p>&#34;v2&#34;</p></td>
<td><p>AutoScalerV2 the cn service use v2. Reference to  <a href="https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/">https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/</a></p>
</td>
</tr><tr><td><p>&#34;v2beta2&#34;</p></td>
<td><p>AutoScalerV2Beta2 the cn service use v2beta2. Reference to  <a href="https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/">https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/</a></p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.AutoScalingPolicy">AutoScalingPolicy
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksCnSpec">StarRocksCnSpec</a>)
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
<a href="#starrocks.com/v1.HPAPolicy">
HPAPolicy
</a>
</em>
</td>
<td>
<p>the policy of autoscaling. operator use autoscaling v2.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
<a href="#starrocks.com/v1.AutoScalerVersion">
AutoScalerVersion
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>version represents the autoscaler version for cn service. only support v1,v2beta2,v2</p>
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
<p>MinReplicas is the lower limit for the number of replicas to which the autoscaler
can scale down. It defaults to 1 pod.</p>
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
<p>MaxReplicas is the upper limit for the number of pods that can be set by the autoscaler;
cannot be smaller than MinReplicas.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ComponentPhase">ComponentPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksComponentStatus">StarRocksComponentStatus</a>)
</p>
<div>
<p>ComponentPhase represent the component phase about be, cn, be. The possible value for component phase are: reconciling, failed, running.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;failed&#34;</p></td>
<td><p>ComponentFailed have at least one service failed.</p>
</td>
</tr><tr><td><p>&#34;reconciling&#34;</p></td>
<td><p>ComponentReconciling the starrocks have component in starting.</p>
</td>
</tr><tr><td><p>&#34;running&#34;</p></td>
<td><p>ComponentRunning all components runs available.</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.ConfigMapInfo">ConfigMapInfo
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksLoadSpec">StarRocksLoadSpec</a>)
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
<h3 id="starrocks.com/v1.ConfigMapReference">ConfigMapReference
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksComponentSpec">StarRocksComponentSpec</a>)
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
<p>This must match the Name of a ConfigMap or Secret in the same namespace, and
the length of name must not more than 50 characters.</p>
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
<p>Path within the container at which the volume should be mounted.  Must
not contain &lsquo;:&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>subPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>SubPath within the volume from which the container&rsquo;s volume should be mounted.
Defaults to &ldquo;&rdquo; (volume&rsquo;s root).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ContainerResourceMetricSource">ContainerResourceMetricSource
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>ContainerResourceMetricSource indicates how to scale on a resource metric known to
Kubernetes, as specified in requests and limits, describing each pod in the
current scale target (e.g. CPU or memory).  The values will be averaged
together before being compared to the target.  Such metrics are built in to
Kubernetes, and have special scaling options on top of those available to
normal per-pod metrics using the &ldquo;pods&rdquo; source.  Only one &ldquo;target&rdquo; type
should be set.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcename-v1-core">
Kubernetes core/v1.ResourceName
</a>
</em>
</td>
<td>
<p>name is the name of the resource in question.</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="#starrocks.com/v1.MetricTarget">
MetricTarget
</a>
</em>
</td>
<td>
<p>target specifies the target value for the given metric</p>
</td>
</tr>
<tr>
<td>
<code>container</code><br/>
<em>
string
</em>
</td>
<td>
<p>container is the name of the container in the pods of the scaling target</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.CrossVersionObjectReference">CrossVersionObjectReference
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.ObjectMetricSource">ObjectMetricSource</a>)
</p>
<div>
<p>CrossVersionObjectReference contains enough information to let you identify the referred resource.</p>
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
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<p>Kind of the referent; More info: <a href="https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds&quot;">https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds&rdquo;</a></p>
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
<p>Name of the referent; More info: <a href="http://kubernetes.io/docs/user-guide/identifiers#names">http://kubernetes.io/docs/user-guide/identifiers#names</a></p>
</td>
</tr>
<tr>
<td>
<code>apiVersion</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>API version of the referent</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ExternalMetricSource">ExternalMetricSource
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>ExternalMetricSource indicates how to scale on a metric not associated with
any Kubernetes object (for example length of queue in cloud
messaging service, or QPS from loadbalancer running outside of cluster).</p>
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
<code>metric</code><br/>
<em>
<a href="#starrocks.com/v1.MetricIdentifier">
MetricIdentifier
</a>
</em>
</td>
<td>
<p>metric identifies the target metric by name and selector</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="#starrocks.com/v1.MetricTarget">
MetricTarget
</a>
</em>
</td>
<td>
<p>target specifies the target value for the given metric</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.HPAPolicy">HPAPolicy
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.AutoScalingPolicy">AutoScalingPolicy</a>)
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
<a href="#starrocks.com/v1.MetricSpec">
[]MetricSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics specifies how to scale based on a single metric
the struct copy from k8s.io/api/autoscaling/v2beta2/types.go. the redundancy code will hide the restriction about
HorizontalPodAutoscaler version and kubernetes releases matching issue.
the splice will have unsafe.Pointer convert, so be careful to edit the struct fields.</p>
</td>
</tr>
<tr>
<td>
<code>behavior</code><br/>
<em>
<a href="#starrocks.com/v1.HorizontalPodAutoscalerBehavior">
HorizontalPodAutoscalerBehavior
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HorizontalPodAutoscalerBehavior configures the scaling behavior of the target.
the struct copy from k8s.io/api/autoscaling/v2beta2/types.go. the redundancy code will hide the restriction about
HorizontalPodAutoscaler version and kubernetes releases matching issue.
the</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.HPAScalingPolicy">HPAScalingPolicy
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HPAScalingRules">HPAScalingRules</a>)
</p>
<div>
<p>HPAScalingPolicy is a single policy which must hold true for a specified past interval.</p>
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
<a href="#starrocks.com/v1.HPAScalingPolicyType">
HPAScalingPolicyType
</a>
</em>
</td>
<td>
<p>Type is used to specify the scaling policy.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Value contains the amount of change which is permitted by the policy.
It must be greater than zero</p>
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
<p>PeriodSeconds specifies the window of time for which the policy should hold true.
PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.HPAScalingPolicyType">HPAScalingPolicyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HPAScalingPolicy">HPAScalingPolicy</a>)
</p>
<div>
<p>HPAScalingPolicyType is the type of the policy which could be used while making scaling decisions.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Percent&#34;</p></td>
<td><p>PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
the current number of pods.</p>
</td>
</tr><tr><td><p>&#34;Pods&#34;</p></td>
<td><p>PodsScalingPolicy is a policy used to specify a change in absolute number of pods.</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.HPAScalingRules">HPAScalingRules
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HorizontalPodAutoscalerBehavior">HorizontalPodAutoscalerBehavior</a>)
</p>
<div>
<p>HPAScalingRules configures the scaling behavior for one direction.
These Rules are applied after calculating DesiredReplicas from metrics for the HPA.
They can limit the scaling velocity by specifying scaling policies.
They can prevent flapping by specifying the stabilization window, so that the
number of replicas is not set instantly, instead, the safest value from the stabilization
window is chosen.</p>
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
<code>stabilizationWindowSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StabilizationWindowSeconds is the number of seconds for which past recommendations should be
considered while scaling up or scaling down.
StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
If not set, use the default values:
- For scale up: 0 (i.e. no stabilization is done).
- For scale down: 300 (i.e. the stabilization window is 300 seconds long).</p>
</td>
</tr>
<tr>
<td>
<code>selectPolicy</code><br/>
<em>
<a href="#starrocks.com/v1.ScalingPolicySelect">
ScalingPolicySelect
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>selectPolicy is used to specify which policy should be used.
If not set, the default value MaxPolicySelect is used.</p>
</td>
</tr>
<tr>
<td>
<code>policies</code><br/>
<em>
<a href="#starrocks.com/v1.HPAScalingPolicy">
[]HPAScalingPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>policies is a list of potential scaling polices which can be used during scaling.
At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.HorizontalPodAutoscalerBehavior">HorizontalPodAutoscalerBehavior
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HPAPolicy">HPAPolicy</a>)
</p>
<div>
<p>HorizontalPodAutoscalerBehavior configures the scaling behavior of the target
in both Up and Down directions (scaleUp and scaleDown fields respectively).</p>
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
<code>scaleUp</code><br/>
<em>
<a href="#starrocks.com/v1.HPAScalingRules">
HPAScalingRules
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>scaleUp is scaling policy for scaling Up.
If not set, the default value is the higher of:
* increase no more than 4 pods per 60 seconds
* double the number of pods per 60 seconds
No stabilization is used.</p>
</td>
</tr>
<tr>
<td>
<code>scaleDown</code><br/>
<em>
<a href="#starrocks.com/v1.HPAScalingRules">
HPAScalingRules
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>scaleDown is scaling policy for scaling Down.
If not set, the default value is to allow to scale down to minReplicas pods, with a
300 second stabilization window (i.e., the highest recommendation for
the last 300sec is used).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.HorizontalScaler">HorizontalScaler
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksCnStatus">StarRocksCnStatus</a>)
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
<p>the horizontal scaler name</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
<a href="#starrocks.com/v1.AutoScalerVersion">
AutoScalerVersion
</a>
</em>
</td>
<td>
<p>the horizontal version.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.MetricIdentifier">MetricIdentifier
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.ExternalMetricSource">ExternalMetricSource</a>, <a href="#starrocks.com/v1.ObjectMetricSource">ObjectMetricSource</a>, <a href="#starrocks.com/v1.PodsMetricSource">PodsMetricSource</a>)
</p>
<div>
<p>MetricIdentifier defines the name and optionally selector for a metric</p>
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
<p>name is the name of the given metric</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>selector is the string-encoded form of a standard kubernetes label selector for the given metric
When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping.
When unset, just the metricName will be used to gather metrics.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.MetricSourceType">MetricSourceType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>MetricSourceType indicates the type of metric.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;ContainerResource&#34;</p></td>
<td><p>ContainerResourceMetricSourceType is a resource metric known to Kubernetes, as
specified in requests and limits, describing a single container in each pod in the current
scale target (e.g. CPU or memory).  Such metrics are built in to
Kubernetes, and have special scaling options on top of those available
to normal per-pod metrics (the &ldquo;pods&rdquo; source).</p>
</td>
</tr><tr><td><p>&#34;External&#34;</p></td>
<td><p>ExternalMetricSourceType is a global metric that is not associated
with any Kubernetes object. It allows autoscaling based on information
coming from components running outside of cluster
(for example length of queue in cloud messaging service, or
QPS from loadbalancer running outside of cluster).</p>
</td>
</tr><tr><td><p>&#34;Object&#34;</p></td>
<td><p>ObjectMetricSourceType is a metric describing a kubernetes object
(for example, hits-per-second on an Ingress object).</p>
</td>
</tr><tr><td><p>&#34;Pods&#34;</p></td>
<td><p>PodsMetricSourceType is a metric describing each pod in the current scale
target (for example, transactions-processed-per-second).  The values
will be averaged together before being compared to the target value.</p>
</td>
</tr><tr><td><p>&#34;Resource&#34;</p></td>
<td><p>ResourceMetricSourceType is a resource metric known to Kubernetes, as
specified in requests and limits, describing each pod in the current
scale target (e.g. CPU or memory).  Such metrics are built in to
Kubernetes, and have special scaling options on top of those available
to normal per-pod metrics (the &ldquo;pods&rdquo; source).</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.MetricSpec">MetricSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HPAPolicy">HPAPolicy</a>)
</p>
<div>
<p>MetricSpec specifies how to scale based on a single metric
(only <code>type</code> and one other matching field should be set at once).</p>
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
<a href="#starrocks.com/v1.MetricSourceType">
MetricSourceType
</a>
</em>
</td>
<td>
<p>type is the type of metric source.  It should be one of &ldquo;ContainerResource&rdquo;, &ldquo;External&rdquo;,
&ldquo;Object&rdquo;, &ldquo;Pods&rdquo; or &ldquo;Resource&rdquo;, each mapping to a matching field in the object.
Note: &ldquo;ContainerResource&rdquo; type is available on when the feature-gate
HPAContainerMetrics is enabled</p>
</td>
</tr>
<tr>
<td>
<code>object</code><br/>
<em>
<a href="#starrocks.com/v1.ObjectMetricSource">
ObjectMetricSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>object refers to a metric describing a single kubernetes object
(for example, hits-per-second on an Ingress object).</p>
</td>
</tr>
<tr>
<td>
<code>pods</code><br/>
<em>
<a href="#starrocks.com/v1.PodsMetricSource">
PodsMetricSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>pods refers to a metric describing each pod in the current scale target
(for example, transactions-processed-per-second).  The values will be
averaged together before being compared to the target value.</p>
</td>
</tr>
<tr>
<td>
<code>resource</code><br/>
<em>
<a href="#starrocks.com/v1.ResourceMetricSource">
ResourceMetricSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>resource refers to a resource metric (such as those specified in
requests and limits) known to Kubernetes describing each pod in the
current scale target (e.g. CPU or memory). Such metrics are built in to
Kubernetes, and have special scaling options on top of those available
to normal per-pod metrics using the &ldquo;pods&rdquo; source.</p>
</td>
</tr>
<tr>
<td>
<code>containerResource</code><br/>
<em>
<a href="#starrocks.com/v1.ContainerResourceMetricSource">
ContainerResourceMetricSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>container resource refers to a resource metric (such as those specified in
requests and limits) known to Kubernetes describing a single container in
each pod of the current scale target (e.g. CPU or memory). Such metrics are
built in to Kubernetes, and have special scaling options on top of those
available to normal per-pod metrics using the &ldquo;pods&rdquo; source.
This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.</p>
</td>
</tr>
<tr>
<td>
<code>external</code><br/>
<em>
<a href="#starrocks.com/v1.ExternalMetricSource">
ExternalMetricSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>external refers to a global metric that is not associated
with any Kubernetes object. It allows autoscaling based on information
coming from components running outside of cluster
(for example length of queue in cloud messaging service, or
QPS from loadbalancer running outside of cluster).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.MetricTarget">MetricTarget
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.ContainerResourceMetricSource">ContainerResourceMetricSource</a>, <a href="#starrocks.com/v1.ExternalMetricSource">ExternalMetricSource</a>, <a href="#starrocks.com/v1.ObjectMetricSource">ObjectMetricSource</a>, <a href="#starrocks.com/v1.PodsMetricSource">PodsMetricSource</a>, <a href="#starrocks.com/v1.ResourceMetricSource">ResourceMetricSource</a>)
</p>
<div>
<p>MetricTarget defines the target value, average value, or average utilization of a specific metric</p>
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
<a href="#starrocks.com/v1.MetricTargetType">
MetricTargetType
</a>
</em>
</td>
<td>
<p>type represents whether the metric type is Utilization, Value, or AverageValue</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<em>(Optional)</em>
<p>value is the target value of the metric (as a quantity).</p>
</td>
</tr>
<tr>
<td>
<code>averageValue</code><br/>
<em>
k8s.io/apimachinery/pkg/api/resource.Quantity
</em>
</td>
<td>
<em>(Optional)</em>
<p>averageValue is the target value of the average of the
metric across all relevant pods (as a quantity)</p>
</td>
</tr>
<tr>
<td>
<code>averageUtilization</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>averageUtilization is the target value of the average of the
resource metric across all relevant pods, represented as a percentage of
the requested value of the resource for the pods.
Currently only valid for Resource metric source type</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.MetricTargetType">MetricTargetType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricTarget">MetricTarget</a>)
</p>
<div>
<p>MetricTargetType specifies the type of metric being targeted, and should be either
&ldquo;Value&rdquo;, &ldquo;AverageValue&rdquo;, or &ldquo;Utilization&rdquo;</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;AverageValue&#34;</p></td>
<td><p>AverageValueMetricType declares a MetricTarget is an</p>
</td>
</tr><tr><td><p>&#34;Utilization&#34;</p></td>
<td><p>UtilizationMetricType declares a MetricTarget is an AverageUtilization value</p>
</td>
</tr><tr><td><p>&#34;Value&#34;</p></td>
<td><p>ValueMetricType declares a MetricTarget is a raw value</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.MountInfo">MountInfo
</h3>
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
<p>This must match the Name of a ConfigMap or Secret in the same namespace, and
the length of name must not more than 50 characters.</p>
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
<p>Path within the container at which the volume should be mounted.  Must
not contain &lsquo;:&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>subPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>SubPath within the volume from which the container&rsquo;s volume should be mounted.
Defaults to &ldquo;&rdquo; (volume&rsquo;s root).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ObjectMetricSource">ObjectMetricSource
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>ObjectMetricSource indicates how to scale on a metric describing a
kubernetes object (for example, hits-per-second on an Ingress object).</p>
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
<code>describedObject</code><br/>
<em>
<a href="#starrocks.com/v1.CrossVersionObjectReference">
CrossVersionObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="#starrocks.com/v1.MetricTarget">
MetricTarget
</a>
</em>
</td>
<td>
<p>target specifies the target value for the given metric</p>
</td>
</tr>
<tr>
<td>
<code>metric</code><br/>
<em>
<a href="#starrocks.com/v1.MetricIdentifier">
MetricIdentifier
</a>
</em>
</td>
<td>
<p>metric identifies the target metric by name and selector</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.Phase">Phase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>, <a href="#starrocks.com/v1.StarRocksWarehouseStatus">StarRocksWarehouseStatus</a>)
</p>
<div>
<p>Phase represent the cluster phase. the possible value for cluster phase are: running, failed, pending, deleting.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;deleting&#34;</p></td>
<td><p>ClusterDeleting waiting all resource deleted</p>
</td>
</tr><tr><td><p>&#34;failed&#34;</p></td>
<td><p>ClusterFailed represents starrocks cluster failed.</p>
</td>
</tr><tr><td><p>&#34;pending&#34;</p></td>
<td><p>ClusterPending represents the starrocks cluster is creating</p>
</td>
</tr><tr><td><p>&#34;running&#34;</p></td>
<td><p>ClusterRunning represents starrocks cluster is running.</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.PodsMetricSource">PodsMetricSource
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>PodsMetricSource indicates how to scale on a metric describing each pod in
the current scale target (for example, transactions-processed-per-second).
The values will be averaged together before being compared to the target
value.</p>
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
<code>metric</code><br/>
<em>
<a href="#starrocks.com/v1.MetricIdentifier">
MetricIdentifier
</a>
</em>
</td>
<td>
<p>metric identifies the target metric by name and selector</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="#starrocks.com/v1.MetricTarget">
MetricTarget
</a>
</em>
</td>
<td>
<p>target specifies the target value for the given metric</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ResourceMetricSource">ResourceMetricSource
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.MetricSpec">MetricSpec</a>)
</p>
<div>
<p>ResourceMetricSource indicates how to scale on a resource metric known to
Kubernetes, as specified in requests and limits, describing each pod in the
current scale target (e.g. CPU or memory).  The values will be averaged
together before being compared to the target.  Such metrics are built in to
Kubernetes, and have special scaling options on top of those available to
normal per-pod metrics using the &ldquo;pods&rdquo; source.  Only one &ldquo;target&rdquo; type
should be set.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcename-v1-core">
Kubernetes core/v1.ResourceName
</a>
</em>
</td>
<td>
<p>name is the name of the resource in question.</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="#starrocks.com/v1.MetricTarget">
MetricTarget
</a>
</em>
</td>
<td>
<p>target specifies the target value for the given metric</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.ScalingPolicySelect">ScalingPolicySelect
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.HPAScalingRules">HPAScalingRules</a>)
</p>
<div>
<p>ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Disabled&#34;</p></td>
<td><p>DisabledPolicySelect disables the scaling in this direction.</p>
</td>
</tr><tr><td><p>&#34;Max&#34;</p></td>
<td><p>MaxPolicySelect selects the policy with the highest possible change.</p>
</td>
</tr><tr><td><p>&#34;Min&#34;</p></td>
<td><p>MinPolicySelect selects the policy with the lowest possible change.</p>
</td>
</tr></tbody>
</table>
<h3 id="starrocks.com/v1.SecretReference">SecretReference
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksComponentSpec">StarRocksComponentSpec</a>)
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
<p>This must match the Name of a ConfigMap or Secret in the same namespace, and
the length of name must not more than 50 characters.</p>
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
<p>Path within the container at which the volume should be mounted.  Must
not contain &lsquo;:&rsquo;.</p>
</td>
</tr>
<tr>
<td>
<code>subPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>SubPath within the volume from which the container&rsquo;s volume should be mounted.
Defaults to &ldquo;&rdquo; (volume&rsquo;s root).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.SpecInterface">SpecInterface
</h3>
<div>
<p>SpecInterface is a common interface for all starrocks component spec.</p>
</div>
<h3 id="starrocks.com/v1.StarRocksBeSpec">StarRocksBeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
<tr>
<td>
<code>runAsNonRoot</code><br/>
<em>
bool
</em>
</td>
<td>
<p>RunAsNonRoot is used to determine whether to run starrocks as a normal user.
If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
default: nil</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapReference">
[]ConfigMapReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which allow users to mount any files to container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="#starrocks.com/v1.SecretReference">
[]SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for secrets.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
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
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksBeStatus">StarRocksBeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
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
<p>FailedInstances failed pod names.</p>
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
<a href="#starrocks.com/v1.ComponentPhase">
ComponentPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of component status. If component have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.</p>
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
<h3 id="starrocks.com/v1.StarRocksCluster">StarRocksCluster
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
<a href="#starrocks.com/v1.StarRocksClusterSpec">
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
<a href="#starrocks.com/v1.StarRocksFeSpec">
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
<a href="#starrocks.com/v1.StarRocksBeSpec">
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
<a href="#starrocks.com/v1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn configuration for start cn service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeProxySpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksFeProxySpec">
StarRocksFeProxySpec
</a>
</em>
</td>
<td>
<p>StarRocksLoadSpec define a proxy for fe.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksClusterStatus">
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
<h3 id="starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksCluster">StarRocksCluster</a>)
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
<a href="#starrocks.com/v1.StarRocksFeSpec">
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
<a href="#starrocks.com/v1.StarRocksBeSpec">
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
<a href="#starrocks.com/v1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn configuration for start cn service.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeProxySpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksFeProxySpec">
StarRocksFeProxySpec
</a>
</em>
</td>
<td>
<p>StarRocksLoadSpec define a proxy for fe.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksCluster">StarRocksCluster</a>)
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
<a href="#starrocks.com/v1.Phase">
Phase
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
<a href="#starrocks.com/v1.StarRocksFeStatus">
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
<a href="#starrocks.com/v1.StarRocksBeStatus">
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
<a href="#starrocks.com/v1.StarRocksCnStatus">
StarRocksCnStatus
</a>
</em>
</td>
<td>
<p>Represents the status of cn. the status have running, failed and creating pods.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksFeProxyStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksFeProxyStatus">
StarRocksFeProxyStatus
</a>
</em>
</td>
<td>
<p>Represents the status of fe proxy. the status have running, failed and creating pods.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksCnSpec">StarRocksCnSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec</a>, <a href="#starrocks.com/v1.StarRocksWarehouseSpec">StarRocksWarehouseSpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
<tr>
<td>
<code>runAsNonRoot</code><br/>
<em>
bool
</em>
</td>
<td>
<p>RunAsNonRoot is used to determine whether to run starrocks as a normal user.
If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
default: nil</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapReference">
[]ConfigMapReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which allow users to mount any files to container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="#starrocks.com/v1.SecretReference">
[]SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for secrets.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
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
<code>autoScalingPolicy</code><br/>
<em>
<a href="#starrocks.com/v1.AutoScalingPolicy">
AutoScalingPolicy
</a>
</em>
</td>
<td>
<p>AutoScalingPolicy auto scaling strategy</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksCnStatus">StarRocksCnStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>, <a href="#starrocks.com/v1.StarRocksWarehouseStatus">StarRocksWarehouseStatus</a>)
</p>
<div>
<p>StarRocksCnStatus represents the status of starrocks cn.</p>
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
<p>FailedInstances failed pod names.</p>
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
<a href="#starrocks.com/v1.ComponentPhase">
ComponentPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of component status. If component have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.</p>
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
<tr>
<td>
<code>hpaName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The policy name of autoScale.
Deprecated</p>
</td>
</tr>
<tr>
<td>
<code>horizontalScaler</code><br/>
<em>
<a href="#starrocks.com/v1.HorizontalScaler">
HorizontalScaler
</a>
</em>
</td>
<td>
<p>HorizontalAutoscaler have the autoscaler information.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksComponentSpec">StarRocksComponentSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1.StarRocksFeSpec">StarRocksFeSpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
<tr>
<td>
<code>runAsNonRoot</code><br/>
<em>
bool
</em>
</td>
<td>
<p>RunAsNonRoot is used to determine whether to run starrocks as a normal user.
If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
default: nil</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapReference">
[]ConfigMapReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which allow users to mount any files to container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="#starrocks.com/v1.SecretReference">
[]SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for secrets.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksComponentStatus">StarRocksComponentStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksBeStatus">StarRocksBeStatus</a>, <a href="#starrocks.com/v1.StarRocksCnStatus">StarRocksCnStatus</a>, <a href="#starrocks.com/v1.StarRocksFeProxyStatus">StarRocksFeProxyStatus</a>, <a href="#starrocks.com/v1.StarRocksFeStatus">StarRocksFeStatus</a>)
</p>
<div>
<p>StarRocksComponentStatus represents the status of a starrocks component.</p>
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
<p>FailedInstances failed pod names.</p>
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
<a href="#starrocks.com/v1.ComponentPhase">
ComponentPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of component status. If component have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.</p>
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
<h3 id="starrocks.com/v1.StarRocksFeProxySpec">StarRocksFeProxySpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
<tr>
<td>
<code>resolver</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksFeProxyStatus">StarRocksFeProxyStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
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
<p>FailedInstances failed pod names.</p>
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
<a href="#starrocks.com/v1.ComponentPhase">
ComponentPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of component status. If component have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.</p>
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
<h3 id="starrocks.com/v1.StarRocksFeSpec">StarRocksFeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
<tr>
<td>
<code>runAsNonRoot</code><br/>
<em>
bool
</em>
</td>
<td>
<p>RunAsNonRoot is used to determine whether to run starrocks as a normal user.
If RunAsNonRoot is true, operator will set RunAsUser and RunAsGroup to 1000 in securityContext.
default: nil</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapReference">
[]ConfigMapReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which allow users to mount any files to container.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="#starrocks.com/v1.SecretReference">
[]SecretReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for secrets.</p>
</td>
</tr>
<tr>
<td>
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
</td>
</tr>
<tr>
<td>
<code>feEnvVars</code><br/>
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
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksFeStatus">StarRocksFeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
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
<p>FailedInstances failed pod names.</p>
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
<a href="#starrocks.com/v1.ComponentPhase">
ComponentPhase
</a>
</em>
</td>
<td>
<p>Phase the value from all pods of component status. If component have one failed pod phase=failed,
also if fe have one creating pod phase=creating, also if component all running phase=running, others unknown.</p>
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
<h3 id="starrocks.com/v1.StarRocksLoadSpec">StarRocksLoadSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksComponentSpec">StarRocksComponentSpec</a>, <a href="#starrocks.com/v1.StarRocksFeProxySpec">StarRocksFeProxySpec</a>)
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>annotation for pods. user can config monitor annotation for collect to monitor system.</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>the pod labels for user select or classify pods.</p>
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
<p>Replicas is the number of desired Pod.
When HPA policy is enabled with a fixed replica count: every time the starrockscluster CR is
applied, the replica count of the StatefulSet object in K8S will be reset to the value
specified by the &lsquo;Replicas&rsquo; field, erasing the value previously set by HPA.
So operator will set it to nil when HPA policy is enabled.</p>
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
<em>(Optional)</em>
<p>Image for a starrocks deployment.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SchedulerName is the name of the kubernetes scheduler that will be used to schedule the pods.</p>
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
<tr>
<td>
<code>probe</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksProbe">
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
<code>service</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksService">
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
<code>storageVolumes</code><br/>
<em>
<a href="#starrocks.com/v1.StorageVolume">
[]StorageVolume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StorageVolumes defines the additional storage for meta storage.</p>
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
<p>serviceAccount for access cloud service.</p>
</td>
</tr>
<tr>
<td>
<code>configMapInfo</code><br/>
<em>
<a href="#starrocks.com/v1.ConfigMapInfo">
ConfigMapInfo
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>the reference for configMap which store the config info to start starrocks. e.g. be.conf, fe.conf, cn.conf.</p>
</td>
</tr>
<tr>
<td>
<code>startupProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbeFailureSeconds defines the total failure seconds of startupProbe.
By default, the startupProbe use HTTP /api/health Probe, the failureThreshold is 60, the periodSeconds is 5.
If FE needs to more time to start, you can change the default value by setting the StartupProbeFailureSeconds field.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksProbe">StarRocksProbe
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksLoadSpec">StarRocksLoadSpec</a>)
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
<h3 id="starrocks.com/v1.StarRocksService">StarRocksService
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksLoadSpec">StarRocksLoadSpec</a>)
</p>
<div>
<p>StarRocksService defines external service for starrocks component.</p>
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
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations store Kubernetes Service annotations.</p>
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
<code>loadBalancerIP</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Only applies to Service Type: LoadBalancer.
This feature depends on whether the underlying cloud-provider supports specifying
the loadBalancerIP when a load balancer is created.
This field will be ignored if the cloud-provider does not support the feature.
This field was under-specified and its meaning varies across implementations,
and it cannot support dual-stack.
As of Kubernetes v1.24, users are encouraged to use implementation-specific annotations when available.
This field may be removed in a future API version.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksServicePort">
[]StarRocksServicePort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports are the ports that are exposed by this service.
You can override the default port information by specifying the same StarRocksServicePort.Name in the ports list.
e.g. if you want to use a dedicated node port, you can just specify the StarRocksServicePort.Name and
StarRocksServicePort.NodePort field.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksServicePort">StarRocksServicePort
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksService">StarRocksService</a>)
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
<em>(Optional)</em>
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
<em>(Optional)</em>
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
<em>(Optional)</em>
<p>The easiest way to expose fe, cn or be is to use a Service of type <code>NodePort</code>.
The range of valid ports is 30000-32767</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksWarehouse">StarRocksWarehouse
</h3>
<div>
<p>StarRocksWarehouse defines a starrocks warehouse.</p>
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
<a href="#starrocks.com/v1.StarRocksWarehouseSpec">
StarRocksWarehouseSpec
</a>
</em>
</td>
<td>
<p>Spec represents the specification of desired state of a starrocks warehouse.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>starRocksCluster</code><br/>
<em>
string
</em>
</td>
<td>
<p>StarRocksCluster is the name of a StarRocksCluster which the warehouse belongs to.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn component configuration for start cn service.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksWarehouseStatus">
StarRocksWarehouseStatus
</a>
</em>
</td>
<td>
<p>Status represents the recent observed status of the starrocks warehouse.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksWarehouseSpec">StarRocksWarehouseSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksWarehouse">StarRocksWarehouse</a>)
</p>
<div>
<p>StarRocksWarehouseSpec defines the desired state of StarRocksWarehouse</p>
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
<code>starRocksCluster</code><br/>
<em>
string
</em>
</td>
<td>
<p>StarRocksCluster is the name of a StarRocksCluster which the warehouse belongs to.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksCnSpec">
StarRocksCnSpec
</a>
</em>
</td>
<td>
<p>StarRocksCnSpec define cn component configuration for start cn service.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksWarehouseStatus">StarRocksWarehouseStatus
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksWarehouse">StarRocksWarehouse</a>)
</p>
<div>
<p>StarRocksWarehouseStatus defines the observed state of StarRocksWarehouse.</p>
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
<a href="#starrocks.com/v1.Phase">
Phase
</a>
</em>
</td>
<td>
<p>Phase represents the state of a warehouse. The possible value are: running, failed, pending and deleting.</p>
</td>
</tr>
<tr>
<td>
<code>starRocksCnStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksCnStatus">
StarRocksCnStatus
</a>
</em>
</td>
<td>
<p>StarRocksCnStatus represents the status of cn service. The status has reconciling, failed and running.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StorageVolume">StorageVolume
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksLoadSpec">StarRocksLoadSpec</a>)
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
<tr>
<td>
<code>subPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>SubPath within the volume from which the container&rsquo;s volume should be mounted.
Defaults to &ldquo;&rdquo; (volume&rsquo;s root).</p>
</td>
</tr>
</tbody>
</table>
<hr/>
