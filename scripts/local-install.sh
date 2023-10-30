#!/usr/bin/env bash

VERSION=
K8S_VERSION="v1.23.4"

UPGRADE=false
CLEAN=false
KIND=true

# Pre-requisites:
# - kubectl
# - helm
# - docker
# - kind

# Check if the binary is installed
# If not, return false, else return true
function checkBinary() {
  if command -v "$1" &>/dev/null; then
    echo "The binary $1 is installed"
    return 0
  else
    echo "The binary $1 is not installed"
    return 1
  fi
}

function mustInstalled() {
  if ! command -v "$1" &>/dev/null; then
    echo "The binary $1 is not installed"
    exit
  else
    echo "The binary $1 is installed"
  fi
}

function installHelm() {
  echo "installing helm"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -LO https://get.helm.sh/helm-v3.12.3-linux-amd64.tar.gz &&
    tar -zxvf helm-v3.12.3-linux-amd64.tar.gz && sudo mv linux-amd64/helm /usr/local/bin/helm

  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -LO https://get.helm.sh/helm-v3.12.3-linux-arm64.tar.gz &&
    tar -zxvf helm-v3.12.3-linux-arm64.tar.gz && sudo mv linux-arm64/helm /usr/local/bin/helm

  # MacOS Intel
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -LO https://get.helm.sh/helm-v3.12.3-darwin-amd64.tar.gz &&
    tar -zxvf helm-v3.12.3-darwin-amd64.tar.gz && sudo mv darwin-amd64/helm /usr/local/bin/helm

  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -LO https://get.helm.sh/helm-v3.12.3-darwin-arm64.tar.gz &&
    tar -zxvf helm-v3.12.3-darwin-arm64.tar.gz && sudo mv darwin-arm64/helm /usr/local/bin/helm
}

function installKubectl() {
  echo "Installing kubectl"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"

  # MacOS Intel Macs
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"

  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"

  chmod +x ./kubectl
  sudo mv ./kubectl /usr/local/bin/kubectl
  sudo chown root: /usr/local/bin/kubectl
}

function installKind() {
  echo "Installing kind"
  # Linux AMD64 / x86_64
  [[ $(uname -m) = x86_64 && $(uname) = Linux ]] &&
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
  # Linux ARM64
  [[ $(uname -m) = aarch64 && $(uname) = Linux ]] &&
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-arm64
  # MacOS Intel
  [[ $(uname -m) = x86_64 && $(uname) = Darwin ]] &&
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-amd64
  # MacOS M1 / ARM
  [[ $(uname -m) = arm64 && $(uname) = Darwin ]] &&
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-arm64

  # install kind
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
}

function checkPrerequisites() {
  mustInstalled docker

  # If kind is not installed, install kind
  if $KIND && ! checkBinary kind; then
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

function cmdHelp() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "       --clean"
  echo "  -h,  --help"
  echo "  -k,  --k8s-version <K8S_VERSION>, specify the version of k8s to install, default is $K8S_VERSION"
  echo "  -u,  --upgrade, equals to helm upgrade"
}

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

function install() {
  helm repo add starrocks-community https://starrocks.github.io/starrocks-kubernetes-operator

  echo "Update helm repo"
  helm repo update starrocks-community

  cmd="helm"

  if $UPGRADE; then
    cmd="$cmd upgrade"
  else
    cmd="$cmd install"
  fi

  cmd="$cmd -n starrocks starrocks starrocks-community/kube-starrocks"

  if [ -n "$VERSION" ]; then
    cmd="$cmd --version $VERSION"
  fi

  if $UPGRADE; then
    cmd="$cmd --timeout 30s"
  else
    cmd="$cmd --create-namespace --timeout 60s"
  fi

  if $UPGRADE; then
    echo "Upgrading starrocks"
  else
    echo "Installing starrocks"
  fi

  cat <<EOF >/tmp/local-install-values.yaml
operator:
  starrocksOperator:
    image:
      repository: starrocks/operator
      tag: v1.8.5

starrocks:
  starrocksFESpec:
    image:
      repository: starrocks/fe-ubuntu
      tag: 3.1.2
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 500m
        memory: 1Gi
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
      tag: 3.1.2
    resources:
      limits:
        cpu: 2
        memory: 4Gi
      requests:
        cpu: 500m
        memory: 1Gi
  starrocksFeProxySpec:
    enabled: true
    resources:
      requests:
        cpu: 100m
        memory: 200Mi
    service:
      type: NodePort
      ports:
      - name: http-port
        nodePort: 30001
        containerPort: 8080
        port: 8080
EOF

  eval "$cmd -f /tmp/local-install-values.yaml" 1>/dev/null
}

function clean() {
  echo "Cleaning kind cluster"
  kind delete cluster --name starrocks
}

function parseInput() {
  while [ $# -gt 0 ]; do
    case $1 in
    -h | --help)
      cmdHelp
      exit
      ;;
    -u | --upgrade)
      UPGRADE=true
      shift
      ;;
    -v | --version)
      VERSION=$2
      shift 2
      ;;
    -k | --k8s-version)
      K8S_VERSION=$2
      shift 2
      ;;
    --clean)
      CLEAN=true
      shift
      ;;
    *)
      echo "Invalid option $1"
      cmdHelp
      exit 1
      ;;
    esac
  done

  if $CLEAN; then
    clean
    exit
  fi

  if $UPGRADE; then
    install && echo 'StarRocks is installed successfully!' || echo 'StarRocks is installed failed!'
    exit
  fi

  checkPrerequisites
  (kind get clusters | grep starrocks) || kindCreateCluster
  install && echo 'StarRocks is installed successfully!' || echo 'StarRocks is installed failed!'
}

parseInput "$@"
