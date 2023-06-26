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

FROM starrocks/static-debian11

COPY --from=build /app/sroperator /sroperator
CMD ["/sroperator"]
