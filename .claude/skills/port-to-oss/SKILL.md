---
name: port-to-oss
description: Use when porting (backporting) a CelerData internal/enterprise PR to the open-source StarRocks project. Cherry-picks ONE internal PR into a worktree off upstream/main, VERIFIES the celerData*->starrocks* renames + field consistency (even with no merge conflict) and runs the PR-template checks, then opens one OSS PR. Trigger phrases: "port to oss", "backport to community", "把企业版 PR 迁移到社区版", "port internal #N to open source".
---

# Port an internal PR to open source

Port exactly **one** internal (enterprise) PR to the open-source StarRocks project as **one**
branch and **one** PR. **Verify before you open the PR**, and verify *even when the cherry-pick
applies cleanly*. This is the reverse of `sync-upstream-pr`: here `celerData*` becomes `starrocks*`.

**Read first:** `doc/internal/git_workflow_howto.md` (the "Enterprise naming conventions" table —
read it as a **reverse** map, `celerData* → starrocks*`) and `.github/pull_request_template.md` (the
checks every PR must pass).

> **A clean cherry-pick can still be semantically wrong.** Internal charts live at `kube-celerdata/…`
> with `celerData*` keys; OSS charts live at `kube-starrocks/…` with `starrocks*` keys. Git either
> (a) stops with a modify/delete conflict on the differing path — obvious; or (b) **silently applies
> a naming-neutral hunk to the OSS file via rename detection** (e.g. `waitForFullRollout: false`
> slotted into the `starrocksCluster` block) — NOT obvious. Both need step 3. No conflict ≠ correct.

The script does the work in a **git worktree** cut from `upstream/main`, so your main checkout (and
this tooling) is never switched away and stays runnable.

## Procedure

### 1. Pick the internal PR
Normally port **after** the internal PR merges, so its squash commit is on
`celerdata-internal-upstream/main` (subject ends `(#N)`). Run everything from a tooling-bearing
checkout (e.g. `celerdata-internal-main`).

### 2. Cherry-pick into a worktree — do NOT push yet
```bash
scripts/internal/port-to-oss.sh --pr <N>          # normal (PR merged to internal main)
scripts/internal/port-to-oss.sh --commit <sha>    # escape hatch (not yet merged)
```
It resolves the source commit, creates worktree `port-oss-<short-sha>` (under
`/tmp/celerdata-port-oss/…`) off `upstream/main`, and cherry-picks. Answer **N** at the push prompt.

- **Conflict** (modify/delete on a `kube-celerdata` / `celerdatacluster.yaml` path) → resolve in the
  worktree dir the script prints: `git rm` the leftover celerdata-path file, re-apply the change by
  hand to the OSS-named file (`kube-starrocks` / `starrockscluster.yaml`), **write an OSS-context
  commit message** (it becomes the PR body), `git -C <worktree> cherry-pick --continue`, then re-run
  the script (from the main checkout) to finalize. `git rerere` replays known rename resolutions.
- **No conflict** → still do step 3.

### 3. VERIFY in the worktree — the heart of this skill
Even with no conflict, in the worktree (`cd` into the path the script printed):

```bash
git show --stat HEAD | grep -i celerdata        # expect: no kube-celerdata / celerdata paths
git show HEAD | grep -nE 'celerData|celerdata\.com|kube-celerdata|app\.celerdata' || echo clean
```
Translate every piece per the reverse table:

| Internal (source) | Open source (must become) |
|---|---|
| `CelerDataCluster` / `CelerDataWarehouse` | `StarRocksCluster` / `StarRocksWarehouse` |
| `celerdata.com/v1`, `app.celerdata.io/…` | `starrocks.com/v1`, `app.starrocks.io/…` |
| `{{ template "celerdatacluster.…" }}` | `{{ template "starrockscluster.…" }}` |
| `.Values.celerDataCluster` / `celerDataFeSpec` / `celerDataFeProxySpec` | `.Values.starrocksCluster` / `starrocksFESpec` / `starrocksFeProxySpec` |
| `kube-celerdata` chart | `kube-starrocks` chart |

**Field-consistency (catches the silent rename-detection case).** A values key only works if the
template reads the same key and the parent chart forwards it — prove the chain end to end:
```bash
helm lint helm-charts/charts/kube-starrocks
helm template t helm-charts/charts/kube-starrocks --set <parent.path.to.field>=<value> | grep <field>
```
**Run the PR-template checks** (`.github/pull_request_template.md`): for operator/CRD changes
`make generate`, `make manifests`, `make test`, `golangci-lint run` — and **commit any regenerated
output** (CI fails on drift); for helm changes run `bash scripts/create-parent-chart-values.sh` so
the parent `values.yaml` is in sync. (These are exactly what `scripts/internal/run-checks.sh
<base> <head>` runs — use it.)

### 4. Finalize — push + open (or update) the PR
Re-run from the main checkout:
```bash
scripts/internal/port-to-oss.sh --pr <N>     # resumes the worktree: push + PR
```
The script: titles the OSS PR after the ported commit; builds the body from the **internal PR's
description**, mechanically translated `celerData*`→`starrocks*` with `Fixes/Closes #N` lines
stripped; is **repeatable** (`gh pr edit`s if the PR already exists); and cross-links by commenting
`Ported to open source: <url>` on the internal PR (idempotent). Then it offers to remove the worktree.

**Then eyeball the OSS PR body** — the translation is mechanical and does NOT fix: inline refs to
internal-only PRs/issues (e.g. "added in #5" → the **upstream** PR #739), internal tooling paths
(`scripts/internal/check-helm.sh` → `helm lint`/`helm template`), or internal-only Notes. Fix those
on the OSS PR (re-running overwrites the body, so make these the last touch-up).

## Notes
- One internal PR ⇒ one `port-oss-<short-sha>` worktree/branch ⇒ one OSS PR. Never batch.
- Provenance lives in the internal-PR comment, not in the public PR body.
- Work happens in a worktree, so the main checkout (with `scripts/internal/`) is never switched away —
  no "the OSS branch lacks the tooling" problem. `git worktree remove` when done.
