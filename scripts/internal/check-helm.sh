#! /bin/bash
#
# Helm chart checks for the enterprise charts. Runnable locally (pre-PR, via
# run-checks.sh) and from CI (.github/workflows/action-helm-template.yml).
#
# Renders with DEFAULT values AND with key feature flags enabled, so conditional
# templates (ServiceMonitor, Datadog logs) are actually exercised -- a plain
# default render leaves those blocks unrendered and hides template errors such
# as an undefined "starrockscluster.name" helper.
#
# Usage: scripts/internal/check-helm.sh
#
# Exit status: 0 = all renders/lints clean, non-zero otherwise.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

require_cmd helm
cd "$REPO_ROOT"

KUBE_CELERDATA="helm-charts/charts/kube-celerdata"
WAREHOUSE="helm-charts/charts/warehouse"

# CRD api-versions some templates gate on (e.g. ServiceMonitor). helm template
# does not know about cluster CRDs, so they must be supplied explicitly.
API_VERSIONS=(--api-versions monitoring.coreos.com/v1)

info "helm lint $KUBE_CELERDATA"
helm lint "$KUBE_CELERDATA"

info "helm template $KUBE_CELERDATA (default values)"
helm template celerdata-test "$KUBE_CELERDATA" >/dev/null

info "helm template $KUBE_CELERDATA (ServiceMonitor + Datadog logs enabled)"
helm template celerdata-test "$KUBE_CELERDATA" \
  --set celerdata.metrics.serviceMonitor.enabled=true \
  --set celerdata.datadog.log.enabled=true \
  --set celerdata.datadog.log.enableMultilineLogParsing=true \
  "${API_VERSIONS[@]}" >/dev/null

if [ -d "$WAREHOUSE" ]; then
  info "helm lint $WAREHOUSE"
  helm lint "$WAREHOUSE"
  info "helm template $WAREHOUSE"
  helm template warehouse-test "$WAREHOUSE" >/dev/null
fi

# Parent-chart values drift. The PR-template checklist requires running
# create-parent-chart-values.sh after editing a subchart's values.yaml, because
# kube-celerdata/values.yaml is GENERATED from the operator + celerdata subchart
# values. Regenerate and fail if it differs from what is committed -- a common
# miss when only the subchart values were edited.
PARENT_VALUES="$KUBE_CELERDATA/values.yaml"
if [ -f scripts/create-parent-chart-values.sh ] && [ -f "$PARENT_VALUES" ]; then
  info "regenerating parent chart values (create-parent-chart-values.sh) and checking for drift"
  bash scripts/create-parent-chart-values.sh
  if ! git diff --quiet -- "$PARENT_VALUES"; then
    err "$PARENT_VALUES is out of sync with the subchart values."
    err "Run 'bash scripts/create-parent-chart-values.sh' and commit the result. Diff:"
    git --no-pager diff -- "$PARENT_VALUES" | sed 's/^/    /'
    die "parent chart values drift"
  fi
  info "parent chart values in sync"
fi

info "helm checks passed."
