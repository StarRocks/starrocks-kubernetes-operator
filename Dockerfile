# Docker file for building and packaing the operator.
#
# Run the following command from the root dir of the git repo:
#   > DOCKER_BUILDKIT=1 docker build -t starrocks/operator:tag .

FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

# Get dependancies
RUN go mod download

# Run all the test cases before build.
# TODO: uncomment this when this issue is resolved: https://github.com/StarRocks/starrocks-kubernetes-operator/issues/37
# RUN make test

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/sroperator cmd/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=build /app/sroperator /app/sroperator
CMD ["/app/sroperator"]
