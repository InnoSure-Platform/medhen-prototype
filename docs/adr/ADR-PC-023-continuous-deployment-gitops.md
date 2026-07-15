# ADR-PC-023: Continuous deployment via GitOps

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-13
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md) · builds on [ADR-PC-021](./ADR-PC-021-cicd-image-build-and-publish.md) (build/publish) and [ADR-PC-020](./ADR-PC-020-integration-environment.md) (integration environment)

## Context

ADR-PC-021 defines how an image is built and published (immutable digest to Zot, no moving tags) but not how it reaches a cluster. Deployment must be reproducible, least-privilege, auditable, and safe to drive from the **agentic framework** — where implementer agents must never hold live cluster credentials or mutate a cluster directly. A push-based pipeline (CI runs `helm upgrade` against the cluster) gives CI cluster access, tracks a mutable tag, and has no review gate — the wrong trust model here.

## Decision

**Continuous deployment is pull-based GitOps, digest-pinned.**

- **CI never deploys and holds no cluster credentials.** The per-service `release.yml` (ADR-PC-021) only builds, scans, signs, and pushes an **immutable digest**.
- **Desired state lives in git** as the golden-path Helm chart (`service-chart`) plus a per-service release manifest that pins the image by **digest** (`repository@sha256:…`), never a tag.
- **A GitOps controller reconciles it** — **Flux** is the platform default (lightweight, CRD/CLI-driven, native image-update automation; fits self-hosted k3s + Zot). Argo CD is the alternative where a multi-tenant UI is needed.
- **A deploy is a git change.** Promoting a new image = a PR that bumps the pinned digest (raised by Flux image-automation, Renovate, or an agent). Merge = deploy; `git revert` = rollback. Agent deploy actions therefore inherit the same review/policy/rollback guardrails as code — an agent proposes a digest-bump PR, it is gated, the cluster pulls; the agent never touches the cluster.
- The `service-chart` is the **single source of deployment manifests** (Deployment, Service, probes, OTLP, ExternalSecret, NetworkPolicy, KafkaTopic, optional Ingress). Scattered per-service k8s YAML and push-based `cd.yaml` workflows are retired.

## Consequences

Requires a GitOps controller in-cluster (Flux) and a desired-state location per service (the in-repo `deploy/` overlay watched centrally, or a central GitOps repo). Enforced by the quality system (IS-AQG-001) and the shared scaffold: the scaffold installs the chart + a digest-pinned release manifest and does **not** install a cluster-credentialed deploy workflow. Consistent with ADR-PC-021 (digest provenance) and ADR-PC-020 (integration environment consumes the pinned digest). This is a stub capturing the platform's stated convention — expand with alternatives considered (push-based CI deploy; Argo CD vs Flux; in-repo vs central GitOps repo) and full rationale before ratifying.
