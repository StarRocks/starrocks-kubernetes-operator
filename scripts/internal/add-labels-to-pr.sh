#! /bin/bash
#
# Label internal-repo PRs with a release/version tag plus a category label
# derived from the PR title (feature / enhancement / bugfix / chore /
# documentation).
#
# This is the internal-repo counterpart of scripts/add-labels-to-pr.sh. The key
# difference: it ALWAYS targets the canonical internal repo ($SLUG_INTERNAL) via
# an explicit `gh --repo`, so it can be run from any checkout (or worktree)
# without accidentally labeling whatever repo `gh` infers from the cwd.
#
# Usage:
#   scripts/internal/add-labels-to-pr.sh <version-tag> <pr> [<pr> ...]
#   e.g. scripts/internal/add-labels-to-pr.sh v1.11.5 12 14 15 16
#
# Env:
#   (none)

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

[ $# -ge 2 ] || {
  usage "$0"
  die "need a version tag and at least one PR number"
}

VERSION_TAG="$1"
shift
PR_NUMBERS=("$@")

require_cmd gh

# Category keywords matched (case-insensitively) against the PR title. Kept in
# sync with scripts/add-labels-to-pr.sh.
KEYWORDS=("feature" "enhancement" "bugfix" "chore" "documentation")

# ensure_label <name> — create the label in the internal repo if it is missing.
# Idempotent: `gh label create` errors when the label already exists, which is
# fine here.
ensure_label() {
  gh label create "$1" --repo "$SLUG_INTERNAL" >/dev/null 2>&1 || true
}

ensure_label "$VERSION_TAG"

for PR_NUMBER in "${PR_NUMBERS[@]}"; do
  PR_NUMBER="${PR_NUMBER#\#}" # tolerate a leading '#'
  info "labeling $SLUG_INTERNAL#$PR_NUMBER with '$VERSION_TAG'"
  gh pr edit "$PR_NUMBER" --repo "$SLUG_INTERNAL" --add-label "$VERSION_TAG"

  PR_TITLE="$(gh pr view "$PR_NUMBER" --repo "$SLUG_INTERNAL" --json title -q '.title' | tr '[:upper:]' '[:lower:]')"
  for KEYWORD in "${KEYWORDS[@]}"; do
    if [[ "$PR_TITLE" == *"$KEYWORD"* ]]; then
      ensure_label "$KEYWORD"
      gh pr edit "$PR_NUMBER" --repo "$SLUG_INTERNAL" --add-label "$KEYWORD"
      info "  + $KEYWORD"
    fi
  done
done

info "done."
