---
name: test-release
description: Use when verifying a CelerData/StarRocks operator release before tagging it — testing what changed since the last release against the local kind cluster, validating Helm charts, CRDs, and operator runtime behavior. Trigger phrases: "测试这个 release", "跑发版测试", "verify v1.11.x", "test the release", "验证企业版/社区版".
---

# Test an operator release against the local kind cluster

Verify the **next release** of the operator by testing only what changed since the last
release. Produces a reviewable **test plan** first, then executes it and writes a **report**.
Works for the enterprise branch (`celerdata-internal-main`) and the open-source branch
(`main`).

**Read first:** `doc/internal/git_workflow_howto.md` — the three-repo model and the
"Enterprise naming conventions" (`starrocks*` ↔ `celerdata*`) table. A clean Helm render can
still be semantically wrong if a `starrocks*` helper name leaks into the celerdata charts.

> Deliverable model: **plan → user reviews → execute → report.** Stop after the plan and let
> the user approve it before touching the cluster.

## Step 0 — Pick the branch (only required input)

Ask the user which branch to test if not given. Everything else derives from this table:

| Dimension | Enterprise `celerdata-internal-main` | Community `main` |
|---|---|---|
| upstream remote | `celerdata-internal-upstream` | `upstream` |
| namespace | `celerdata` | `starrocks` |
| chart dir | `helm-charts/charts/kube-celerdata` | `helm-charts/charts/kube-starrocks` |
| operator image | `celerdata/operator:<tag>` | `starrocks/operator:<tag>` |
| diff floor | brand commit `6ba6ae6` | latest tag (`git describe --tags --abbrev=0`) |

**Diff-floor rationale:** enterprise tags (`v1.11.x`) were inherited from upstream *before* the
rebrand (`git tag --contains 6ba6ae6` is empty), so enterprise release work lives after the
brand commit. Use the latest post-rebrand tag once one exists; until then the floor is
`6ba6ae6`.

The two branches have **different chart trees** (`kube-celerdata` vs `kube-starrocks`), so you
must be on the target branch. Use a git worktree so the user's active checkout is never
disturbed:

```bash
git fetch <remote>
git worktree add /tmp/test-release-<branch> <remote>/main
cd /tmp/test-release-<branch>
```

## Step 1 — Compute the diff and enumerate PRs

```bash
FLOOR=6ba6ae6                      # enterprise; or: FLOOR=$(git describe --tags --abbrev=0) for main
RANGE="$FLOOR..<remote>/main"
git log --oneline "$RANGE"         # PRs, grouped by the squash-merge (#NNN) marker
git log --stat --patch "$RANGE"    # what each PR actually changed (use this, no helper script)
```

Bucket each PR's changed files:

| Bucket | Path | Test dimensions it triggers |
|---|---|---|
| Helm chart | `helm-charts/**` | Helm correctness + (often) runtime |
| Operator logic | `pkg/**`, `cmd/**` | Runtime |
| CRD / config | `config/**`, `pkg/apis/**` | CRD correctness + runtime |
| Docs only | `doc/**`, `*.md`, `examples/**` | skip (note it as "no test needed") |

**Look up the upstream PR for synced changes.** Most enterprise PRs are synced from open source
— the title carries the upstream PR number, e.g. `sync #747: …`, `… (#751) (#10)`,
`Fix-738 … (#739)`. The enterprise commit message and diff alone often miss the *motivation*
and *edge cases* the upstream author documented. For every PR whose title references an upstream
number `<U>`, fetch its context and fold it into the test cases:

```bash
gh pr view <U> --repo StarRocks/starrocks-kubernetes-operator --json number,title,body,url
```

The upstream body frequently names the real acceptance criterion the local diff hides — e.g.
upstream #747 revealed the `group` label exists because the **official StarRocks Grafana
dashboards require it**; #751 documented FE=log4j vs BE/CN=glog patterns and that user-supplied
`log_processing_rules` take priority. Cite the upstream PR URL and the criterion you derived from
it in the test plan.

