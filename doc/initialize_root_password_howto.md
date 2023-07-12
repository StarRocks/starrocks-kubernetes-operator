# Initialize root user password HOWTO

The password is empty for the `root` user when deploying a StarRocks cluster from fresh installation. This can be a security concern. This document describes steps to initialize root password from fresh installation.

In the following examples, `mysql_password` is taken as the password for the root user, it can be replaced with any password chosen for the root.

## Prerequisites

Deploy StarRocks cluster using Helm and custom configuration.

## Steps to Change the Password For root User

1. Add the Helm Chart Repo for StarRocks. The Helm Chart contains the definitions of the StarRocks Operator and the custom resource StarRocksCluster.Refer to: https://docs.mirrorship.cn/en-us/latest/deployment/helm#procedure
    
2. Create a YAML file, for example, my-values.yaml, add the following snippets to my-values.yaml to initialize the password.

```yaml
# To initialize root user password
initPassword: 
    enabled: true
    password: "mysql_password"
```

Complete other custom configurations for the StarRocks Operator and StarRocks cluster in the YAML file.

3. Run the following command to deploy the StarRocks Operator and StarRocks cluster with the custom configurations in my-values.yaml.

```shell
helm install -f my-values.yaml starrocks starrocks-community/kube-starrocks
```

4. After completing all deployment steps, use `mysql_password` to access the cluster.