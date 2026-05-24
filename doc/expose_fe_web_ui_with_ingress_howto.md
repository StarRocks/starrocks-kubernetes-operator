# Expose the FE Web UI With an Ingress

The StarRocks Kubernetes Operator can create an `Ingress` that routes external HTTP traffic
to the FE **web UI** (the `http` port, default `8030`). This is useful when you want to reach
the FE web UI from outside the cluster without exposing a `NodePort` or `LoadBalancer`.

The Ingress backend is pinned to the FE external service's `http` port by name, so it can
never route to the FE MySQL/query port (`9030`). `9030` is an L4 (TCP) protocol that a
standard Ingress cannot handle; for SQL access use `service.type` (`NodePort` or
`LoadBalancer`) instead.

> **Security**
> The FE web UI exposes an administrative HTTP API. An Ingress makes it reachable from
> outside the cluster. Protect it with authentication (for example an ingress-controller
> basic-auth annotation), source-IP allow lists, or a private `IngressClass` before using
> this on an untrusted network.

You need an Ingress controller installed in the cluster (for example
[ingress-nginx](https://kubernetes.github.io/ingress-nginx/), or a cloud controller such as
AWS ALB or GKE).

## 1. Enable the Ingress by StarRocks CRD YAML File

Set `starRocksFeSpec.ingress` on the `StarRocksCluster`. Leaving it unset creates no Ingress.

```yaml
apiVersion: starrocks.com/v1
kind: StarRocksCluster
metadata:
  name: kube-starrocks
  namespace: starrocks
spec:
  starRocksFeSpec:
    image: "starrocks/fe-ubuntu:3.5-latest"
    replicas: 1
    ingress:
      # host is required.
      host: starrocks.example.com
      # ingressClassName is optional; if unset the cluster default IngressClass is used.
      ingressClassName: nginx
      # annotations are optional, e.g. ingress-controller config or cert-manager hints.
      annotations:
        nginx.ingress.kubernetes.io/auth-type: basic
        nginx.ingress.kubernetes.io/auth-secret: fe-basic-auth
```

After applying it, the operator creates an Ingress named `<cluster-name>-fe-ingress`:

```bash
kubectl -n starrocks get ingress kube-starrocks-fe-ingress
```

Clearing the `ingress` field removes the previously created Ingress. The Ingress carries a
controller owner reference, so it is also garbage-collected when the cluster is deleted.

## 2. Enable the Ingress by Helm Chart

Set `starrocksFESpec.ingress` in your values file. It maps one-to-one to the CRD field.

```yaml
starrocksFESpec:
  ingress:
    host: starrocks.example.com
    ingressClassName: nginx
    annotations:
      nginx.ingress.kubernetes.io/auth-type: basic
```

## 3. TLS (HTTPS)

Two patterns are supported, depending on your ingress controller:

- **Controllers that read `spec.tls`** (ingress-nginx with cert-manager, or a pre-provisioned
  TLS secret): set the `tls` field.

  ```yaml
  ingress:
    host: starrocks.example.com
    ingressClassName: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls:
      - hosts:
          - starrocks.example.com
        secretName: starrocks-tls
  ```

- **Controllers that configure TLS through annotations** (AWS ALB `certificate-arn`, GKE
  managed certificates): leave `tls` unset and use the `annotations` field instead.