## Step 2 — Write the test plan, then STOP for review

For each non-docs PR, reason over its diff and write concrete cases against the three
dimensions below. Cases are **generated from the diff**, not a fixed matrix — each must name an
explicit pass criterion.

1. **Helm chart correctness** — `helm lint` + `helm template` with the PR-relevant feature
   flags on, then assert the *specific* rendered output the PR touched.
   - Enterprise: reuse `bash scripts/internal/check-helm.sh` (renders kube-celerdata with
     ServiceMonitor + Datadog logs enabled).
   - Community: `helm template <chart> --set <relevant flags>` and grep the rendered output.
   - **Always** check the `starrocks*`↔`celerdata*` table — a textual match is not enough.
2. **CRD correctness** — CRD applies/upgrades cleanly (`kubectl apply` / `helm upgrade` has no
   schema error); any new field appears in the installed CRD schema
   (`kubectl get crd <name> -o yaml | grep <field>`).
3. **Operator runtime correctness** — deploy to the namespace; operator pod healthy;
   FE/BE/CN/feproxy pods reach `Running`; the PR's behavior is observable live (a pod env var,
   service annotation, configmap value, probe change, etc.).

Write to `doc/internal/test/<YYYY-MM-DD>-<branch>-test-plan.md` (under `doc/internal/`, which is
in the `scripts/internal/sync-to-public-repo.sh` EXCLUDE list — never published). Include: diff range, PR list,
per-PR cases with pass criteria and dimensions. **Then stop and ask the user to review the plan
before executing.**

## Step 3 — Execute (after plan approval)

```bash
# Save the current base values BEFORE any uninstall (enterprise example):
helm get values kube-celerdata -n celerdata > /tmp/base-values-celerdata.yaml
```

Community values: derive from the saved enterprise values by `starrocks*`↔`celerdata*`
substitution, validate with `helm template kube-starrocks -f <derived>`, and fall back to the
chart's default `values.yaml` where the structure diverges.

```bash
# Build + load the operator image (kind cluster is named "starrocks"):
docker build -t <image>:<tag> .
kind load docker-image --name starrocks <image>:<tag>
```

**FE/BE/CN images: reuse what is already cached in kind.** These data-plane images are NOT the
test target. Keep the base values' image refs and set `imagePullPolicy: IfNotPresent` so they are
not re-pulled — `docker exec starrocks-control-plane crictl images` lists what is already on the
node. (You can also pin a known-cached tag in the override values.) Only the *operator* image is
built and loaded.

**If `docker build` can't reach docker hub** (e.g. `golang:1.22` / `static-debian11` pull times
out), don't block — the binary already builds locally (the static-check step proves it). Overlay
it onto a prior operator image, which carries the correct runtime base:
```bash
ARCH=$(docker exec starrocks-control-plane uname -m)   # aarch64 -> arm64, x86_64 -> amd64
CGO_ENABLED=0 GOOS=linux GOARCH=<arch> go build -mod=vendor -o /tmp/celerdata-operator ./cmd/main.go
printf 'FROM <prior-image>\nCOPY celerdata-operator /celerdata-operator\nCMD ["/celerdata-operator"]\n' > /tmp/Dockerfile.overlay
docker build --platform linux/<arch> -f /tmp/Dockerfile.overlay -t <image>:<tag> /tmp
kind load docker-image --name starrocks <image>:<tag>
```
Note the deviation in the report (binary-overlay, not a pristine `docker build .`). Verify the
operator pod starts cleanly (leader election, no fatal in logs) to confirm the binary is valid.

**`helm --set` mangles commas and JSON.** A value like `denyList="a,b"` or a JSON `logConfig`
gets split/parsed by `--set`. Use a `-f values.yaml` file (or `--set-string` with `\,`) for any
value containing `,` `{` `}` `[` `]` — otherwise the render silently drops the value and a test
"fails" for the wrong reason.

