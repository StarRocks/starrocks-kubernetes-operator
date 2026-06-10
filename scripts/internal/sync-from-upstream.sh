#! /bin/bash
#
# Workflow 2: sync ONE open-source (upstream) PR into the internal enterprise
# repository as its own PR. One upstream PR -> one internal branch -> one PR
# (a strict one-to-one mapping; the script never batches PRs together).
#
# Because the enterprise version renames many CRD fields (starrocks* ->
# celerdata*), the cherry-pick may conflict. The script stops cleanly on
# conflict; resolve it, run 'git cherry-pick --continue', then re-run this
# script with the same --pr to finalize (push + PR). git rerere is enabled so a
# given rename conflict only has to be resolved once.
#
# "Already synced" is detected by the upstream PR number appearing in the
# canonical internal main history (e.g. "(#747)") -- no state file is kept.
# PRs that need not be synced at all (already present via the fork, or
# deliberately skipped) can be recorded in scripts/internal/sync-from-upstream-ignore.txt
# so they stop showing up; use --ignore to add one.
#
# Usage:
#   scripts/internal/sync-from-upstream.sh [--since <sha>]            # list un-synced PRs
#   scripts/internal/sync-from-upstream.sh --pr <number> [--since <sha>]  # sync one PR
#   scripts/internal/sync-from-upstream.sh --ignore <number> [reason]  # mark as never-sync
#
#   --pr <number>     The upstream PR number to sync (one per run).
#   --ignore <number> Record the PR in sync-ignore.txt (with optional reason) so
#                     it is skipped from now on. Commit the file to share it.
#   --since <sha>     Oldest upstream commit to scan from. Default: the fork point,
#                     i.e. merge-base of the internal and upstream mains -- so only
#                     PRs merged upstream AFTER the enterprise fork are considered.
#
# Env:
#   AUTO_YES=1     Skip interactive confirmations (non-interactive runs).

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

SINCE=""
PR=""
IGNORE_PR=""
IGNORE_REASON=""
while [ $# -gt 0 ]; do
  case "$1" in
    --pr)
      PR="${2:-}"
      [ -n "$PR" ] || die "--pr requires a PR number"
      PR="${PR#\#}" # tolerate a leading '#'
      shift 2
      ;;
    --ignore)
      IGNORE_PR="${2:-}"
      [ -n "$IGNORE_PR" ] || die "--ignore requires a PR number"
      IGNORE_PR="${IGNORE_PR#\#}"
      shift 2
      IGNORE_REASON="$*" # rest of the args, if any, form the reason
      break
      ;;
    --since)
      SINCE="${2:-}"
      [ -n "$SINCE" ] || die "--since requires an upstream commit SHA"
      shift 2
      ;;
    -h | --help)
      usage "$0"
      exit 0
      ;;
    *) die "unknown argument: $1" ;;
  esac
done

require_cmd git gh
cd "$REPO_ROOT"

# is_ignored <prnum> — is this PR recorded in the never-sync list?
is_ignored() {
  [ -f "$SYNC_IGNORE_FILE" ] || return 1
  grep -qE "^[[:space:]]*#?$1([^0-9]|\$)" "$SYNC_IGNORE_FILE"
}

# --- --ignore <num> [reason]: record a PR as never-sync and exit -----------
if [ -n "$IGNORE_PR" ]; then
  if is_ignored "$IGNORE_PR"; then
    info "PR #$IGNORE_PR is already in $(basename "$SYNC_IGNORE_FILE"); nothing to do."
    exit 0
  fi
  [ -f "$SYNC_IGNORE_FILE" ] || printf '# Upstream PRs intentionally NOT synced. Format: "<pr>  # reason".\n' >"$SYNC_IGNORE_FILE"
  if [ -n "$IGNORE_REASON" ]; then
    printf '%s  # %s\n' "$IGNORE_PR" "$IGNORE_REASON" >>"$SYNC_IGNORE_FILE"
  else
    printf '%s\n' "$IGNORE_PR" >>"$SYNC_IGNORE_FILE"
  fi
  info "recorded PR #$IGNORE_PR in $(basename "$SYNC_IGNORE_FILE"). Commit the file to share it."
  exit 0
