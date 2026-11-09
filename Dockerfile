# Docker file for building and packaing the operator.
#
# Run the following command from the root dir of the git repo:
#   > DOCKER_BUILDKIT=1 docker build -t starrocks/operator:tag .

FROM golang:1.25.13 AS build
ARG LDFLAGS
WORKDIR /go/src/app

# Download modules first so this layer is cached across source-only changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${LDFLAGS:-}" -o /app/sroperator cmd/main.go

FROM starrocks/static-debian11:nonroot
COPY --from=build /app/sroperator /sroperator
CMD ["/sroperator"]
