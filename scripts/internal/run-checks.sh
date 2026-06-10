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
#   - operator: go build + `make test` (= generate/manifests/fmt/vet/UT, per the
#     PR template) + a generated-code/CRD drift check               (if Go/*.mod/Makefile/apis/CRD changed)
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

# 3. Operator (Go / CRD) -- mirror the PR-template checklist (make generate,
#    make manifests, make test, golangci-lint). Triggered by Go sources, the Go
#    module, the Makefile, the API types, or the generated CRD manifests.
if echo "$FILES" | grep -qE '\.go$|^go\.(mod|sum)$|^Makefile$|^pkg/apis/|^config/crd/|^deploy/'; then
  if command -v go >/dev/null 2>&1; then
    run "go build" go build ./...
    # `make test` depends on manifests -> generate -> fmt -> vet -> envtest, so it
    # covers `make generate` and `make manifests` from the PR-template checklist and
    # runs the unit tests. (Downloads kubebuilder/envtest assets on first run.)
    run "make test" make test
    # Drift: if `make generate` / `make manifests` rewrote tracked files, the branch
    # is missing the regenerated output -- CI (action-make-test.yml / ci-manifests)
    # fails on exactly this. git status (vs HEAD) shows what regeneration changed.
    DRIFT="$(git status --porcelain -- pkg/apis config/crd deploy)"
    if [ -n "$DRIFT" ]; then
      err "generated-code / CRD drift after 'make test' -- run 'make generate' and 'make manifests', then commit:"
      printf '%s\n' "$DRIFT" | sed 's/^/    /'
      fail=1
    else
      info "no generated-code / CRD drift"
    fi
  else
    warn "go not installed; skipping go build / make test"
  fi
  if command -v golangci-lint >/dev/null 2>&1; then
    run "golangci-lint" golangci-lint run --timeout=30m ./...
  else
    warn "golangci-lint not installed; skipping (CI still runs it)"
  fi
else
  info "no operator (Go / CRD) changes; skipping operator checks"
fi

echo
if [ "$fail" -eq 0 ]; then
  info "ALL CHECKS PASSED"
else
  die "one or more checks FAILED (see above) -- fix before opening the PR"
fi
