#! /bin/bash
#
# Sync the public-facing subset of this repository into the sibling public
# clone (celerdata-kubernetes-operator), then strip internal-only paths so the
# enterprise tooling never leaks. This script itself lives under
# scripts/internal/ and is therefore removed from the published copy.
#
# Usage: scripts/internal/sync-to-public-repo.sh
#
# Exit status: 0 = sync completed, non-zero if the target repo is missing.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

# The public repository must be cloned alongside this checkout.
if [ ! -d "$PUBLIC_REPO_PATH" ]; then
  die "target directory $PUBLIC_REPO_PATH does not exist. Please clone the public repository in the same parent directory."
fi

# List of items to sync (directories or files), relative to the repo root.
ITEMS=("doc" "examples" "helm-charts" "deploy" "scripts" "README.md" "README_ZH-CN.md")

# Internal-only paths that must NOT be published to the public repository.
# Paths are relative to the target repository root.
EXCLUDE=("doc/internal" "scripts/internal")

info "Starting synchronization from $REPO_ROOT to $PUBLIC_REPO_PATH..."

for item in "${ITEMS[@]}"; do
  if [ -e "$REPO_ROOT/$item" ]; then
    info "Processing: $item"

    # Remove the old item in target to ensure a clean copy.
    rm -rf "$PUBLIC_REPO_PATH/$item"

    # Copy the item recursively.
    cp -R "$REPO_ROOT/$item" "$PUBLIC_REPO_PATH/"

    info "Successfully synced $item"
  else
    warn "Source $REPO_ROOT/$item does not exist, skipping."
  fi
done

# Remove internal-only files that should never be published.
for excluded in "${EXCLUDE[@]}"; do
  if [ -e "$PUBLIC_REPO_PATH/$excluded" ]; then
    rm -rf "$PUBLIC_REPO_PATH/$excluded"
    info "Excluded internal-only file: $excluded"
  fi
done

info "---------------------------------------------------"
info "Synchronization completed successfully!"
info "Target: $PUBLIC_REPO_PATH"
