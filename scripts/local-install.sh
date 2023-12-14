#!/usr/bin/env bash

# This script is used to install starrocks on local k8s cluster.
# Make sure `docker` is installed，See [install docker](https://github.com/StarRocks/starrocks-kubernetes-operator/blob/main/doc/local_installation_how_to.md#11-install-docker) for more details.
#
# It will do the following things:
#   1. install kubectl, helm, kind on your machine.
#   2. create a kind cluster named `starrocks`.
#   3. install [kube-starrocks](https://github.com/StarRocks/starrocks-kubernetes-operator/tree/main/helm-charts/charts/kube-starrocks) helm chart.

# specify the k8s version installed by kind
K8S_VERSION="v1.23.4"

# helm, kind, kubectl download url
HELM_URL="https://get.helm.sh"
KIND_URL="https://kind.sigs.k8s.io/dl/v0.20.0"
KUBECTL_URL="https://dl.k8s.io/release/v1.28.3/bin"
HELM_CHART_URL="https://github.com/StarRocks/starrocks-kubernetes-operator/releases/download/v1.8.6/kube-starrocks-1.8.6.tgz"
# NOTE:
# if you can not access the following url, you can try to use the following url.
# And you can specify the url by command arguments.
# HELM_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
# KIND_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
# KUBECTL_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com"
# HELM_CHART_URL="https://ydx-starrocks-public.oss-cn-hangzhou.aliyuncs.com/kube-starrocks-1.8.6.tgz"

# checkBinary checks if the binary is installed. If not, return 1, else return 0
function checkBinary() {
  if command -v "$1" &>/dev/null; then
    echo "The binary $1 is installed"
    return 0
  else
    echo "The binary $1 is not installed"
    return 1
  fi
}

# mustInstalled checks if the binary is installed. If not, exit
function mustInstalled() {
  if ! command -v "$1" &>/dev/null; then
    echo "The binary $1 is not installed"
    exit
  else
    echo "The binary $1 is installed"
  fi
}

# installHelm installs helm
function installHelm() {
  echo "installing helm"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -LO $HELM_URL/helm-v3.12.3-linux-amd64.tar.gz &&
    tar -zxvf helm-v3.12.3-linux-amd64.tar.gz && sudo mv linux-amd64/helm /usr/local/bin/helm

  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -LO $HELM_URL/helm-v3.12.3-linux-arm64.tar.gz &&
    tar -zxvf helm-v3.12.3-linux-arm64.tar.gz && sudo mv linux-arm64/helm /usr/local/bin/helm

  # MacOS Intel
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -LO $HELM_URL/helm-v3.12.3-darwin-amd64.tar.gz &&
    tar -zxvf helm-v3.12.3-darwin-amd64.tar.gz && sudo mv darwin-amd64/helm /usr/local/bin/helm

  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -LO $HELM_URL/helm-v3.12.3-darwin-arm64.tar.gz &&
    tar -zxvf helm-v3.12.3-darwin-arm64.tar.gz && sudo mv darwin-arm64/helm /usr/local/bin/helm
}

# installKubectl installs kubectl
function installKubectl() {
  echo "Installing kubectl"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -LO "$KUBECTL_URL/linux/amd64/kubectl"

  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -LO "$KUBECTL_URL/linux/arm64/kubectl"

  # MacOS Intel Macs
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -LO "$KUBECTL_URL/darwin/amd64/kubectl"

  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -LO "$KUBECTL_URL/darwin/arm64/kubectl"

  chmod +x ./kubectl
  sudo mv ./kubectl /usr/local/bin/kubectl
}

# installKind installs kind
function installKind() {
  echo "Installing kind"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -Lo ./kind $KIND_URL/kind-linux-amd64
  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -Lo ./kind $KIND_URL/kind-linux-arm64
  # MacOS Intel
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -Lo ./kind $KIND_URL/kind-darwin-amd64
  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -Lo ./kind $KIND_URL/kind-darwin-arm64

  # install kind
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
}

# checkPrerequisites checks if the prerequisites are installed. If not, install them
function checkPrerequisites() {
  mustInstalled docker

  # If kind is not installed, install kind
  if ! checkBinary kind; then
    installKind
  fi

  # If kubectl is not installed, install kubectl
  if ! checkBinary kubectl; then
    installKubectl
  fi

  # If helm is not installed, install helm
  if ! checkBinary helm; then
    installHelm
  fi
}

# cmdHelp prints the help message
function cmdHelp() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  -h,  --help"
  echo "  -k,  --k8s-version <K8S_VERSION>, specify the version of k8s to install, default is $K8S_VERSION"
}

# kindCreateCluster creates a kind cluster
function kindCreateCluster() {
  echo "Creating kind cluster"

  CONTAINER_NAME="starrocks-control-plane"

  echo "kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30001
    hostPort: 30001
    protocol: TCP
  - containerPort: 30002
    hostPort: 30002
    protocol: TCP" >/tmp/kind.yaml

  kind create cluster --image=kindest/node:$K8S_VERSION --name=starrocks --config=/tmp/kind.yaml || exit 1
}

# install installs kube-starrocks helm chart
function install() {
  echo "Installing kube-starrocks helm chart"
  helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator
  helm repo update starrocks-community

  cat <<EOF >/tmp/local-install-values.yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.8.6

starrocks:
  starrocksFESpec:
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.1.4
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 100m
        memory: 200Mi
    service:
      type: NodePort
      ports:
      - containerPort: 9030
        name: query
        nodePort: 30002
        port: 9030
  starrocksBeSpec:
    image:
      repository: starrocks/be-ubuntu
      tag: 3.1.4
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 100m
        memory: 200Mi
  starrocksFeProxySpec:
    enabled: true
    resources:
      requests:
        cpu: 100m
        memory: 200Mi
    service:
      type: NodePort
      ports:
      - name: http-port   # fill the name from the fe proxy service ports
        nodePort: 30001
        containerPort: 8080
        port: 8080
EOF

  cmd="helm install -n starrocks starrocks $HELM_CHART_URL --create-namespace --timeout 60s"
  eval "$cmd -f /tmp/local-install-values.yaml" 1>/dev/null
}

# parseInput parses the input parameters
function parseInput() {
  while [ $# -gt 0 ]; do
    case $1 in
    -h | --help)
      cmdHelp
      exit
      ;;
    -k | --k8s-version)
      K8S_VERSION=$2
      shift 2
      ;;
    --helm-url)
      HELM_URL=$2
      shift 2
      ;;
    --kind-url)
      KIND_URL=$2
      shift 2
      ;;
    --kubectl-url)
      KUBECTL_URL=$2
      shift 2
      ;;
    --helm-chart-url)
      HELM_CHART_URL=$2
      shift 2
      ;;
    *)
      echo "Invalid option $1"
      cmdHelp
      exit 1
      ;;
    esac
  done

  checkPrerequisites
  (sudo kind get clusters | grep starrocks) || kindCreateCluster
  install && echo 'StarRocks is installed successfully!' || echo 'StarRocks is installed failed!'
}

parseInput "$@"
