# GEMINI.md - CelerData Kubernetes Operator

## Project Overview

This repository is the distribution and deployment center for the **CelerData Kubernetes Operator**, which originated from the **StarRocks Kubernetes Operator** project. It has been transitioned into a new upstream for the **CelerData Enterprise Edition**.

The project is currently in its early stages as a standalone enterprise version (with initial commits) focused on rebranding and strengthening brand influence by replacing StarRocks identifiers with **CelerData**.

### Key Technologies
- **Kubernetes Operator**: Level 2 operator for lifecycle management.
- **Helm**: Primary method for installation and configuration.
- **Go**: The operator itself is written in Go (source is managed in an internal enterprise repository).
- **CelerData/StarRocks**: The target sub-second MPP OLAP database.

## Project Evolution & Status

- **Origin**: Forked/Derived from the StarRocks Kubernetes Operator.
- **Current Focus**: Transitioning the codebase to the **CelerData** brand. This involves:
    - Updating CRDs and API groups (e.g., `celerdata.com`).
    - Renaming components and Helm charts.
    - Strengthening the brand identity for the Enterprise Edition.
- **Commit History**: Currently a fresh repository with initial commits dedicated to this rebranding and upstream setup.

## Directory Structure

- `/helm-charts`: Contains Helm charts for the operator (`operator`), CelerData cluster (`celerdata`), and the combined `kube-celerdata` chart.
- `/deploy`: Contains static Kubernetes manifests and Custom Resource Definitions (CRDs).
    - `celerdata.com_celerdataclusters.yaml`: Main cluster resource CRD.
    - `celerdata.com_celerdatawarehouses.yaml`: Warehouse resource CRD.
    - `operator.yaml`: Manifest for deploying the operator (generated via scripts).
- `/examples/celerdata`: Sample YAML configurations for various Enterprise deployment scenarios.
- `/doc`: Documentation and "How-to" guides, including enterprise-specific features.
- `/bin` & `/cmd`: Pre-built binaries for the CelerData operator.
- `/scripts`: Utility scripts for packaging, manifest generation, and release management.

## Core Concepts

### Custom Resources (CRDs)
- **CelerDataCluster**: The primary resource for defining FE, BE, and CN configurations.
- **CelerDataWarehouse**: Manages elastic warehouse resources.

### Architecture
1. **Frontend (FE)**: Metadata management and query planning.
2. **Backend (BE)**: Data storage and computation (Shared-Nothing).
3. **Compute Node (CN)**: Stateless compute for elastic scaling (Shared-Data).

## Building and Running

### Installation

#### 1. Via Helm (Recommended)
```bash
helm install kube-celerdata ./helm-charts/charts/kube-celerdata
```

#### 2. Via Manifests
```bash
# Apply CelerData CRDs
kubectl apply -f deploy/celerdata.com_celerdataclusters.yaml
# Deploy CelerData Operator
kubectl apply -f deploy/operator.yaml
```

### Scripts
- `scripts/operator.sh`: Syncs Helm templates to the `deploy/operator.yaml` manifest.
- `scripts/artifacts.sh`: Handles the packaging of Enterprise Helm charts and provider signatures.
- `scripts/gen-api-reference-docs.sh`: Generates API documentation from the internal enterprise source.

## Development Conventions
- **Branding Consistency**: All new contributions must use the **CelerData** branding.
- **Distribution Focused**: This repo distributes stable enterprise releases.
- **Helm-Centric**: Manifests in `/deploy` should be kept in sync with Helm templates via provided scripts.
