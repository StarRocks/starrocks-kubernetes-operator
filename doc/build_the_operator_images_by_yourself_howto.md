# Build the operator images by yourself

Get the official operator image from [here](https://hub.docker.com/r/us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator/tags).

### Build celerdata operator docker image

Follow below instructions if you want to build your own image.

```console
DOCKER_BUILDKIT=1 docker build -t us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator:<tag> .
```

E.g.

```console
DOCKER_BUILDKIT=1 docker build -t us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator:latest .
```

### Publish celerdata operator docker image

```console
docker push starrocksr/operator:latest
```

E.g. Publish image to dockerhub

```console
docker push us-west1-docker.pkg.dev/phrasal-verve-350013/celerdata/operator:latest
```
