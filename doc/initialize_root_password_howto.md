# Initializing the Root User Password in StarRocks

Upon deploying a fresh StarRocks cluster, the `root` user's password remains unset, potentially posing a security risk.
This guide delineates the process to establish a root password when a new installation.

> Note that this only works for helm install, can't use it in helm upgrade

## Prerequisites

- Ensure that you have installed the Kubernetes cluster. v1.23.0+ is recommended.
- Ensure that you have installed the [Helm](https://helm.sh/) package manager. 3.0.0+ is recommended.
- Ensure the helm chart repo for StarRocks is added.
  See [Add the Helm Chart Repo for StarRocks](./add_helm_repo_howto.md).

In this guide, we will use `starrocks/kube-starrocks` chart to deploy both StarRocks operator and cluster.

## 1. Download the values.yaml file for the kube-starrocks chart

The values.yaml file contains the default configurations for the StarRocks Operator and the StarRocks cluster.

```Bash
helm show values starrocks/kube-starrocks > values.yaml
```

The following is a snippet of the values.yaml file:

```yaml
starrocks:
  # This configuration is used to modify the root password during initial deployment.
  # After deployment is completed, it won't take effect to modify the password here and to do a `helm upgrade`.
  # It also supports providing secret name that contains password, using the password in the secret instead of the plaintext in the values.yaml.
  # When both password and passwordSecret are set, only passwordSecret takes effect.
  initPassword:
    enabled: false
    password: ""
    # The secret name that contains password, the key of the secret is "password", and you should create it first.
    passwordSecret: ""
```

## 2. Initialize the root password

Configure a YAML File for Custom Settings (for instance, `my-values.yaml`). There are two ways to
initialize the root password. You can choose one of them.

### 2.1. Setting the Root User Password By plaintext

We use `mysql_password` serves as a placeholder for the root user's password. You can substitute this
with any preferred password.

To initialize the root password, embed the following snippet:

```yaml
starrocks:
  initPassword:
    enabled: true
    password: "mysql_password"
```

### 2.2. Setting the Root User Password By Secret

You can also use a secret to set the root password. The secret must be created before deploying the helm chart.

> Note the key of the secret must be `password`.

```bash
kubectl create secret generic starrocks-root-password --from-literal=password=mysql_password
```

To initialize the root password, embed the following snippet:

```yaml
starrocks:
  initPassword:
    enabled: true
    passwordSecret: starrocks-root-password
```

## 3. Deploy the StarRocks Operator and the StarRocks cluster

Execute Deployment with Custom Specifications. Run the subsequent command to deploy the StarRocks Operator and the
StarRocks cluster.

```shell
helm install -f my-values.yaml starrocks starrocks/kube-starrocks
```

## 4. Access Your Cluster

```bash
# in one terminal
kubectl port-forward service/kube-starrocks-fe-service 9030:9030

# in another terminal
mysql -h 127.0.0.1 -P 9030 -u root -p mysql_password
```