Run in order:
1. **Static checks** (fast, no cluster): `helm lint`, `helm template`, CRD render. For
   enterprise, `bash scripts/internal/run-checks.sh <FLOOR> <remote>/main` covers conventions +
   helm + go.
2. **Runtime:** `helm uninstall <release> -n <ns>` if present (the existing cluster may be torn
   down) → `helm upgrade --install <release> <chart> -n <ns> -f <values> --set <operator image
   override>` → wait for rollout (`kubectl rollout status` / `kubectl get pods -n <ns> -w`) →
   run each PR's live assertions.

Record every case as PASS/FAIL **with evidence** (the command run + an output snippet).

## Step 4 — Write the report

`doc/internal/test/<YYYY-MM-DD>-<branch>-test-report.md`: summary table (PR → dimension →
result), failures with evidence, and environment info (image tag, chart version, kind cluster,
namespace). Ask the user whether to `helm uninstall` for cleanup. Remove the worktree when
done: `git worktree remove /tmp/test-release-<branch>`.

**Report feature-reachability gaps as bugs.** If a PR lands a feature in the operator / CRD / Go
types but it is **not reachable through the Helm chart** — the chart template never renders the
field, or `values.yaml` exposes no knob — that is a **release bug**, even when the operator side
works in isolation. A Helm user can't use the feature. Call it out in a dedicated "Bugs found"
section with the missing template/values path and a concrete fix. (Real example: #5
`waitForFullRollout` worked via `kubectl patch` but the `celerdatacluster.yaml` template never
rendered it — shipped unreachable via Helm.) Also separate **incidental findings** outside the
tested PR range (e.g. a pre-existing bad default) from release-blocking bugs.

## Quick reference

| Need | Command |
|---|---|
| List release PRs | `git log --oneline <floor>..<remote>/main` |
| See per-PR diffs | `git log --stat --patch <floor>..<remote>/main` |
| Enterprise static checks | `bash scripts/internal/run-checks.sh <floor> <remote>/main` |
| Enterprise helm render | `bash scripts/internal/check-helm.sh` |
| Save base values | `helm get values <release> -n <ns> > base.yaml` |
| Build + load image | `docker build -t <img>:<tag> . && kind load docker-image --name starrocks <img>:<tag>` |
| Watch rollout | `kubectl get pods -n <ns> -w` |

## Common mistakes

- **Trusting a clean Helm render.** A `starrocks*` helper name in the celerdata charts renders
  empty without error. Always cross-check the conventions table for chart/CRD PRs.
- **Wrong diff floor for enterprise.** Do not use the latest `v1.11.x` tag for
  `celerdata-internal-main` — those predate the rebrand. Use `6ba6ae6`.
- **Testing on the wrong branch's chart tree.** `kube-starrocks` only exists on `main`,
  `kube-celerdata` only on `celerdata-internal-main`. Always work in a worktree of the target
  branch.
- **Uninstalling before saving base values.** Run `helm get values` first.
- **Executing before the user reviews the plan.** Step 2 ends with a hard stop.
- **Building data-plane images.** Only the *operator* image is built locally; FE/BE/CN images
  reuse what is already cached on the kind node (`imagePullPolicy: IfNotPresent`), not re-pulled.
- **Blocking on a failed `docker build`.** If docker hub is unreachable, overlay the locally-built
  binary onto a prior operator image instead of giving up (see Step 3).
- **`helm --set` with commas/JSON.** Values containing `,{}[]` get split — use a `-f` values file.
- **Calling a chart-unreachable feature "passing".** Operator-side works ≠ release-ready. If the
  Helm chart can't reach the feature, it's a bug (see Step 4).
- **Expecting `helm upgrade` to refresh CRDs.** Helm does not upgrade CRDs already installed from
  `crds/`; update them manually (`kubectl apply --server-side -f config/crd/bases/...` — plain
  `apply` can hit the 262144-byte annotation limit on large CRDs).
