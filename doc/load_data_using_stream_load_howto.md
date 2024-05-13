# Load Data Using Stream Load

The issue is that when the load client residents in a different network other than FE/BE's private network. FE's HTTP
307 brings BE's private network address to the client who does not recognize and can't process the redirection.

FE Proxy is a reverse proxy that can be used to solve this problem. It is a nginx server that listens on port 8080 and
proxies the HTTP request to FE and BE, including the HTTP 307 redirection.
> You need to switch the traffic, that is, the request sent to FE HTTP Port (8030) is sent to FE Proxy HTTP Port (8080).

**Note: FE proxy solves the data transfer link through HTTP protocol, non-HTTP traffic can't be proxied by the FE proxy
such as the spark connector reading data directly from BE nodes through thrift protocol.**

The following solutions for other read and write data scenarios are listed (will continue to be supplemented):

1. If you are unloading (reading) the data with the spark connector outside of k8s, a workaround
   is [INSERT INTO FILES](https://docs.starrocks.io/docs/unloading/unload_using_insert_into_files/), and then use spark
   to load data from the exported files.

## Deploy Fe Proxy Using Helm

If you install StarRocks with Helm, you can add the following configuration to the `values.yaml` file:

For `kube-starrocks` Helm chart:

```yaml
starrocks:
  starRocksFeProxySpec:
    enabled: true
    replicas: 1
    # set the resolver for nginx server, default kube-dns.kube-system.svc.cluster.local
    resolver: ""
    limits:
      cpu: 1
      memory: 2Gi
    requests:
      cpu: 1
      memory: 2Gi
    service:
      type: NodePort
      ports:
        - name: http-port   # fill the name from the fe proxy service ports
          containerPort: 8080
          nodePort: 30180
          port: 8080
```

For `starrocks` Helm chart:

```yaml
starRocksFeProxySpec:
  enabled: true
  replicas: 1
  # set the resolver for nginx server, default kube-dns.kube-system.svc.cluster.local
  resolver: ""
  limits:
    cpu: 1
    memory: 2Gi
  requests:
    cpu: 1
    memory: 2Gi
  service:
    type: NodePort
    ports:
      - name: http-port   # fill the name from the fe proxy service ports
        containerPort: 8080
        nodePort: 30180
        port: 8080
```

Please
see https://github.com/StarRocks/starrocks-kubernetes-operator/blob/main/helm-charts/charts/kube-starrocks/values.yaml
for more details about how to configure `starRocksFeProxySpec`.

## Deploy Fe Proxy Using StarRocksCluster CR

If you install StarRocks with StarRocksCluster CR yaml, please see [deploy_a_starrocks_cluster_with_fe_proxy.md](
../examples/starrocks/deploy_a_starrocks_cluster_with_fe_proxy.yaml)