fi

enable_rerere

UPSTREAM_REF="$REMOTE_UPSTREAM/$BRANCH_OSS_MAIN"
INTERNAL_REF="$REMOTE_INTERNAL_UPSTREAM/$REMOTE_INTERNAL_MAIN"

info "fetching $REMOTE_UPSTREAM and $REMOTE_INTERNAL_UPSTREAM ..."
git fetch "$REMOTE_UPSTREAM"
git fetch "$REMOTE_INTERNAL_UPSTREAM"

# Canonical internal main subjects, captured ONCE. Used for already-synced
# detection. Captured into a variable (not piped to grep -q) on purpose: a
# `git log | grep -q` pipeline is racy under `set -o pipefail` because grep -q
# closes the pipe on first match and git log dies with SIGPIPE, making the whole
# pipeline "fail" even though a match was found -> everything looks un-synced.
INTERNAL_SUBJECTS="$(git log "$INTERNAL_REF" --format='%s')"

# pr_number_of <commit> — extract the trailing "#NNN" PR number from a subject.
pr_number_of() {
  git log -1 --format='%s' "$1" | grep -oE '#[0-9]+' | tail -1 | tr -d '#'
}

# is_synced <prnum> — is this PR already present in the canonical internal main?
# Matches "#N" as a whole token (paren-agnostic) so it catches both the upstream
# squash form "... (#N)" inherited at the fork AND sync commits titled
# "sync #N: ... (#<internal-pr>)" where the parenthesized number is the internal
# PR, not N.
is_synced() {
  printf '%s\n' "$INTERNAL_SUBJECTS" | grep -qE "#$1([^0-9]|\$)"
}

# Build the candidate commit list (oldest first) from the scan range. The
# default floor is the FORK POINT: merge-base(internal, upstream). Everything at
# or before it is already in the enterprise base (the internal repo was forked
# from the open-source repo there), so only post-fork PRs are real candidates.
if [ -n "$SINCE" ]; then
  git cat-file -e "${SINCE}^{commit}" 2>/dev/null || die "--since commit '$SINCE' not found"
  FLOOR="$SINCE"
  RANGE_DESC="${SINCE}..${UPSTREAM_REF}"
else
  FLOOR="$(git merge-base "$INTERNAL_REF" "$UPSTREAM_REF")" \
    || die "could not compute fork point (merge-base of $INTERNAL_REF and $UPSTREAM_REF)"
  RANGE_DESC="post-fork (${FLOOR:0:12}..${UPSTREAM_REF})"
fi
CANDIDATES_CMD=(git rev-list --reverse --no-merges "${FLOOR}..${UPSTREAM_REF}")

# --- No --pr: list the un-synced PRs and exit ------------------------------
if [ -z "$PR" ]; then
  info "scanning $RANGE_DESC for PRs not yet in $INTERNAL_REF ..."
  found=0
  while IFS= read -r c; do
    [ -n "$c" ] || continue
    num="$(pr_number_of "$c")"
    [ -n "$num" ] || continue
    if ! is_synced "$num" && ! is_ignored "$num"; then
      found=1
      printf '  #%-6s %s\n' "$num" "$(git log -1 --format='%s' "$c")"
    fi
  done < <("${CANDIDATES_CMD[@]}")
  [ "$found" -eq 1 ] || info "nothing to sync in range."
  echo
  info "run again with --pr <number> to sync one of these, or --ignore <number> to skip it."
  exit 0
fi

# --- --pr <num>: sync exactly that PR --------------------------------------
if is_synced "$PR"; then
  info "PR #$PR already present in $INTERNAL_REF; nothing to do."
  exit 0
fi
if is_ignored "$PR"; then
  die "PR #$PR is in $(basename "$SYNC_IGNORE_FILE") (never-sync). Remove it there first to sync anyway."
fi

