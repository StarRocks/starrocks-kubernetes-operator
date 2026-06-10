#! /bin/bash
#
# Pre-PR check suite: run the same kinds of checks CI runs, on the changes
# between <base-ref> and the head, BEFORE opening a PR. The sync skill invokes
# this so a PR is already validated when it is opened.
#
# Each check is delegated to a shared script that CI also calls, so local and CI
# stay in lock-step:
#   - enterprise conventions  -> check-celerdata-conventions.sh   (always)
#   - helm charts             -> check-helm.sh                    (if helm-charts/** changed)
#   - go build / vet / tests                                       (if Go/*.mod/Makefile changed)
#   - golangci-lint                                                (if installed; matches CI lint)
#
# Usage: scripts/internal/run-checks.sh <base-ref> [<head-ref>]
#
# Exit status: 0 = everything passed, 1 = at least one check failed.

set -uo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

BASE="${1:-}"
HEAD_REF="${2:-HEAD}"
if [ -z "$BASE" ] || [ "$BASE" = "-h" ] || [ "$BASE" = "--help" ]; then
  usage "$0"
  [ -z "$BASE" ] && exit 2 || exit 0
fi

cd "$REPO_ROOT"
FILES="$(git diff --name-only "$BASE" "$HEAD_REF")"
fail=0
run() { info "== $1 =="; shift; "$@" || fail=1; }

# 1. Enterprise naming conventions -- always.
run "enterprise conventions" bash scripts/internal/check-celerdata-conventions.sh "$BASE" "$HEAD_REF"

# 2. Helm charts -- only if any chart changed.
if echo "$FILES" | grep -q '^helm-charts/'; then
  run "helm charts" bash scripts/internal/check-helm.sh
else
  info "no helm-charts/** changes; skipping helm checks"
fi

# 3. Go -- only if Go sources / module / Makefile changed.
if echo "$FILES" | grep -qE '\.go$|^go\.(mod|sum)$|^Makefile$'; then
  if command -v go >/dev/null 2>&1; then
    run "go build" go build ./...
    run "go vet" go vet ./...
    run "go test ./pkg/..." go test ./pkg/... -timeout 120s
  else
    warn "go not installed; skipping go build/vet/test"
  fi
  if command -v golangci-lint >/dev/null 2>&1; then
    run "golangci-lint" golangci-lint run --timeout=30m ./...
  else
    warn "golangci-lint not installed; skipping (CI still runs it)"
  fi
else
  info "no Go changes; skipping go checks"
fi

echo
if [ "$fail" -eq 0 ]; then
  info "ALL CHECKS PASSED"
else
  die "one or more checks FAILED (see above) -- fix before opening the PR"
fi
