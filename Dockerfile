# Docker file for building and packaing the operator.
#
# Run the following command from the root dir of the git repo:
#   > DOCKER_BUILDKIT=1 docker build -t starrocks/operator:tag .

FROM golang:1.19 as build
ARG LDFLAGS
WORKDIR /go/src/app
COPY . .

# Build the binary
# if vendor directory exists, add -mod=vendor flag
RUN if [ -d vendor ]; then \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="${LDFLAGS:-}" -o /app/sroperator cmd/main.go; \
    else \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="${LDFLAGS:-}" -o /app/sroperator cmd/main.go; \
    fi

FROM ubuntu:22.04

COPY --from=build /app/sroperator /sroperator

ARG USER=starrocks
ARG GROUP=starrocks

RUN groupadd --gid 1000 $GROUP && useradd --home-dir /nonexistent --uid 1000 --gid 1000 \
             --shell /usr/sbin/nologin $USER  \
        && chown $USER:$GROUP /sroperator


USER $USER
ENV USER $USER

CMD ["/sroperator"]