# Locate the upstream commit for this PR within the scan range.
COMMIT=""
while IFS= read -r c; do
  [ -n "$c" ] || continue
  if [ "$(pr_number_of "$c")" = "$PR" ]; then
    COMMIT="$c"
    break
  fi
done < <("${CANDIDATES_CMD[@]}")
[ -n "$COMMIT" ] || die "PR #$PR not found in $RANGE_DESC; widen the scan with --since <older-sha>"

SUBJECT="$(git log -1 --format='%s' "$COMMIT")"
BRANCH="${BRANCH_INTERNAL_MAIN}-sync-${PR}"

# A cherry-pick stopped mid-flight must be finished by the user before we go on.
if [ -d "$REPO_ROOT/.git/sequencer" ] || [ -f "$REPO_ROOT/.git/CHERRY_PICK_HEAD" ]; then
  die "a cherry-pick is in progress. Resolve conflicts (rename starrocks* -> celerdata*),
     run 'git cherry-pick --continue', then re-run this script with --pr $PR."
fi

if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
  # Resume: the branch exists and no cherry-pick is in flight -> finalize it.
  info "found existing branch '$BRANCH' (no cherry-pick in progress) -> finalizing"
  git checkout -q "$BRANCH"
  require_clean_tree
else
  require_clean_tree
  info "PR #$PR -> $COMMIT"
  info "  $SUBJECT"
  info "creating branch '$BRANCH' off $INTERNAL_REF"
  git checkout -b "$BRANCH" "$INTERNAL_REF"
  # --empty=drop: if the change is already present, drop it instead of stalling.
  if ! git cherry-pick --empty=drop "$COMMIT"; then
    err "cherry-pick of PR #$PR stopped on a conflict."
    warn "Resolve it (rename starrocks* -> celerdata* where applicable),"
    warn "run 'git cherry-pick --continue', then re-run this script with --pr $PR to finalize."
    warn "To abort:  git cherry-pick --abort && git checkout $BRANCH_INTERNAL_MAIN && git branch -D $BRANCH"
    exit 1
  fi
fi

# Nothing landed (e.g. the only commit was dropped as already-present).
if [ -z "$(git rev-list "$INTERNAL_REF..$BRANCH")" ]; then
  warn "branch '$BRANCH' has no new commits over $INTERNAL_REF (already present?); nothing to push."
  exit 0
fi

echo
info "Branch '$BRANCH' is ready for PR #$PR:"
git --no-pager log --oneline "$INTERNAL_REF..$BRANCH" | sed 's/^/    /'

# Push the branch to your personal enterprise fork (not the canonical repo).
FORK_OWNER="$(fork_owner "$REMOTE_INTERNAL_ORIGIN")"
if ! confirm "Push '$BRANCH' to $REMOTE_INTERNAL_ORIGIN ($FORK_OWNER fork)?"; then
  warn "skipped push; branch '$BRANCH' left locally for inspection"
  exit 0
fi
git push -u "$REMOTE_INTERNAL_ORIGIN" "$BRANCH"

# Creating a PR is an outward-facing action: always confirm first.
if ! confirm "Create a DRAFT PR on $SLUG_INTERNAL ($FORK_OWNER:$BRANCH -> $REMOTE_INTERNAL_MAIN)?"; then
  warn "skipped PR creation; you can open it later with:"
  warn "  gh pr create --repo $SLUG_INTERNAL --base $REMOTE_INTERNAL_MAIN --head $FORK_OWNER:$BRANCH --draft"
  exit 0
fi

gh pr create \
  --repo "$SLUG_INTERNAL" \
  --base "$REMOTE_INTERNAL_MAIN" \
  --head "$FORK_OWNER:$BRANCH" \
  --draft \
  --title "sync #$PR: $SUBJECT" \
  --body "$(printf 'Sync of upstream PR #%s via scripts/internal/sync-from-upstream.sh\n\nUpstream commit: %s\n  %s\n' "$PR" "$COMMIT" "$SUBJECT")"

info "done."
