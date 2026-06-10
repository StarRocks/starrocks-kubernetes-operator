# Git Workflow (Internal)

> **Internal document.** This file describes the relationship between the open-source
> repository, the internal enterprise repository, and the public-facing repository, plus
> the day-to-day branching, syncing, and release workflows. It is intended for CelerData
> maintainers. It lives under `doc/internal/`, which is in the `EXCLUDE` list of
> `scripts/sync-to-public-repo.sh`, so it is never published to the public repository.

## The three repositories

This project spans three GitHub repositories. The open-source project existed first; the
enterprise (internal) version was forked from it later. The main difference between them is
in the **CRD definitions** — many fields that are prefixed with `starrocks` in the
open-source version are prefixed with `celerdata` in the enterprise version.

| # | Repository | Purpose | Local branch | Canonical remote |
|---|------------|---------|--------------|------------------|
| 1 | [`StarRocks/starrocks-kubernetes-operator`](https://github.com/StarRocks/starrocks-kubernetes-operator) | Open-source project | `main` | `upstream` |
| 2 | [`CelerData/celerdata-kubernetes-operator-internal`](https://github.com/CelerData/celerdata-kubernetes-operator-internal) | Internal enterprise version (full source) | `celerdata-internal-main` | `celerdata-internal-upstream` |
| 3 | [`celerdata/celerdata-kubernetes-operator`](https://github.com/celerdata/celerdata-kubernetes-operator) | Public-facing subset of the enterprise version (Helm charts, deploy YAML, docs) | — (separate clone) | — (not a remote) |

### Fork model — remotes

Both projects use the same fork model: you **push branches to your personal fork**
(`*origin`) and **open PRs against the canonical org repo** (`upstream` / `*upstream`).
You never push branches directly to the canonical repos.

| Project | Canonical remote (PR target) | Personal-fork remote (push branches here) |
|---------|------------------------------|-------------------------------------------|
| Open source | `upstream` → `StarRocks/...` | `origin` → `yandongxiao/starrocks-kubernetes-operator` |
| Enterprise | `celerdata-internal-upstream` → `CelerData/...-internal` | `celerdata-internal-origin` → `yandongxiao/celerdata-kubernetes-operator-internal` |

### Branch ↔ remote mapping

- Local `main` tracks **`upstream/main`** — the open-source `main`.
- Local `celerdata-internal-main` tracks **`celerdata-internal-upstream/main`** — the
  internal enterprise `main`.

### The public repository (#3) is a separate checkout

Repository #3 is **not** a git remote of this checkout. It is a separate clone that must
live in the same parent directory as this repository:

```
<parent>/
├── starrocks-kubernetes-operator/        # this checkout (open source + internal remotes)
└── celerdata-kubernetes-operator/        # the public repo (#3)
```

Content is pushed into it by copying files (see [Release](#workflow-3-release), not by
git push/pull. `scripts/sync-to-public-repo.sh` resolves the target as
`<this-repo>/../celerdata-kubernetes-operator`.

## One-time setup

If the remotes are not yet configured:

```bash
# Open source: canonical + your personal fork
git remote add upstream https://github.com/StarRocks/starrocks-kubernetes-operator.git
git remote add origin   https://github.com/<you>/starrocks-kubernetes-operator.git

# Enterprise: canonical + your personal fork
git remote add celerdata-internal-upstream https://github.com/CelerData/celerdata-kubernetes-operator-internal.git
git remote add celerdata-internal-origin   https://github.com/<you>/celerdata-kubernetes-operator-internal.git

git fetch --all
```

Clone the public repository next to this one so the release script can find it:

```bash
cd ..
git clone https://github.com/celerdata/celerdata-kubernetes-operator.git
```

## Automation

Local helper scripts under `scripts/internal/` automate the mechanical parts of the three
workflows below. They live in `scripts/internal/` precisely because that directory is in the
`EXCLUDE` list of `scripts/sync-to-public-repo.sh`, so this internal tooling is never
published to the public repository.

| Workflow | Script |
|----------|--------|
| 2 — sync upstream → internal | `scripts/internal/sync-from-upstream.sh` |
| 1 (step 3) — port internal → open source | `scripts/internal/port-to-oss.sh` |

Releasing (Workflow 3) is **not** wrapped in a helper script — it is a low-frequency,
human-supervised process done with the existing `scripts/artifacts.sh` plus a manually
created GitHub release. See Workflow 3 below.

Two things are true of both sync/port scripts:

- **Conflicts are expected and not automated away.** Because the enterprise version renames
  many CRD fields (`starrocks*` ↔ `celerdata*`), cross-repo cherry-picks conflict. The scripts
  do the mechanical work and **stop cleanly on conflict** so you resolve it by hand, then you
  re-run `git cherry-pick --continue`. The scripts enable `git rerere`, so once you resolve a
  given rename conflict, git replays that resolution automatically on later syncs.
- **PR / push steps always prompt for confirmation** before running.
  Set `AUTO_YES=1` to skip the prompts in non-interactive runs.

## Enterprise naming conventions (`starrocks*` → `celerdata*`)

A **clean cherry-pick can still be semantically wrong**: upstream uses `starrocks*`
identifiers that have `celerdata*` equivalents here, and a textual match produces no conflict
(e.g. `{{ template "starrockscluster.name" . }}` refers to a template that does not exist in
the celerdata charts, so it silently renders empty). Every synced PR must be reviewed against
these mappings:

| Upstream (open source) | Enterprise (this fork) | Where |
|---|---|---|
| `StarRocksCluster` / `StarRocksWarehouse` | `CelerDataCluster` / `CelerDataWarehouse` | CRD kind |
| `starrocks.com/v1` | `celerdata.com/v1` | apiVersion / API group |
| `app.starrocks.io/…` | `app.celerdata.io/…` | label / annotation keys |
| `{{ template "starrockscluster.…" }}` | `{{ template "celerdatacluster.…" }}` | Helm template helpers |
| `kube-starrocks` chart | `kube-celerdata` chart | Helm chart references |
| `-n starrocks` / `namespace: starrocks` | `-n celerdata` / `namespace: celerdata` | docs & examples |
| datadog `"service":"starrocks"` | `"service":"celerdata"` | helm annotations |

**Intentionally left as `starrocks`** (do NOT translate): on-disk binary paths such as
`/opt/starrocks/…`, upstream image repository names, and the legacy `kube-starrocks/` chart
tree itself.

## Pre-PR checks (shared local ⇄ CI)

The same check scripts run **locally before a PR is opened** and **in CI** on every PR, so
"passes locally" means "will pass CI". Run the whole suite against your branch before pushing:

```bash
scripts/internal/run-checks.sh celerdata-internal-upstream/main <branch>
```

`run-checks.sh` dispatches, based on what the branch changed, to the shared scripts that CI
also invokes:

| Check | Script | CI workflow | Runs when |
|---|---|---|---|
| Enterprise naming conventions | `check-celerdata-conventions.sh` | `action-check-conventions.yml` (advisory — never blocks) | every PR |
| Helm lint + render (flags on) | `check-helm.sh` | `action-helm-template.yml` | `helm-charts/**` changed |
| Go build / vet / test | `make test` | `action-make-test.yml` | `pkg/**` changed |
| golangci-lint | — | `action-golangci-lint.yml` | every PR |

The convention check flags high-signal `starrocks*` identifiers; `check-helm.sh` renders the
charts with `serviceMonitor` / `datadog.log` **enabled** so conditional templates are actually
exercised (a plain default render hides errors like an undefined `starrockscluster.name`
helper). Namespace / prose / example cases still need a human eye against the table above.

## Workflow 1: Develop a feature in the internal repository

This is the common case — new enterprise work happens on the internal repository.

1. **Branch off `celerdata-internal-main`.** The branch name **must contain
   `celerdata-internal`** to mark it as internal-based work.

   ```bash
   git checkout celerdata-internal-main
   git pull celerdata-internal-upstream main
   git checkout -b celerdata-internal-<short-description>
   ```

2. **Develop, commit, push to your fork, and open a PR** targeting the canonical internal
   `main` (`celerdata-internal-upstream/main`).

   ```bash
   git push celerdata-internal-origin celerdata-internal-<short-description>
   # then open the PR against CelerData/celerdata-kubernetes-operator-internal:main
   ```

3. **Optionally cherry-pick into the open-source project.** After the internal PR merges,
   selected changes are cherry-picked onto the open-source `main`. Only changes that are
   appropriate for the open-source version belong here — be mindful of the
   `celerdata`-vs-`starrocks` CRD naming differences when porting.

   ```bash
   git checkout main
   git pull upstream main
   git checkout -b <feature-on-open-source>
   git cherry-pick <commit-sha>           # resolve naming conflicts as needed
   git push origin <feature-on-open-source>
   # open a PR against StarRocks/starrocks-kubernetes-operator:main
   ```

   **Automation:** step 3 is scripted. Pass the internal commit SHA(s) to port:

   ```bash
   scripts/internal/port-to-oss.sh <commit-sha> [<commit-sha> ...]
   ```

   It branches off `upstream/main`, cherry-picks the commits, and — after you confirm —
   pushes to `origin` and opens the PR against `StarRocks/starrocks-kubernetes-operator:main`.

## Workflow 2: Sync an upstream (open-source) contribution into internal

When another contributor merges a PR into the open-source project, that change must be
fully synced into the internal repository via its own PR.

1. **Get the upstream change locally.**

   ```bash
   git checkout main
   git pull upstream main                 # now main has the merged contribution
   ```

2. **Create a sync branch off `celerdata-internal-main`** and bring the upstream commit(s)
   over. Name the branch so it is clearly a sync branch and contains `celerdata-internal`
   (e.g. `celerdata-internal-main-sync-<pr-number>` or
   `celerdata-internal-sync-<commit-sha>`).

   ```bash
   git checkout celerdata-internal-main
   git pull celerdata-internal-upstream main
   git checkout -b celerdata-internal-main-sync-<pr-number>
   git cherry-pick <upstream-commit-sha>  # adapt starrocks* -> celerdata* where needed
   ```

3. **Push to your fork and open a PR** against `celerdata-internal-upstream/main`.

   ```bash
   git push celerdata-internal-origin celerdata-internal-main-sync-<pr-number>
   # open the PR against CelerData/celerdata-kubernetes-operator-internal:main
   ```

**Automation:** scripted as a strict **one-to-one mapping** — one upstream PR becomes one
internal branch and one internal PR. The script never batches multiple PRs together.

```bash
# 1. List upstream PRs merged AFTER the enterprise fork that are not yet synced:
scripts/internal/sync-from-upstream.sh

# 2. Sync ONE of them. Creates celerdata-internal-main-sync-<pr> off the canonical
#    internal main, cherry-picks just that PR's commit, then — after you confirm —
#    pushes to your fork and opens a draft PR:
scripts/internal/sync-from-upstream.sh --pr 741
```

**Fork floor:** the enterprise repo was forked from the open-source repo, so every PR that
predates the fork is already in the enterprise base. The scan therefore starts at the **fork
point**, computed automatically as `git merge-base celerdata-internal-upstream/main
upstream/main` — no hardcoded SHA. Only PRs merged upstream after the fork are candidates.
Override the floor with `--since <upstream-sha>` if ever needed.

**Already-synced detection** is by PR number in the canonical internal main history (matching
`#NNN` as a token, so it catches both the upstream squash form `… (#NNN)` inherited at the
fork and sync commits titled `sync #NNN: … (#<internal-pr>)`). No state file; selectively
skipped PRs are never lost.

**Never-sync list.** Some post-fork PRs need not be synced at all — e.g. their content is
already present in the enterprise base independently, so the cherry-pick is a no-op and no
`#NNN` marker will ever appear. Record these so they stop showing up:

```bash
scripts/internal/sync-from-upstream.sh --ignore 741 "already in the enterprise base"
```

This appends to `scripts/internal/sync-from-upstream-ignore.txt` (one `<pr>  # reason` per
line); commit the file to share the decision. Listed/`--pr` runs skip ignored PRs; to sync one
anyway, remove its line from the file.

On conflict the script stops; resolve it, run `git cherry-pick --continue`, then re-run with
the same `--pr` to finalize (the branch is resumed → push + PR). Branches are cut from
`celerdata-internal-upstream/main`, so a sync PR never mixes in unrelated local commits.

## Workflow 3: Release

Releases are cut from the **internal** repository, then published to the public repository.

1. **Release in the internal repository.** Tag/release on `celerdata-internal-upstream/main`
   and build/push the operator container image to the private image registry.

2. **Sync the public-facing content** into repository #3 using the script:

   ```bash
   ./scripts/sync-to-public-repo.sh
   ```

   The script copies the following items from this checkout into
   `../celerdata-kubernetes-operator`, removing the old copy first to ensure a clean sync:

   - `doc/` (`doc/internal/` is excluded — internal docs stay private)
   - `examples/`
   - `helm-charts/`
   - `deploy/`
   - `scripts/` (`scripts/internal/` is excluded — internal tooling stays private)
   - `README.md`, `README_ZH-CN.md`

3. **Commit and release in the public repository.** In
   `../celerdata-kubernetes-operator`, review the synced changes, commit, push, and cut the
   corresponding public release.

**Automation:** releasing is intentionally **not** scripted into a single command — it is a
low-frequency, human-supervised process. The signed Helm artifacts are produced with the
existing helper, and the GitHub release is created by hand:

```bash
bash scripts/artifacts.sh <vX.Y.Z>   # build + sign Helm chart artifacts under artifacts/
# then create the GitHub release manually and attach the artifacts
```

> TODO: the exact release steps (which repo the GitHub release lives in, operator image
> build/push, and whether `scripts/sync-to-public-repo.sh` is run as part of a release) will
> be filled in here from the next real release.

## Quick reference

| I want to… | Base branch | Push to (fork) → PR target |
|------------|-------------|----------------------------|
| Build an enterprise feature | `celerdata-internal-main` | `celerdata-internal-origin` → `CelerData/...-internal:main` |
| Port an enterprise change to open source | `upstream/main` | `origin` → `StarRocks/...:main` |
| Pull an open-source PR into enterprise | `celerdata-internal-main` | `celerdata-internal-origin` → `CelerData/...-internal:main` |
| Publish a release | — | manual: `scripts/artifacts.sh <tag>` + GitHub release (see Workflow 3) |
