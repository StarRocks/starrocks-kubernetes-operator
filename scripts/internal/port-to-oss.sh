#! /bin/bash
#
# Workflow 1 (step 3): port selected internal commits back to the open-source
# project and open a PR against StarRocks/starrocks-kubernetes-operator.
#
# Cherry-picks may conflict on the celerdata* -> starrocks* renames; the script
# stops cleanly on conflict for manual resolution. git rerere is enabled.
#
# Usage:
#   scripts/internal/port-to-oss.sh <commit-sha> [<commit-sha> ...]
#
# Env:
#   AUTO_YES=1     Skip interactive confirmations.

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

if [ $# -eq 0 ] || [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage "$0"
  [ $# -eq 0 ] && exit 1 || exit 0
fi
SHAS=("$@")

require_cmd git gh
cd "$REPO_ROOT"

# Validate the requested commits exist before touching anything.
for sha in "${SHAS[@]}"; do
  git cat-file -e "${sha}^{commit}" 2>/dev/null || die "commit not found: $sha"
done

require_clean_tree
enable_rerere

# Derive the open-source fork owner so the cross-fork PR head is "owner:branch".
FORK_OWNER="$(fork_owner "$REMOTE_ORIGIN")"

info "fetching $REMOTE_UPSTREAM ..."
git fetch "$REMOTE_UPSTREAM"

BRANCH="port-oss-$(today_stamp)"
if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
  die "branch '$BRANCH' already exists; delete it or rename the in-progress port first"
fi

info "creating branch '$BRANCH' off $REMOTE_UPSTREAM/$BRANCH_OSS_MAIN"
git checkout -b "$BRANCH" "$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN"

info "cherry-picking ${#SHAS[@]} commit(s): ${SHAS[*]}"
# --empty=drop skips commits that are already present on the open-source side.
if ! git cherry-pick --empty=drop "${SHAS[@]}"; then
  err "cherry-pick stopped on a conflict."
  warn "Resolve it now (remember: rename celerdata* -> starrocks* where applicable),"
  warn "then run:  git cherry-pick --continue"
  warn "To abort:  git cherry-pick --abort && git checkout -"
  exit 1
fi

echo
info "Branch '$BRANCH' ready with ${#SHAS[@]} cherry-picked commit(s)."
git --no-pager log --oneline "$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN".."$BRANCH" | sed 's/^/    /'

if ! confirm "Push '$BRANCH' to $REMOTE_ORIGIN ($FORK_OWNER fork)?"; then
  warn "skipped push; branch '$BRANCH' left locally for inspection"
  exit 0
fi
git push -u "$REMOTE_ORIGIN" "$BRANCH"

# Creating a PR is an outward-facing action: always confirm first.
if ! confirm "Create a PR on $SLUG_UPSTREAM ($FORK_OWNER:$BRANCH -> $BRANCH_OSS_MAIN)?"; then
  warn "skipped PR creation; you can open it later with:"
  warn "  gh pr create --repo $SLUG_UPSTREAM --base $BRANCH_OSS_MAIN --head $FORK_OWNER:$BRANCH"
  exit 0
fi

gh pr create \
  --repo "$SLUG_UPSTREAM" \
  --base "$BRANCH_OSS_MAIN" \
  --head "$FORK_OWNER:$BRANCH" \
  --title "port: internal -> open source ($(today_stamp))" \
  --body "Ported from the internal repo via scripts/internal/port-to-oss.sh.

Source commits: ${SHAS[*]}"

info "done."
