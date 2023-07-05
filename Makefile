
# Image URL to use all building/pushing image targets
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
ifneq (,$(shell go env GOPATH))
GOBIN=$(shell go env GOPATH)/bin
endif
else
GOBIN=$(shell go env GOBIN)
endif

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
ifeq (,$(GOBIN))
GOBIN=$PROJECT_DIR/bin
endif

$(shell mkdir -p $(GOBIN))

LATEST_TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)
LATEST_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null)

BUILD_DATE=$(shell date +"%Y%m%d-%T")
ifeq (,$(LATEST_TAG))
VERSION="UNKNOWN"
else
VERSION=$(shell echo $(LATEST_TAG) | tr -d "v")
endif

GOVERSION=$(shell go version | awk -F ' ' '{print $$3}')

LDFLAGS="-s -X \"main.VERSION=$(VERSION)\" -X \"main.GOVERSION=$(GOVERSION)\" -X \"main.COMMIT=$(LATEST_COMMIT)\" -X \"main.BUILD_DATE=$(BUILD_DATE)\""

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects in config/crd/bases and deploy.
	$(CONTROLLER_GEN) rbac:roleName=starrocks-manager crd:maxDescLen=0 webhook paths="./pkg/apis/..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=starrocks-manager crd:maxDescLen=0 webhook paths="./pkg/apis/..." output:crd:artifacts:config=deploy/ output:rbac:artifacts:config=deploy/


.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile=".header" paths="./pkg/apis/..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

##@ Build
.PHONY: tidy
tidy: ## Run go vet against code.
	go mod tidy

.PHONY: build
build: tidy generate fmt vet crd-all ## Build operator binary,name=manager, path=bin/ .
	GOOS=linux go build -ldflags=$(LDFLAGS) -o bin/sroperator cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run -ldflags=$(LDFLAGS) cmd/main.go

.PHONY: docker
docker: ## use docker build
	docker build --rm --no-cache --build-arg LDFLAGS=$(LDFLAGS) -f Dockerfile -t "$(IMG)"  .

.PHONY: push
push: docker
	docker push "$(IMG)"

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	# if kubectl create command failed, because crd already exists, use kubectl replace instead.
	$(KUSTOMIZE) build config/crd | kubectl create -f - || $(KUSTOMIZE) build config/crd | kubectl replace -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

#CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
CONTROLLER_GEN = $(GOBIN)/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

GEN_DOCS = $(GOBIN)/gen-crd-api-reference-docs
.PHONY: gen-tool
gen-tool: ## Download gen-crd-api-reference-docs locally 0f necessary.
	$(call go-get-tool,$(GEN_DOCS),github.com/ahmetb/gen-crd-api-reference-docs@v0.3.0)

KUSTOMIZE = $(GOBIN)/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

ENVTEST = $(GOBIN)/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: gen-api
gen-api: gen-tool
	$(GEN_DOCS)  -api-dir "./pkg/apis/starrocks/v1" -config "./scripts/docs/config.json" -template-dir "./scripts/docs/template" -out-file "doc/api.html"
	$(GEN_DOCS)  -api-dir "./pkg/apis/starrocks/v1alpha1" -config "./scripts/docs/config.json" -template-dir "./scripts/docs/template" -out-file "doc/v1alpha1_api.html"


.PHONY: crd-all
crd-all: generate manifests gen-api

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
echo $TMP_DIR; \
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(GOBIN) go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

