# Build the operator images by yourself

Get the official operator image from [here](https://hub.docker.com/r/starrocks/operator/tags).

### Build starrocks operator docker image

Follow below instructions if you want to build your own image.

```console
DOCKER_BUILDKIT=1 docker build -t starrocks/operator:<tag> .
```

E.g.

```console
DOCKER_BUILDKIT=1 docker build -t starrocks/operator:latest .
```

### Publish starrocks operator docker image

```console
docker push starrocksr/operator:latest
```

E.g. Publish image to dockerhub

```console
docker push starrocks/operator:latest
```
