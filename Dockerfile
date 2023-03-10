# Docker file for building and packaing the operator.
#
# Run the following command from the root dir of the git repo:
#   > DOCKER_BUILDKIT=1 docker build -t starrocks/operator:tag .

FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

ENV http_proxy=http://172.26.92.139:28888
ENV https_proxy=http://172.26.92.139:28888

# Run all the test cases before build.
RUN make test

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/sroperator cmd/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=build /app/sroperator /app/sroperator
CMD ["/app/sroperator"]
