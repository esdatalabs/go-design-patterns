# Design

## Overview

This application is a sequential pipeline: discover clusters, query each
cluster for metrics, emit per-cluster Prometheus metrics, analyze KPIs
against thresholds, remediate breaches, and notify on remediations applied.

The codebase is organized around **Step Modules** — a pattern where each
step in the pipeline owns everything it needs to run: its execution logic,
the interfaces describing what it needs from the outside world, and the
concrete implementations of those interfaces. Nothing about a step lives
outside its own folder.

This document explains the pattern, why it fits this app, and the
conventions to follow when adding to it.

---

## The Step Modules Pattern

### Core idea

Each step is a **vertical slice** of the pipeline rather than a layer
shared across the whole app. A step's folder contains:

| File | Purpose |
|---|---|
| `step.go` | The step itself — implements `pipeline.Step`, contains orchestration logic |
| `gateway.go` | **Gateway** — an interface describing what this step needs from the outside world |
| `provider_*.go` | **Provider** — a concrete implementation of a Gateway (one file per implementation) |
| `step_test.go` | Unit tests for the step, using hand-written fakes of its Gateway(s) |

**Gateway** and **Provider** are this codebase's terms for what's commonly
called a "port" and "adapter" in hexagonal architecture:

- A **Gateway** is an interface. It says *what* a step needs (e.g., "the
  ability to query a cluster"), with zero knowledge of *how* that's
  fulfilled.
- A **Provider** is a concrete implementation of a Gateway (e.g., a
  Prometheus-backed implementation of `ClusterGateway`). Providers are
  the only place that import external SDKs, HTTP clients, or
  cloud-specific libraries.

A step's logic depends only on its own Gateway interface(s) — never
directly on a Provider, and never on another step's Gateway or Provider.

### Example: the `query` step

```
/internal/steps/query
  step.go                  <- Step, orchestrates querying every cluster
  gateway.go                 <- ClusterGateway interface
  provider_prometheus.go       <- PrometheusProvider implements ClusterGateway
  provider_sql.go                <- SQLProvider implements ClusterGateway
  step_test.go
```

`step.go` only ever calls methods on `ClusterGateway`. It has no idea
whether the concrete implementation talks to Prometheus, a SQL database,
or a test fake — that decision is made once, in `main.go`, at wiring time.

### Full repo layout

```
/cmd
  /clustermon
    main.go                  <- wires Gateways to Providers, builds the pipeline

/internal
  /steps
    /discover
      step.go
      gateway.go
      provider_k8s.go
      step_test.go
    /query
      step.go
      gateway.go
      provider_prometheus.go
      provider_sql.go
      step_test.go
    /emitmetrics
      step.go
      gateway.go
      provider_prometheus.go
      step_test.go
    /analyzekpi
      step.go                 <- pure logic, no Gateway needed
      step_test.go
    /remediate
      step.go
      gateway.go
      provider_restart.go
      provider_scale.go
      provider_failover.go
      step_test.go
    /notify
      step.go
      gateway.go
      provider_slack.go
      provider_pagerduty.go
      step_test.go

  /pipeline
    pipeline.go                <- Step interface + runner
    context.go                  <- PipelineContext, shared across all steps

  /cluster
    types.go                     <- Cluster (shared domain type)
  /kpi
    types.go                      <- KPI, Threshold, KPIResult (shared domain type)
  /remediation
    types.go                       <- Remediation, Record (shared domain type)
  /config
    config.go

/config
  config.yaml
```

Domain types used by *multiple* steps (`cluster.Cluster`, `kpi.KPIResult`,
`remediation.Record`) live outside `/steps`, in their own small packages.
Everything else — anything specific to a single step — stays local to
that step's folder.

---

## Why This Pattern

The app's growth pattern is narrow and predictable: new steps get added
to the pipeline, and more clusters get added to the config. It doesn't
grow new bounded domains, new architectural layers, or multiple parallel
pipelines. Step Modules is deliberately the *minimum* structure that
supports that growth cleanly:

- **Adding a step** = new folder under `/steps`, one line added to the
  pipeline construction in `main.go`. No existing files change.
- **Adding clusters** = a config change. No code changes, since Gateways
  are indexed by cluster name at wiring time.
- **Swapping an implementation** (e.g., replacing Prometheus with
  OpenTelemetry for cluster queries) is a self-contained change: add a
  new `provider_*.go` file in that step's folder, update the wiring in
  `main.go`. Nothing else in the codebase is aware the change happened.
- **No cross-step coupling.** A step can only see its own Gateway(s) and
  the shared `PipelineContext` — never another step's internals — so
  steps can be reordered, removed, or reasoned about independently.

This is intentionally lighter-weight than a full hexagonal or DDD layout
(no app-wide `ports/`, `adapters/`, or `domain/` split). Those patterns
pay off when a codebase has multiple bounded contexts or complex,
evolving business rules. This app has neither — it's a fixed linear flow
— so Step Modules gives the same testability and swappability benefits
without the extra folders and indirection that a bigger, more
domain-heavy system would need.

---

## Conventions

**Naming.** Within a step's package, the step type is just `Step`
(e.g., `query.Step`, `notify.Step`) — the package name already
disambiguates it, so `query.NewStep(...)` at the call site is enough.
Gateways are named `<Thing>Gateway` (`ClusterGateway`,
`NotificationGateway`). Providers are named `<Implementation>Provider`
(`PrometheusProvider`, `SlackProvider`).

**Dependency direction.** `step.go` depends only on `gateway.go` in the
same package. `provider_*.go` files implement the Gateway interface and
are the only files allowed to import external SDKs/clients. Nothing
outside `/cmd/clustermon/main.go` should construct a Provider directly —
wiring Gateways to Providers happens in exactly one place.

**Shared vs. step-local types.** If a type is used by more than one
step (e.g., `Cluster`, `KPIResult`), it belongs in its own top-level
package (`/internal/cluster`, `/internal/kpi`), not inside a step's
folder. If a type only matters to one step's Gateway/Provider contract
(e.g., a raw `QueryResult` shape specific to how `query` talks to its
Gateway), keep it local to that step.

**Error handling.** Steps that operate per-cluster (e.g., `query`,
`remediate`) should not fail the whole step because one cluster errored.
Per-cluster failures are appended to `PipelineContext.Errors` and
logged; the step continues processing the remaining clusters.

**Testing.** Each step's tests use hand-written fakes of its own
Gateway interface(s) — no mocking framework, no live infrastructure.
Pure-logic steps (e.g., `analyzekpi`, remediation selection) should be
table-driven tests with no fakes at all. Full end-to-end pipeline tests
are kept separate, behind an `integration` build tag, and are minimal —
their job is to catch wiring issues, not re-verify logic already covered
by unit tests.

**Adding a new step — checklist:**

1. Create `/internal/steps/<name>/`.
2. Add `gateway.go` if the step needs anything from the outside world;
   skip it if the step is pure logic.
3. Add `provider_*.go` implementing the Gateway, if any.
4. Implement `Step` in `step.go` (`Name() string`, `Run(ctx, *pipeline.PipelineContext) error`).
5. Add `step_test.go` with fakes for the Gateway and table-driven tests
   for any decision logic.
6. Wire it into the pipeline in `main.go`.