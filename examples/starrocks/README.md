This document explains how to deploy a StarRocks cluster using the Operator. Users can take these examples as
references, and tailor them as necessary to fit the requirement. It primarily covers:

1. [Deploying a very simple StarRocks cluster](./deploy_a_starrocks_cluster_with_no_ha.yaml)
2. [Deploying a HA StarRocks cluster](./deploy_a_ha_starrocks_cluster.yaml)
3. [Deploying a StarRocks cluster with the CN component](./deploy_a_starrocks_cluster_with_cn.yaml)
4. [Deploying a StarRocks cluster with custom configurations](./deploy_a_starrocks_cluster_with_custom_configurations.yaml)
5. [Deploying a StarRocks cluster with persistent storage](./deploy_a_starrocks_cluster_with_persistent_storage.yaml)
6. [Deploying a StarRocks cluster running in shared_data mode](./deploy_a_starrocks_cluster_running_in_shared_data_mode.yaml)
7. [Deploying a StarRocks cluster with the FE Proxy component](./deploy_a_starrocks_cluster_with_fe_proxy.yaml)
8. [Deploying a StarRocks cluster with all the above features](./deploy_a_starrocks_cluster_with_all_features.yaml)

> Note:
>
> Some of the example YAML files need to be edited before using them. For example, the `shared_data mode` example needs editing to specify the shared data (MinIO, AWS, OSS, etc.) location and credentials. When editing these examples you will generally be editing ConfigMaps in the example file.
