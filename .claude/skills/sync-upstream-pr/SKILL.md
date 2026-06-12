---
name: sync-upstream-pr
description: Use when syncing an open-source (upstream StarRocks) PR into the CelerData internal enterprise repo. Cherry-picks ONE upstream PR onto its own branch, runs the full shared check suite locally, fixes any issues, and ONLY THEN opens a one-to-one draft PR — so the PR is already validated when it appears. Trigger phrases: "sync upstream PR", "把 upstream 的 PR 同步到 internal", "sync #NNN into internal", "把社区 PR 合到企业版".
---

# Sync an upstream PR into the internal repo

Sync exactly **one** open-source PR into the CelerData internal repo as **one** branch and
**one** draft PR. **Check before you open the PR**, not after: cherry-pick → run the shared
check suite locally → fix → then push & open the PR.

**Read first:** `doc/internal/git_workflow_howto.md` — the "Enterprise naming conventions"
table and "Workflow 2".

> The same shared scripts run locally (this skill) and in CI:
> `scripts/internal/run-checks.sh` orchestrates `check-celerdata-conventions.sh` (starrocks*→
> celerdata* identifiers), `check-helm.sh` (lint + render with feature flags + parent-chart
> values drift via `create-parent-chart-values.sh`), and — for operator/CRD changes — `make test`
> (= generate/manifests/fmt/vet/UT, the `.github/pull_request_template.md` checklist) plus a
> generated-code/CRD drift check and golangci-lint. So "passes locally" means "will pass CI".

Run scripts from a branch that contains `scripts/internal/`. In steady state the sync branch
has it (it is cut from the canonical main, which carries the merged tooling). Until the
tooling PR is merged, run the scripts from `celerdata-internal-main`.

## Procedure

### 1. Pick the PR
```bash
git checkout celerdata-internal-main
bash scripts/internal/sync-from-upstream.sh            # lists un-synced upstream PRs
```
Choose a PR number `<N>` (or the user supplies it). If a listed PR turns out to need no sync
(content already in the enterprise base → its cherry-pick drops to a no-op), record it so it
stops showing up: `sync-from-upstream.sh --ignore <N> "<reason>"` (commit the updated
`sync-from-upstream-ignore.txt`).

### 2. Cherry-pick onto its branch — do NOT open the PR yet
```bash
bash scripts/internal/sync-from-upstream.sh --pr <N>
```
At the **"Push …?"** prompt answer **N**. This leaves `celerdata-internal-main-sync-<N>`
built locally with the cherry-picked commit, nothing pushed.

- **Conflict** → resolve applying `starrocks*`→`celerdata*` mappings, `git cherry-pick
  --continue`. `git rerere` replays known resolutions.
- **Upstream patch targets renamed/deleted files** (e.g. the legacy `kube-starrocks`
  templates) → the mechanical cherry-pick is a no-op; re-implement the change in the
  `kube-celerdata` tree as the sync commit, keeping `(#<N>)` in the subject.

### 3. Run the shared check suite locally and FIX until green
```bash
bash scripts/internal/run-checks.sh \
  celerdata-internal-upstream/main celerdata-internal-main-sync-<N>
```
It runs (based on what changed): the conventions check, the helm checks, and go
build/vet/test. **Also eyeball the diff** for things the regex misses (namespaces in prose,
example output, etc.) against the conventions table.

Fix every issue with a follow-up `fix(...)` commit on the sync branch and re-run until it
prints `ALL CHECKS PASSED`. (For helm changes, run from the sync branch so the working tree
is the synced content.)

### 4. Now open the draft PR
```bash
bash scripts/internal/sync-from-upstream.sh --pr <N>     # resumes the branch
```
Answer **Y** to push (to your fork) and **Y** to open the draft PR. The PR appears already
validated; CI re-runs the same shared checks.

## Never sync release artifacts (`index.yaml`, `doc/api.md`)
`index.yaml` (the Helm repo index) and `doc/api.md` are **generated per-repo** by
`scripts/artifacts.sh` — and that script is itself **different** on `main` vs
`celerdata-internal-main` (OSS publishes `kube-starrocks` from
`github.com/starrocks/…`; internal publishes `kube-celerdata` from
`github.com/celerdata/…`). They must **NEVER** be carried across repos in either direction.

- A PR whose changes are **only** these files (e.g. `[Chore] Update index.yaml for vX`) is never
  synced → `sync-from-upstream.sh --ignore <N> "release artifact, generated per-repo by artifacts.sh"`.
  (The internal edition regenerates its own at its next release; there is nothing else to carry over.)
- If a real PR **also** touches `index.yaml` / `doc/api.md`, do **not** carry the upstream copy.
  Cherry-pick the rest, then **regenerate the internal edition's artifacts**: `bash
  scripts/artifacts.sh <tag>` from `celerdata-internal-main` (which has the internal `artifacts.sh`
  → `kube-celerdata`, `github.com/celerdata/…`), commit the regenerated `index.yaml` / `doc/api.md`
  in place of the upstream ones, then continue the sync.

## Notes
- One upstream PR ⇒ one branch `celerdata-internal-main-sync-<N>` ⇒ one PR. Never batch.
- Fixes are always follow-up commits (never amend / force-push the sync commit).
- The scan only considers PRs merged upstream AFTER the enterprise fork (floor =
  `git merge-base` of the internal and upstream mains). "Already synced" = the PR number
  `#N` appears in `celerdata-internal-upstream/main`, so a PR shows un-synced until its sync
  PR merges into the canonical internal main.
- Scripts are bash-3.2 compatible (stock macOS `/bin/bash`).
