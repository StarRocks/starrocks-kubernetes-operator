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
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1.WarehouseComponentSpec">WarehouseComponentSpec</a>)
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
<p>ComponentPhase represent the component phase. e.g.
1. StarRocksCluster contains three components: FE, CN, BE.
2. StarRocksWarehouse reuse the CN component.
The possible value for component phase are: reconciling, failed, running.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#metricspec-v2beta2-autoscaling">
[]Kubernetes autoscaling/v2beta2.MetricSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#horizontalpodautoscalerbehavior-v2beta2-autoscaling">
Kubernetes autoscaling/v2beta2.HorizontalPodAutoscalerBehavior
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
<h3 id="starrocks.com/v1.Phase">Phase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterStatus">StarRocksClusterStatus</a>)
</p>
<div>
<p>Phase is defined under status, e.g.
1. StarRocksClusterStatus.Phase represents the phase of starrocks cluster.
2. StarRocksWarehouseStatus.Phase represents the phase of starrocks warehouse.
The possible value for cluster phase are: running, failed, pending, deleting.</p>
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
<code>StarRocksComponentSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentSpec">
StarRocksComponentSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentSpec</code> are embedded into this type.)
</p>
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
<code>StarRocksComponentStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentStatus">
StarRocksComponentStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentStatus</code> are embedded into this type.)
</p>
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
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksClusterSpec">StarRocksClusterSpec</a>)
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
<code>StarRocksComponentSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentSpec">
StarRocksComponentSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentSpec</code> are embedded into this type.)
</p>
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
<p>WarehouseComponentStatus represents the status of component.</p>
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
<code>StarRocksComponentStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentStatus">
StarRocksComponentStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentStatus</code> are embedded into this type.)
</p>
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
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksBeSpec">StarRocksBeSpec</a>, <a href="#starrocks.com/v1.StarRocksCnSpec">StarRocksCnSpec</a>, <a href="#starrocks.com/v1.StarRocksFeSpec">StarRocksFeSpec</a>, <a href="#starrocks.com/v1.WarehouseComponentSpec">WarehouseComponentSpec</a>)
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
<code>StarRocksLoadSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksLoadSpec">
StarRocksLoadSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksLoadSpec</code> are embedded into this type.)
</p>
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
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>TerminationGracePeriodSeconds defines duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down).
If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in the pod are sent
a termination signal and the time when the processes are forcibly halted with a kill signal.
Set this value longer than the expected cleanup time for your process.
Defaults to 120 seconds.</p>
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
<p>ResourceNames the statefulset names of fe.</p>
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
<code>StarRocksLoadSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksLoadSpec">
StarRocksLoadSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksLoadSpec</code> are embedded into this type.)
</p>
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
<code>StarRocksComponentStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentStatus">
StarRocksComponentStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentStatus</code> are embedded into this type.)
</p>
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
<code>StarRocksComponentSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentSpec">
StarRocksComponentSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentSpec</code> are embedded into this type.)
</p>
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
<code>StarRocksComponentStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentStatus">
StarRocksComponentStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentStatus</code> are embedded into this type.)
</p>
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
<code>ResourceRequirements</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>
(Members of <code>ResourceRequirements</code> are embedded into this type.)
</p>
<em>(Optional)</em>
<p>defines the specification of resource cpu and mem.</p>
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
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the
images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for
them to use.
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
<p>StartupProbeFailureSeconds defines the total failure seconds of startup Probe.
Default failureThreshold is 60 and the periodSeconds is 5, this means the startup
will fail if the pod can&rsquo;t start in 300 seconds. Your StartupProbeFailureSeconds is
the total time of seconds before startupProbe give up and fail the container start.
If startupProbeFailureSeconds can&rsquo;t be divided by defaultPeriodSeconds, the failureThreshold
will be rounded up.</p>
</td>
</tr>
<tr>
<td>
<code>livenessProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>LivenessProbeFailureSeconds defines the total failure seconds of liveness Probe.
Default failureThreshold is 3 and the periodSeconds is 5, this means the liveness
will fail if the pod can&rsquo;t respond in 15 seconds. Your LivenessProbeFailureSeconds is
the total time of seconds before the container restart. If LivenessProbeFailureSeconds
can&rsquo;t be divided by defaultPeriodSeconds, the failureThreshold will be rounded up.</p>
</td>
</tr>
<tr>
<td>
<code>readinessProbeFailureSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>ReadinessProbeFailureSeconds defines the total failure seconds of readiness Probe.
Default failureThreshold is 3 and the periodSeconds is 5, this means the readiness
will fail if the pod can&rsquo;t respond in 15 seconds. Your ReadinessProbeFailureSeconds is
the total time of seconds before pods becomes not ready. If ReadinessProbeFailureSeconds
can&rsquo;t be divided by defaultPeriodSeconds, the failureThreshold will be rounded up.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="starrocks.com/v1.StarRocksProbe">StarRocksProbe
</h3>
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
<code>template</code><br/>
<em>
<a href="#starrocks.com/v1.WarehouseComponentSpec">
WarehouseComponentSpec
</a>
</em>
</td>
<td>
<p>Template define component configuration.</p>
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
<code>template</code><br/>
<em>
<a href="#starrocks.com/v1.WarehouseComponentSpec">
WarehouseComponentSpec
</a>
</em>
</td>
<td>
<p>Template define component configuration.</p>
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
<code>WarehouseComponentStatus</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksCnStatus">
StarRocksCnStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>WarehouseComponentStatus</code> are embedded into this type.)
</p>
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
<h3 id="starrocks.com/v1.WarehouseComponentSpec">WarehouseComponentSpec
</h3>
<p>
(<em>Appears on:</em><a href="#starrocks.com/v1.StarRocksWarehouseSpec">StarRocksWarehouseSpec</a>)
</p>
<div>
<p>WarehouseComponentSpec defines the desired state of component.</p>
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
<code>StarRocksComponentSpec</code><br/>
<em>
<a href="#starrocks.com/v1.StarRocksComponentSpec">
StarRocksComponentSpec
</a>
</em>
</td>
<td>
<p>
(Members of <code>StarRocksComponentSpec</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>envVars</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>envVars is a slice of environment variables that are added to the pods, the default is empty.</p>
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
<p>AutoScalingPolicy defines auto scaling policy</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>c86f10d</code>.
</em></p>
