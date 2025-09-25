# Makefile for StarRocks Operator Upgrade Hook Feature

.PHONY: build test generate manifests

# Build the manager binary
build: generate
	go build -o bin/manager main.go

# Run tests
test: generate
	go test ./pkg/... -v

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
generate: 
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests e.g. CRD, RBAC etc.
manifests: 
	controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster
deploy: manifests
	kubectl apply -f config/default

# Generate and update API documentation
api-docs:
	@echo "Generating API documentation..."
	@echo "Please add api-docs generation tooling as needed"

# Clean build artifacts
clean:
	rm -rf bin/

# Check code style
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet lint test
