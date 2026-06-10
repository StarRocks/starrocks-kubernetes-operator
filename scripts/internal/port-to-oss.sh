#! /bin/bash
#
# Workflow 1 (step 3): port ONE internal PR back to the open-source project and
# open ONE PR against StarRocks/starrocks-kubernetes-operator.
#
# Strict one-to-one: one internal PR -> one OSS branch -> one OSS PR. No batching.
#
# The work happens in a dedicated git WORKTREE cut from upstream/main, so the main
# checkout (and this tooling / the port-to-oss skill) is never switched away and
# stays runnable -- the OSS base has no scripts/internal/. Resolve cherry-pick
# conflicts in that worktree, run `git -C <worktree> cherry-pick --continue`, then
# re-run this script (from the main checkout) to finalize. git rerere is enabled so
# a given celerData* -> starrocks* rename conflict replays automatically next time.
#
# Usage:
#   scripts/internal/port-to-oss.sh --pr <internal-pr-number>   # normal: PR already merged to internal main
#   scripts/internal/port-to-oss.sh --commit <sha>             # escape hatch: not-yet-merged commit
#
# Env:
#   AUTO_YES=1                     Skip interactive confirmations.
#   PORT_OSS_WORKTREE_ROOT=<dir>   Where to place port worktrees (default: /tmp/celerdata-port-oss).

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

# --- args: exactly one of --pr / --commit -----------------------------------
PR=""
COMMIT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --pr) PR="${2:-}"; shift 2 ;;
    --commit) COMMIT="${2:-}"; shift 2 ;;
    -h | --help) usage "$0"; exit 0 ;;
    *) die "unknown argument: $1 (use --pr <n> or --commit <sha>)" ;;
  esac
done
if { [ -n "$PR" ] && [ -n "$COMMIT" ]; } || { [ -z "$PR" ] && [ -z "$COMMIT" ]; }; then
  usage "$0"
  die "specify exactly one of --pr <internal-pr-number> or --commit <sha>"
fi

require_cmd git gh
cd "$REPO_ROOT"

enable_rerere
FORK_OWNER="$(fork_owner "$REMOTE_ORIGIN")"

info "fetching $REMOTE_UPSTREAM and $REMOTE_INTERNAL_UPSTREAM ..."
git fetch "$REMOTE_UPSTREAM"
git fetch "$REMOTE_INTERNAL_UPSTREAM"

