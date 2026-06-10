#! /bin/bash
#
# Check that changed files do not carry upstream "starrocks*" identifiers that
# should be their "celerdata*" equivalents in the enterprise version.
#
# A clean cherry-pick from the open-source repo can still be semantically wrong:
# e.g. {{ template "starrockscluster.name" . }} produces NO text conflict but
# refers to a template that does not exist in the celerdata charts, so it
# silently renders empty. This catches that class of problem.
#
# It only flags HIGH-SIGNAL identifiers that almost always have a celerdata
# counterpart -- not every mention of the word "starrocks" (binary paths like
# /opt/starrocks, image repositories, and prose are intentionally left alone).
#
# Usage:
#   scripts/internal/check-celerdata-conventions.sh <base-ref> [<head-ref>]
#
#   <base-ref>  Compare against this ref to find changed files (e.g. the
#               canonical internal main).
#   <head-ref>  Tree-ish to scan (default: HEAD).
#
# Exit status: 0 = clean, 1 = violations found.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

BASE="${1:-}"
HEAD_REF="${2:-HEAD}"
if [ -z "$BASE" ] || [ "$BASE" = "-h" ] || [ "$BASE" = "--help" ]; then
  usage "$0"
  [ -z "$BASE" ] && exit 2 || exit 0
fi

cd "$REPO_ROOT"

# High-signal upstream identifiers that should be celerdata* in this fork.
#   starrockscluster.<x>   -> celerdatacluster.<x>   (Helm template helpers)
#   starrockswarehouse.<x> -> celerdatawarehouse.<x>
#   StarRocksCluster       -> CelerDataCluster       (CRD kind)
#   StarRocksWarehouse     -> CelerDataWarehouse
#   starrocks.com/         -> celerdata.com/         (API group / annotations)
#   app.starrocks.io/      -> app.celerdata.io/       (label/annotation keys)
#   kube-starrocks         -> kube-celerdata          (chart references)
#   -n starrocks           -> -n celerdata            (kubectl namespace in docs)
#   namespace: starrocks   -> namespace: celerdata    (yaml namespace in docs)
PATTERN='starrockscluster\.|starrockswarehouse\.|StarRocksCluster|StarRocksWarehouse|starrocks\.com/|app\.starrocks\.io/|kube-starrocks|-n starrocks|namespace:[[:space:]]+starrocks'

# Changed files between base and the scanned ref. Exclude paths that legitimately
# spell out starrocks* tokens and are NOT the branded chart/CRD/code content this
# check targets: the internal tooling (scripts/internal/, doc/internal/, .claude/)
# and CI infra (.github/ — workflow files reference upstream repo names, and this
# very check's own docs mention "starrockscluster.name" as an example).
FILES="$(git diff --name-only "$BASE" "$HEAD_REF" | grep -vE '^(scripts/internal/|doc/internal/|\.claude/|\.github/)' || true)"
if [ -z "$FILES" ]; then
  info "no relevant changed files between $BASE and $HEAD_REF; nothing to check."
  exit 0
fi

# git grep over the changed files at the scanned ref. word-style identifiers.
HITS="$(echo "$FILES" | tr '\n' '\0' | xargs -0 git grep -nE "$PATTERN" "$HEAD_REF" -- 2>/dev/null || true)"

if [ -n "$HITS" ]; then
  warn "enterprise-convention check FAILED: upstream starrocks* identifiers found in changed files."
  warn "These usually need their celerdata* equivalents (e.g. starrockscluster.name -> celerdatacluster.name):"
  echo "$HITS" | sed 's/^/   /'
  exit 1
fi

info "enterprise-convention check passed: no high-signal starrocks* identifiers in changed files."
