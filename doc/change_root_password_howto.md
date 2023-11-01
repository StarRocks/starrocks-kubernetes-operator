# Change root user password HOWTO

The password is empty for the `root` user when deploying a StarRocks cluster from fresh installation. This can be a
security concern. This document describes steps to change root password and still the operator can manage the cluster
correctly.

In the following examples, `mysql_password` is taken as the password for the root user, it can be replaced with any
password chosen for the root.

## Prerequisites

**A StarRocks cluster is deployed and up with empty root password by the operator.**

## 1. Change the password for root user

Connect to StarRocks FE with a MySQL client and change the root user password.

```SQL
mysql
-h <FE_IP/FE_SERVICE> -P 9030 -u root

# change root password to `mysql_password`
MySQL [(none)]> SET PASSWORD = PASSWORD('mysql_password');
```

## 2. Inject MYSQL_PWD environment variable to StarRocks components

There are two ways to deploy StarRocks cluster:

1. Deploy StarRocks cluster with `StarRocksCluster` CR yaml.
2. Deploy StarRocks cluster with Helm chart.

Therefore, there are two ways to inject the MYSQL_PWD environment variable into StarRocks components.

### 2.1 inject MYSQL_PWD environment variable with StarRocksCluster CRD yaml

1. Create a secret **rootcredential** with the key **password** to store the root password

   ```shell
   kubectl create secret generic rootcredential --from-literal=password=mysql_password
   ```

2. Add the following snippets to `starRocksFeSpec/starRocksBeSpec/starRocksCnSpec` respectively if the corresponding
   components are deployed.

   ```yaml
   # for starRocksFeSpec
   feEnvVars:
   - name: "MYSQL_PWD"
     valueFrom:
       secretKeyRef:
         name: rootcredential
         key: password

   # for starRocksBeSpec
   beEnvVars:
   - name: "MYSQL_PWD"
     valueFrom:
       secretKeyRef:
         name: rootcredential
         key: password

   # for starRocksCnSpec
   cnEnvVars:
   - name: "MYSQL_PWD"
     valueFrom:
       secretKeyRef:
         name: rootcredential
         key: password
   ```

3. Apply the crd yaml

   ```shell
   kubectl apply -f <crd_yaml>
   ```

It will trigger a rolling restart of the cluster, wait until the cluster restart completed.

### 2.2 Inject MYSQL_PWD environment variable with helm chart

If you are using the `kube-starrocks` Helm chart, add the following snippets to `values.yaml`.

```yaml

starrocks:
  # create secrets if necessary.
  secrets:
    - name: rootcredential
      data:
        password: mysql_password

  starrocksFESpec:
    feEnvVars:
      - name: "MYSQL_PWD"
        valueFrom:
          secretKeyRef:
            name: rootcredential
            key: password

  starrocksBeSpec:
    beEnvVars:
      - name: "MYSQL_PWD"
        valueFrom:
          secretKeyRef:
            name: rootcredential
            key: password

  starrocksCnSpec:
    cnEnvVars:
      - name: "MYSQL_PWD"
        valueFrom:
          secretKeyRef:
            name: rootcredential
            key: password
```

If you are using the `starrocks` Helm chart, add the following snippets to `values.yaml`.

```yaml
# create secrets if necessary.
secrets:
- name: rootcredential
  data:
    password: mysql_password

starrocksFESpec:
  feEnvVars:
  - name: "MYSQL_PWD"
    valueFrom:
      secretKeyRef:
        name: rootcredential
        key: password

starrocksBeSpec:
  beEnvVars:
  - name: "MYSQL_PWD"
    valueFrom:
      secretKeyRef:
        name: rootcredential
        key: password

starrocksCnSpec:
  cnEnvVars:
  - name: "MYSQL_PWD"
    valueFrom:
      secretKeyRef:
        name: rootcredential
        key: password
```

Run the following command to upgrade the cluster.

```shell
helm upgrade <release_name> <chart_path> -f values.yaml
```

It will trigger a rolling restart of the cluster, wait until the cluster restart completed.

## 3. Verify the password is all set

After the pods are restarted, run the following command to check the correctness of the password.

``` shell
kubectl exec <podName> -- sh -c 'echo $MYSQL_PWD'
```
