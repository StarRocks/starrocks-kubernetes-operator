# Change root user password HOWTO

The password is empty for the `root` user when deploying a StarRocks cluster from fresh installation. This can be a security concern. This document describes steps to change root password and still the operator can manage the cluster correctly.

In the following examples, `mysql_password` is taken as the password for the root user, it can be replaced with any password chosen for the root.

## Prerequisition

A StarRocks cluster is deployed and up with empty root password by the operator.

## Steps to Change the Password For root User

1. connect to starRocks FE with mysql client, and change the root user password.

```SQL
mysql -h <FE_IP/FE_SERVICE> -P 9030 -u root

# change root password to `mysql_password`
MySQL [(none)]> SET PASSWORD = PASSWORD('mysql_password');
```

2. create a secret **rootcredential** with the key **password** to store the root password

```shell
kubectl create secret generic rootcredential --from-literal=password=mysql_password
```

3. Update StarRocks crd yaml

Add the following snippets to `starRocksFeSpec/starRocksBeSpec/starRocksCnSpec` respectively if the corresponding components are deployed.
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
4. Apply the crd yaml

```shell
kubectl apply -f <crd_yaml>
```

It will trigger a rolling restart of the cluster, wait until the cluster restart completed.

5. Verify the password is all set

After the pods are restarted, run the following command to check the correctness of the password.
``` shell
kubectl exec <podName> -- sh -c 'echo $MYSQL_PWD'
```