# --- resolve the source commit + the originating internal PR ----------------
if [ -n "$PR" ]; then
  INT_PR="$PR"
  # The squash-merge commit on internal main carries "(#PR)" in its subject.
  SHA="$(git log "$REMOTE_INTERNAL_UPSTREAM/$REMOTE_INTERNAL_MAIN" \
    --grep="(#${PR})" --fixed-strings --format='%H' -1 || true)"
  [ -n "$SHA" ] || die "no squash commit containing \"(#${PR})\" on ${REMOTE_INTERNAL_UPSTREAM}/${REMOTE_INTERNAL_MAIN}. If PR #${PR} is not merged yet, use --commit <sha>."
else
  SHA="$COMMIT"
  git cat-file -e "${SHA}^{commit}" 2>/dev/null || die "commit not found: $SHA"
  # Best-effort: find the internal PR this commit belongs to (for body + cross-link).
  INT_PR="$(gh pr list --repo "$SLUG_INTERNAL" --search "$SHA" --state all \
    --json number --jq '.[0].number // empty' 2>/dev/null || true)"
fi

SHORT="$(git rev-parse --short "$SHA")"
BRANCH="port-oss-${SHORT}"
WT_ROOT="${PORT_OSS_WORKTREE_ROOT:-/tmp/celerdata-port-oss}"
WT="${WT_ROOT}/${SHORT}"

info "source commit: $SHA${INT_PR:+  (internal PR #$INT_PR)}"

# --- create or resume the worktree, cherry-pick the commit ------------------
# Is the branch already attached to a worktree? -> resume there.
EXISTING_WT="$(git worktree list --porcelain | awk -v b="refs/heads/$BRANCH" '
  $1=="worktree"{p=$2} $1=="branch"&&$2==b{print p}')"

if [ -n "$EXISTING_WT" ]; then
  WT="$EXISTING_WT"
  info "resuming in existing worktree: $WT"
elif git show-ref --verify --quiet "refs/heads/$BRANCH"; then
  info "attaching a worktree to existing branch '$BRANCH' at $WT"
  mkdir -p "$WT_ROOT"
  git worktree add "$WT" "$BRANCH"
else
  mkdir -p "$WT_ROOT"
  info "creating worktree '$BRANCH' at $WT (off $REMOTE_UPSTREAM/$BRANCH_OSS_MAIN)"
  git worktree add -b "$BRANCH" "$WT" "$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN"
  info "cherry-picking $SHA in the worktree"
  if ! git -C "$WT" cherry-pick --empty=drop "$SHA"; then
    err "cherry-pick stopped on a conflict in the worktree."
    warn "Resolve it there (rename celerData* -> starrocks*, kube-celerdata -> kube-starrocks paths/keys),"
    warn "write an open-source-context commit message (it becomes the PR body), then:"
    warn "    git -C $WT cherry-pick --continue"
    warn "and re-run THIS script (from the main checkout) to finalize:"
    warn "    scripts/internal/port-to-oss.sh ${PR:+--pr $PR}${COMMIT:+--commit $COMMIT}"
    warn "To abort:  git -C $WT cherry-pick --abort && git worktree remove --force $WT && git branch -D $BRANCH"
    exit 1
  fi
fi

# A cherry-pick may still be in progress on resume (conflict not yet continued).
if [ -f "$(git -C "$WT" rev-parse --git-path CHERRY_PICK_HEAD)" ]; then
  die "a cherry-pick is still in progress in $WT. Resolve it and run 'git -C $WT cherry-pick --continue', then re-run this script."
fi

if [ -z "$(git -C "$WT" log --oneline "$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN".."$BRANCH")" ]; then
  die "branch '$BRANCH' has no commits over $REMOTE_UPSTREAM/$BRANCH_OSS_MAIN (cherry-pick empty or dropped?)"
fi

echo
info "Branch '$BRANCH' ready (worktree $WT):"
git -C "$WT" --no-pager log --oneline "$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN".."$BRANCH" | sed 's/^/    /'

# --- VERIFY: even a clean cherry-pick can be semantically wrong -------------
# Reverse-direction sanity: no enterprise identifiers should survive in the diff.
RESIDUAL="$(git -C "$WT" show HEAD | grep -nE 'celerData|celerdata\.com|kube-celerdata|app\.celerdata' || true)"
if [ -n "$RESIDUAL" ]; then
  warn "the ported diff still contains enterprise identifiers -- translate celerData* -> starrocks*:"
  printf '%s\n' "$RESIDUAL" | sed 's/^/    /'
  warn "(see the port-to-oss skill: also run helm template / make test in the worktree to verify field consistency)"
fi

# --- title + body -----------------------------------------------------------
# Title = ported commit subject. Body = the internal PR's (richer) description,
# mechanically translated celerData* -> starrocks* with auto-close lines stripped.
# Falls back to the commit message body if there is no internal PR. The mechanical
# translation does NOT fix inline cross-refs to internal-only PRs/issues, internal
# tooling paths, or internal-only Notes -- eyeball the rendered PR (see the skill).
TITLE="$(git -C "$WT" log -1 --format=%s)"
if [ -n "${INT_PR:-}" ]; then
  PR_BODY="$(gh pr view "$INT_PR" --repo "$SLUG_INTERNAL" --json body --jq .body | sed -E \
    -e 's/CelerDataCluster/StarRocksCluster/g' \
    -e 's/CelerDataWarehouse/StarRocksWarehouse/g' \
    -e 's/celerDataFeProxySpec/starrocksFeProxySpec/g' \
    -e 's/celerDataFeSpec/starrocksFESpec/g' \
    -e 's/celerDataBeSpec/starrocksBeSpec/g' \
    -e 's/celerDataCnSpec/starrocksCnSpec/g' \
    -e 's/celerDataCluster/starrocksCluster/g' \
    -e 's/kube-celerdata/kube-starrocks/g' \
    -e 's/celerdatacluster/starrockscluster/g' \
    -e 's/celerdatawarehouse/starrockswarehouse/g' \
    -e 's#celerdata\.com#starrocks.com#g' \
    -e 's#app\.celerdata\.io#app.starrocks.io#g' \
    -e 's/celerdata/starrocks/g' \
    | sed -E '/^(Fixes|Closes|Resolves) #[0-9]+\.?[[:space:]]*$/d')"
else
  PR_BODY="$(git -C "$WT" log -1 --format=%b)"
fi

# --- push + open/update the PR (repeatable) ---------------------------------
if ! confirm "Push '$BRANCH' to $REMOTE_ORIGIN ($FORK_OWNER fork)?"; then
  warn "skipped push; worktree left at $WT"
  exit 0
fi
git -C "$WT" push -u "$REMOTE_ORIGIN" "$BRANCH"

EXISTING_PR="$(gh pr list --repo "$SLUG_UPSTREAM" --head "$BRANCH" --state all \
  --json number --jq '.[0].number // empty' 2>/dev/null || true)"

if [ -n "$EXISTING_PR" ]; then
  if ! confirm "PR $SLUG_UPSTREAM#$EXISTING_PR already exists; update its title + body?"; then
    exit 0
  fi
  gh pr edit "$EXISTING_PR" --repo "$SLUG_UPSTREAM" --title "$TITLE" --body "$PR_BODY"
  OSS_URL="$(gh pr view "$EXISTING_PR" --repo "$SLUG_UPSTREAM" --json url --jq .url)"
  info "updated $OSS_URL"
else
  if ! confirm "Create a PR on $SLUG_UPSTREAM ($FORK_OWNER:$BRANCH -> $BRANCH_OSS_MAIN)?"; then
    warn "skipped PR creation; open it later with:"
    warn "  gh pr create --repo $SLUG_UPSTREAM --base $BRANCH_OSS_MAIN --head $FORK_OWNER:$BRANCH"
    exit 0
  fi
  OSS_URL="$(gh pr create \
    --repo "$SLUG_UPSTREAM" \
    --base "$BRANCH_OSS_MAIN" \
    --head "$FORK_OWNER:$BRANCH" \
    --title "$TITLE" \
    --body "$PR_BODY")"
  info "created $OSS_URL"
fi

# --- cross-link: comment the OSS PR on the originating internal PR ----------
if [ -z "${INT_PR:-}" ]; then
  warn "no originating internal PR resolved; skipped cross-link comment"
elif gh pr view "$INT_PR" --repo "$SLUG_INTERNAL" --json comments \
       --jq '.comments[].body' 2>/dev/null | grep -qF "$OSS_URL"; then
  info "internal PR $SLUG_INTERNAL#$INT_PR already links $OSS_URL"
else
  gh pr comment "$INT_PR" --repo "$SLUG_INTERNAL" --body "Ported to open source: $OSS_URL"
  info "commented on internal PR $SLUG_INTERNAL#$INT_PR -> $OSS_URL"
fi

echo
warn "Eyeball the OSS PR body for internal-only refs the mechanical translation can't fix:"
warn "  inline #N to internal PRs/issues, scripts/internal/ paths, internal-only Notes."
if confirm "Remove the worktree $WT now? (say N to keep it for inspection)"; then
  git worktree remove "$WT"
  info "removed worktree $WT (branch '$BRANCH' kept)"
fi
info "done."
