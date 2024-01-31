You can install the StarRocks operator and StarRocks cluster by kubectl or helm. No matter which way you choose, you
may need the following permissions:

> Note: Operator will use its own service account, cluster role and cluster role binding to create and manage StarRocks

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: install-starrocks-rb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: install-starrocks-role
subjects:
  - kind: ServiceAccount
    name: your-sa-name
    namespace: your-namespace

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: install-starrocks-role
rules:
  - apiGroups:
      - ""
    resources:
      - secrets
      - serviceaccounts
      - configmaps
    verbs:
      - '*'
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterrolebindings
      - rolebindings
      - clusterroles
      - roles
    verbs:
      - '*'
  - apiGroups:
      - apps
    resources:
      - deployments
    verbs:
      - '*'
  - apiGroups:
      - monitoring.coreos.com
    resources:
      - servicemonitors
    verbs:
      - '*'
  - apiGroups:
      - starrocks.com
    resources:
      - starrocksclusters
      - starrockswarehouses
    verbs:
      - '*'
  - apiGroups:
      - apiextensions.k8s.io
    resources:
      - customresourcedefinitions
    verbs:
      - '*'
  - apiGroups:
      - batch
    resources:
      - jobs
    verbs:
      - '*'
```
