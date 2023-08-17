# Build the operator images by yourself

Get the official operator image from [here](https://hub.docker.com/r/starrocks/operator/tags).

### Build starrocks operator docker image

Follow below instructions if you want to build your own image.

```console
DOCKER_BUILDKIT=1 docker build -t starrocks-kubernetes-operator/operator:<tag> .
```

E.g.

```console
DOCKER_BUILDKIT=1 docker build -t starrocks-kubernetes-operator/operator:latest .
```

### Publish starrocks operator docker image

```console
docker push ghcr.io/OWNER/starrocks-kubernetes-operator/operator:latest
```

E.g.
Publish image to ghcr

```console
docker push ghcr.io/dengliu/starrocks-kubernetes-operator/operator:latest
```
