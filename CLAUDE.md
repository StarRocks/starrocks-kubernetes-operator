# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

The **CelerData** Kubernetes Operator — an internal fork of the StarRocks operator (MPP OLAP database). It reconciles two CRDs into FE / BE / CN / FE-proxy workloads via sub-controllers:

- `CelerDataCluster` (short name `cdc`) — full cluster: FE (required) + BE and/or CN, optionally an FE-proxy.
- `CelerDataWarehouse` (short name `cdw`) — an extra compute warehouse (CN only), attached to an existing cluster.

> The Go module is `github.com/CelerData/celerdata-kubernetes-operator-internal`. Many file and chart names still use the historical `starrocks` prefix (e.g. `starrockscluster_controller.go`, the `starrocks.com/...` annotation keys) even though the CRD types are CelerData-branded.

## Build, test, and lint

```bash
# Unit tests (fast, no cluster) — use this in worktrees and day-to-day
go test ./pkg/... -timeout 120s

# Quick compile check
go build ./cmd/...

# Static analysis
go vet ./...

# Lint (matches CI; golangci-lint v2.2.2, config in .golangci.yml)
golangci-lint run --timeout=30m ./...

# Run a single package's tests
go test ./pkg/subcontrollers/fe/... -run TestName -v
```

`go test ./pkg/...` is fine for a fast inner loop. **Before opening a PR, run the full checks the
[PR template](.github/pull_request_template.md) requires** — for operator changes: `make generate`
(deepcopy), `make manifests` (CRD yaml under `config/crd/bases` and `deploy/`), `make test` (envtest
UT; downloads kubebuilder assets on first run), and `golangci-lint run`; for helm-chart changes:
update the relevant subchart `values.yaml` and run `bash scripts/create-parent-chart-values.sh` to
regenerate the parent `kube-celerdata/values.yaml`. `make generate`/`make manifests` rewrite
generated files — **commit the regenerated output**, since CI (`action-make-test.yml`) fails on
drift. `scripts/internal/run-checks.sh <base-ref> <head-ref>` runs this whole suite locally and is
what the sync / port skills invoke before opening a PR.

## Architecture

The operator has a **single binary** (`cmd/main.go`) running two top-level reconcilers wired in `pkg/controllers/controllers.go`:

- **`CelerDataClusterReconciler`** (`starrockscluster_controller.go`) holds an ordered list of `ClusterSubController`s — `fe`, `be`, `cn`, `feproxy`. Each `Reconcile` walks the sub-controllers in order; FE must come up before BE/CN. Per-sub-controller failures are aggregated via `handleSyncClusterError`.
- **`CelerDataWarehouseReconciler`** (`starrockswarehouse_controller.go`) holds only a `cn` `WarehouseSubController`, since a warehouse is compute-only.

Sub-controllers live in `pkg/subcontrollers/{be,cn,fe,feproxy}`, each constructed via `New(client, eventRecorderFor)`. The shared `subcontroller.go` decides whether a component renders as a Deployment (`DeploymentLoadType`) or StatefulSet (`StatefulSetLoadType`) and delegates to the workload templates.

`pkg/k8sutils/templates/{deployment,statefulset,pod,service,object}` build the actual Kubernetes objects; `pkg/k8sutils/k8sutils.go` holds the apply/get helpers. `pkg/predicates` filters reconcile events — notably an `ignored` annotation deny-list (`starrocks.com/ignored`) caches at init so flagged objects are skipped.

## Module layout

```
pkg/
  apis/celerdata/v1/      CRD types (celerdatacluster_types.go, celerdatawarehouse_types.go,
                          auto_scale.go, component_type.go, load_type.go) — DO NOT hand-edit zz_generated.deepcopy.go
  controllers/            Top-level reconcilers (files keep the starrocks* prefix)
  subcontrollers/{be,cn,fe,feproxy}   One sub-controller per component
  k8sutils/               Apply/get helpers; templates/ builds deployment/statefulset/pod/service objects
  k8sutils/fake           Fake controller-runtime client + manager for unit tests
  predicates/             Reconcile event filtering (deny-list via annotation)
  common/                 Shared utilities
cmd/main.go               Single operator entrypoint (cmd/config/ holds startup config)
helm-charts/charts/       kube-celerdata, warehouse
```

Anything matching `zz_generated*.go` is produced by `controller-gen` / `deepcopy-gen`. Regenerate via `make generate` from the main checkout — never hand-edit.

## Code conventions

- **`goimports` local-prefix** is `github.com/CelerData/celerdata-kubernetes-operator-internal` (set in `.golangci.yml`); local imports group last.
- **Pass `ctx` explicitly** through reconcile/sub-controller call chains. (Note: a few legacy `context.TODO()` / `context.Background()` calls remain in `pkg/`; don't add new ones — thread the incoming `ctx` instead.)
- **Unit tests use the fake client at `pkg/k8sutils/fake`**, not a real cluster.
- **Test files live in the same package as the code under test** (`package foo`, not `package foo_test`).
- **Line length cap:** 140 chars (`lll`). Other notable linters: `funlen` (60 statements), `gocyclo` (complexity 16), `gosec`, `revive`, `dupl`. `logrus` is banned via `depguard` — use the project logger.
- **Go version:** `go.mod` declares 1.22; CI runs tests on 1.22 and lint on 1.24.
- `go.sum` / `vendor/` — this repo vendors dependencies. Adding a new dependency changes both and the lockfile; avoid introducing new deps in a focused task.

## Commit & PR titles

Bracket-tag style, matching the open-source StarRocks project so both repos read the same:
`[Feature]`, `[BugFix]`, `[Enhancement]`, `[Chore]`, `[Documentation]` — e.g.
`[BugFix] Quote feProxy resolver so empty value renders "" not null`. Sync PRs (an upstream PR
pulled into this repo) keep the original upstream subject behind a `sync #N:` prefix, e.g.
`sync #747: [Feature] Add cluster and group metric labels`. Per global instructions, write code,
commit messages, and PR titles in English. See `doc/internal/git_workflow_howto.md` for the full
cross-repo title conventions.
