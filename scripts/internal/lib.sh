#! /bin/bash
#
# Shared helpers for the internal git-workflow automation scripts.
#
# This file is meant to be sourced, not executed directly. It is kept under
# scripts/internal/ which is excluded from scripts/sync-to-public-repo.sh, so
# none of this tooling leaks into the public repository.

# Resolve the repository root via the existing common.sh helper.
# shellcheck source=/dev/null
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/common.sh"
REPO_ROOT="$(
  cd "$(dirname "${BASH_SOURCE[0]}")/../.."
  pwd
)"

# --- Constants --------------------------------------------------------------

# Git remotes (see `git remote -v`). Both projects use a fork model: branches
# are pushed to your personal fork (*_ORIGIN) and PRs target the canonical
# org repo (*_UPSTREAM).
REMOTE_UPSTREAM="upstream"                              # StarRocks/...               (open source, canonical)
REMOTE_ORIGIN="origin"                                  # yandongxiao/...             (open source, personal fork)
REMOTE_INTERNAL_UPSTREAM="celerdata-internal-upstream"  # CelerData/...-internal      (enterprise, canonical)
REMOTE_INTERNAL_ORIGIN="celerdata-internal-origin"      # yandongxiao/...-internal    (enterprise, personal fork)

# GitHub repository slugs (PR targets — the canonical org repos).
SLUG_UPSTREAM="StarRocks/starrocks-kubernetes-operator"
SLUG_INTERNAL="CelerData/celerdata-kubernetes-operator-internal"

# Branch names.
BRANCH_OSS_MAIN="main"                        # tracks upstream/main
BRANCH_INTERNAL_MAIN="celerdata-internal-main" # tracks celerdata-internal/main
REMOTE_INTERNAL_MAIN="main"                   # branch name on the internal remote

# Public-facing repository, a sibling clone of this checkout.
PUBLIC_REPO_PATH="${REPO_ROOT}/../celerdata-kubernetes-operator"

# Upstream PRs intentionally NOT synced (already present via the fork, or
# deliberately skipped). One "<pr-number>  # reason" per line. Travels with the
# tooling so the decision is shared and auditable.
SYNC_IGNORE_FILE="${REPO_ROOT}/scripts/internal/sync-from-upstream-ignore.txt"

# State file recording the last upstream commit synced into internal.
# Lives under scripts/internal/ so it is committed to the internal repo only.
UPSTREAM_SYNC_STATE="${REPO_ROOT}/scripts/internal/.upstream-sync-state"

# --- Output helpers ---------------------------------------------------------

info() { printf '\033[1;34m[info]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[warn]\033[0m %s\n' "$*" >&2; }
err()  { printf '\033[1;31m[error]\033[0m %s\n' "$*" >&2; }
die()  { err "$*"; exit 1; }

# --- Guards -----------------------------------------------------------------

# require_cmd <cmd> [<cmd> ...] — fail if any command is missing.
require_cmd() {
  local missing=0 c
  for c in "$@"; do
    if ! command -v "$c" >/dev/null 2>&1; then
      err "required command not found: $c"
      missing=1
    fi
  done
  [ "$missing" -eq 0 ] || die "please install the missing command(s) above and retry"
}

# require_clean_tree — fail if there are uncommitted changes to TRACKED files.
# Untracked files (.claude/, bin/, coverage.data, ...) are ignored: they do not
# affect cherry-picks and would otherwise trip the guard on branches cut from a
# main that predates this repo's .gitignore entries.
require_clean_tree() {
  if [ -n "$(git -C "$REPO_ROOT" status --porcelain --untracked-files=no)" ]; then
    die "working tree has uncommitted changes to tracked files; commit or stash first"
  fi
}

# require_branch <branch> — fail unless currently on <branch>.
require_branch() {
  local want="$1" cur
  cur="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD)"
  [ "$cur" = "$want" ] || die "expected to be on branch '$want' but on '$cur'"
}

# confirm "<message>" — prompt y/N before an outward-facing or destructive step.
# Returns 0 on yes, 1 on no. Honors AUTO_YES=1 for non-interactive runs.
confirm() {
  local msg="$1" reply
  if [ "${AUTO_YES:-0}" = "1" ]; then
    info "AUTO_YES=1 -> proceeding: $msg"
    return 0
  fi
  printf '\033[1;36m[confirm]\033[0m %s [y/N] ' "$msg"
  read -r reply
  case "$reply" in
    y | Y | yes | YES) return 0 ;;
    *) return 1 ;;
  esac
}

# enable_rerere — let git remember conflict resolutions so the recurring
# starrocks<->celerdata rename conflicts only need to be solved once.
enable_rerere() {
  git -C "$REPO_ROOT" config rerere.enabled true
  info "git rerere enabled (conflict resolutions will be remembered)"
}

# today_stamp — date suffix for generated branch names (YYYYMMDD).
today_stamp() { date +"%Y%m%d"; }

# fork_owner <remote> — extract the GitHub <owner> from a remote's URL, for
# building cross-fork PR heads like "<owner>:<branch>". Handles https/ssh forms
# with or without a trailing ".git".
fork_owner() {
  local url
  url="$(git -C "$REPO_ROOT" remote get-url "$1")" \
    || die "no such remote: $1"
  local owner
  owner="$(printf '%s' "$url" | sed -E 's#.*[/:]([^/]+)/[^/]+(\.git)?$#\1#')"
  [ -n "$owner" ] || die "could not derive fork owner from $1 URL: $url"
  printf '%s' "$owner"
}

# usage <file> — print the leading comment block (after the shebang) of a script
# as help text, stopping at the first non-comment line.
usage() {
  awk 'NR==1 { next } /^#/ { sub(/^#[[:space:]]?/, ""); print; next } { exit }' "$1"
}
