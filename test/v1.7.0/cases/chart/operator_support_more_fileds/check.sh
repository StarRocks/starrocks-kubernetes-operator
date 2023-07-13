#! /bin/zsh

# check the rendered manifests, make sure annotations, nodeSelector, tolerations are rendered correctly.
# note: you should use zsh to run this script.

# if command fails, exit
set -e

# make manifests
manifests=$(helm template -f ../../../../../helm-charts/charts/kube-starrocks/values.yaml -f values.yaml ../../../../../helm-charts/charts/kube-starrocks/)

# there are many resources in ${manifests}, we only care about the Deployment of kube-starrocks-operator
# like this:
# ---
## Source: kube-starrocks/templates/starrocks-operator/namespace.yaml
#apiVersion: v1
#kind: Namespace
#metadata:
#  labels:
#    control-plane: cn-controller-manager
#  name: starrocks
#---
## Source: kube-starrocks/templates/starrocks-operator/service_account.yaml
#apiVersion: v1
#kind: ServiceAccount
#metadata:
#  name: starrocks
#  namespace: starrocks
#---
## Source: kube-starrocks/templates/starrocks/beconfigmap.yaml
#apiVersion: v1
#kind: ConfigMap
#metadata:
#  name: kube-starrocks-be-cm
#  namespace: starrocks
#  labels:
#    cluster: kube-starrocks
#    app: "be"
#data:
#  be.conf: |
#    be_port = 9060
#    webserver_port = 8040
#    heartbeat_service_port = 9050
#    brpc_port = 8060
#    sys_log_level = INFO
#    default_rowset_type = beta
#---
## Source: kube-starrocks/templates/starrocks/cnconfigmap.yaml
#apiVersion: v1
#kind: ConfigMap
#metadata:
#  name: kube-starrocks-cn-cm
#  namespace: starrocks
#  labels:
#    cluster: kube-starrocks
#    app: "cn"
#data:
#  cn.conf: |
#    sys_log_level = INFO
#    # ports for admin, web, heartbeat service
#    thrift_port = 9060
#    webserver_port = 8040
#    heartbeat_service_port = 9050
#    brpc_port = 8060
#---
## Source: kube-starrocks/templates/starrocks/feconfigmap.yaml
#apiVersion: v1
#kind: ConfigMap
#metadata:
#  name: kube-starrocks-fe-cm
#  namespace: starrocks
#  labels:
#    cluster: kube-starrocks
#    app: "fe"
#data:
#  fe.conf: |
#    query_port = 9030
#    edit_log_port = 9010
#---
## Source: kube-starrocks/templates/starrocks-operator/clusterrole.yaml
#apiVersion: rbac.authorization.k8s.io/v1
#kind: ClusterRole
#metadata:
#  name: kube-starrocks-operator
#  labels:
#    app: kube-starrocks-operator
#rules:
#- apiGroups:
#  - apps
#  resources:
#  - deployments
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - apps
#  resources:
#  - statefulsets
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - autoscaling
#  resources:
#  - horizontalpodautoscalers
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - batch
#  resources:
#  - cronjobs
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - configmaps
#  verbs:
#  - get
#  - list
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - endpoints
#  verbs:
#  - get
#  - list
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - pods
#  verbs:
#  - get
#  - list
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - secrets
#  verbs:
#  - get
#  - list
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - serviceaccounts
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - ""
#  resources:
#  - services
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - rbac.authorization.k8s.io
#  resources:
#  - clusterrolebindings
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - rbac.authorization.k8s.io
#  resources:
#  - rolebindings
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - starrocks.com
#  resources:
#  - computenodegroups
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - starrocks.com
#  resources:
#  - computenodegroups/finalizers
#  verbs:
#  - update
#- apiGroups:
#  - starrocks.com
#  resources:
#  - computenodegroups/status
#  verbs:
#  - get
#  - patch
#  - update
#- apiGroups:
#  - starrocks.com
#  resources:
#  - starrocksclusters
#  verbs:
#  - create
#  - delete
#  - get
#  - list
#  - patch
#  - update
#  - watch
#- apiGroups:
#  - starrocks.com
#  resources:
#  - starrocksclusters/finalizers
#  verbs:
#  - update
#- apiGroups:
#  - starrocks.com
#  resources:
#  - starrocksclusters/status
#  verbs:
#  - get
#  - patch
#  - update
#---
## Source: kube-starrocks/templates/starrocks-operator/clusterrolebinding.yaml
#apiVersion: rbac.authorization.k8s.io/v1
#kind: ClusterRoleBinding
#metadata:
#  name: kube-starrocks-operator
#roleRef:
#  apiGroup: rbac.authorization.k8s.io
#  kind: ClusterRole
#  name: kube-starrocks-operator
#subjects:
#- kind: ServiceAccount
#  name: starrocks
#  namespace: starrocks
#---
## Source: kube-starrocks/templates/starrocks-operator/leader-election-role.yaml
#apiVersion: rbac.authorization.k8s.io/v1
#kind: Role
#metadata:
#  name: cn-leader-election-role
#  namespace: starrocks
#rules:
#- apiGroups:
#  - ""
#  resources:
#  - configmaps
#  verbs:
#  - get
#  - list
#  - watch
#  - create
#  - update
#  - patch
#  - delete
#- apiGroups:
#  - coordination.k8s.io
#  resources:
#  - leases
#  verbs:
#  - get
#  - list
#  - watch
#  - create
#  - update
#  - patch
#  - delete
#- apiGroups:
#  - ""
#  resources:
#  - events
#  verbs:
#  - create
#  - patch
#---
## Source: kube-starrocks/templates/starrocks-operator/leader-election-role-binding.yaml
#apiVersion: rbac.authorization.k8s.io/v1
#kind: RoleBinding
#metadata:
#  name: cn-leader-election-rolebinding
#  namespace: starrocks
#roleRef:
#  apiGroup: rbac.authorization.k8s.io
#  kind: Role
#  name: cn-leader-election-role
#subjects:
#- kind: ServiceAccount
#  name: starrocks
#  namespace: starrocks
#---
## Source: kube-starrocks/templates/starrocks-operator/deployment.yaml
#apiVersion: apps/v1
#kind: Deployment
#metadata:
#  name: kube-starrocks-operator
#  namespace: starrocks
#  labels:
#    app: kube-starrocks-operator
#spec:
#  selector:
#    matchLabels:
#      app: kube-starrocks-operator
#      version: 1.7.0
#  replicas: 1
#  template:
#    metadata:
#      annotations:
#        kubectl.kubernetes.io/default-container: manager
#      labels:
#        app: kube-starrocks-operator
#        version: 1.7.0
#    spec:
#      securityContext:
#        runAsNonRoot: false
#      containers:
#      - command:
#        - /sroperator
#        args:
#        - --leader-elect
#        image: "starrocks/operator:latest"
#        imagePullPolicy: Always
#        name: manager
#        securityContext:
#          allowPrivilegeEscalation: false
#        livenessProbe:
#          httpGet:
#            path: /healthz
#            port: 8081
#          initialDelaySeconds: 15
#          periodSeconds: 20
#        readinessProbe:
#          httpGet:
#            path: /readyz
#            port: 8081
#          initialDelaySeconds: 5
#          periodSeconds: 10
#        # TODO(user): Configure the resources accordingly based on the project requirements.
#        # More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
#        resources:
#          limits:
#            cpu: 500m
#            memory: 200Mi
#          requests:
#            cpu: 500m
#            memory: 200Mi
#      serviceAccountName: starrocks
#      terminationGracePeriodSeconds: 10
#---
## Source: kube-starrocks/templates/starrocks/starrockscluster.yaml
#apiVersion: starrocks.com/v1
#kind: StarRocksCluster
#metadata:
#  name: kube-starrocks
#  namespace: starrocks
#  labels:
#    cluster: kube-starrocks
#    app.kubernetes.io/instance: release-name
#    app.kubernetes.io/managed-by: Helm
#spec:
#  starRocksFeSpec:
#    image: "starrocks/fe-ubuntu:3.0-latest"
#    replicas: 1
#    limits:
#      cpu: 8
#      memory: 8Gi
#    requests:
#      cpu: 4
#      memory: 4Gi
#    service:
#      type: ClusterIP
#    annotations:
#      app.starrocks.io/fe-config-hash: 55cf0059
#    fsGroup: 0
#    configMapInfo:
#      configMapName: kube-starrocks-fe-cm
#      resolveKey: fe.conf
#  starRocksBeSpec:
#    image: "starrocks/be-ubuntu:3.0-latest"
#    replicas: 1
#    limits:
#      cpu: 8
#      memory: 8Gi
#    requests:
#      cpu: 4
#      memory: 4Gi
#    service:
#      type: ClusterIP
#    annotations:
#      app.starrocks.io/be-config-hash: b1896d1b
#    fsGroup: 0
#    configMapInfo:
#      configMapName: kube-starrocks-be-cm
#      resolveKey: be.conf
#  starRocksCnSpec:
#    image: "starrocks/cn-ubuntu:3.0-latest"
#    fsGroup: 0
#    limits:
#      cpu: 8
#      memory: 8Gi
#    requests:
#      cpu: 4
#      memory: 8Gi
#    service:
#      type: ClusterIP
#    annotations:
#      app.starrocks.io/cn-config-hash: 47abe326
#    configMapInfo:
#      configMapName: kube-starrocks-cn-cm
#      resolveKey: cn.conf
#---
## Source: kube-starrocks/templates/tests/test-connection.yaml
#apiVersion: v1
#kind: Pod
#metadata:
#  name: "kube-starrocks-test-connection"
#  labels:
#    app.kubernetes.io/instance: release-name
#    app.kubernetes.io/managed-by: Helm
#  annotations:
#    "helm.sh/hook": test
#spec:
#  containers:
#    - name: wget
#      image: busybox
#      command: ['wget']
#      args: ['kube-starrocks']
#  restartPolicy: Never

# parse ${manifests} to get the Deployment manifest of kube-starrocks-operator
deployment=$(echo "${manifests}" | awk '/^---$/{i++}i==9' | sed 's/^---$//g')
echo $deployment | yq '.metadata.annotations' | grep annotationKey >/dev/null || (echo "can not find annotations!!!")
echo $deployment | yq '.spec.template.spec.nodeSelector' | grep nodeSelectorKey >/dev/null || echo "can not find nodeSelector!!!"
echo $deployment | yq '.spec.template.spec.tolerations' | grep tolerationKey >/dev/null || echo "can not find tolerations!!!"
