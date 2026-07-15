# ADR-PC-020: Shared integration environment

- **Status:** **Accepted** — 2026-07-13 (supersedes the 2026-07-10 stub)
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services; owned by `pc-integration-tests` (the Verifier) + `pc-infra` (the stack).
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md) · **[Technology Stack Registry (IS-TECH-REG-001)](../../../shared/docs/conventions/InnoSure-Tech-Stack-Registry.md)** · [Integration Quality Playbook (IS-IQP-001) §5](../../../shared/docs/quality/InnoSure-Integration-Quality-Playbook.md)
- **Related:** ADR-PC-004 (gRPC/REST) · ADR-PC-005 (Kafka/Avro + Apicurio) · ADR-PC-007 (idempotency) · ADR-PC-008 (outbox) · ADR-PC-015 (identity/tenancy/mTLS) · ADR-PC-016/017 (audit/observability) · ADR-PC-019 (contract registry) · ADR-PC-021 (CI/CD) · ADR-PC-023 (Argo CD GitOps)

## Context

Integration, contract, and e2e tiers (IS-IQP-001 §4) need **real** backing services, not
in-process fakes: the bugs integration exists to catch — schema drift, cross-tenant leaks,
cross-wire idempotency replay, mTLS identity rejection, outbox→Kafka delivery — are invisible
to fakes. Per-developer setups diverge; a shared, versioned definition is required.

Two distinct needs, historically conflated: (1) **cheap, isolated, per-PR** proof that a
contract/seam holds; (2) **deployed-artifact** integration of a whole flow against the real
stack with real identities. They demand two environments, not one.

The stack must be the **actual platform stack** (IS-TECH-REG-001), so integration exercises what
production runs — not a simplified stand-in.

## Decision

`pc-integration-tests` defines **two environments** (IS-IQP-001 §5); `pc-infra` provisions the
shared backbone. Both use only technologies locked in IS-TECH-REG-001.

### 1. Ephemeral CI environment (per-PR, throwaway) — IS-IQP-001 §5.2

Each repo spins its dependencies with **`testcontainers-go`**, torn down per run
(`TESTCONTAINERS_RYUK_DISABLED=true` on dev Macs). Gates contract + seam PRs **before** any
deploy. Real, ephemeral: **PostgreSQL 18**, **Kafka 4.x (KRaft)**, **Valkey 8** (Redis-compatible),
**Apicurio Registry 3.x**, **MinIO**.

### 2. Integration server (persistent) — IS-IQP-001 §5.1, the IG3/IG4 system of record

Deployed **OCI artifacts** (digest-pinned, ADR-PC-021) on the shared backbone. Runs on **k3s**
via **Helm**; the Integration Record (IS-IQP-001 §1.4) names each service's image digest. The
backbone, entirely from IS-TECH-REG-001:

| Concern | Technology (IS-TECH-REG-001 §) | Notes |
|:--|:--|:--|
| Transactional state | **PostgreSQL 18** (CloudNativePG) §3/§9 | per-service `pc_{domain}_db` |
| Cache + idempotency | **Valkey 8** §3.1 | drop-in Redis; `go-redis` unchanged |
| Event backbone | **Kafka 4.x KRaft** (Strimzi) §3 | RF3/min-ISR2; clients `twmb/franz-go` §2.1 |
| Schema registry | **Apicurio Registry 3.x** §3 | Avro `BACKWARD`; CCompat API at `/apis/ccompat/v7` |
| Object store | **MinIO** §3 | generated documents (schedule/COI/QR) |
| Identity | **Keycloak 26** §4 | OIDC realm `medhen`; Kafka topic ACLs by service identity |
| **Secrets / PKI / transit** | **OpenBao 2.x** §4 | **direct — NOT HashiCorp Vault, and NOT via External Secrets Operator.** Services read secrets from OpenBao directly; seeds Transit keys for PII envelope (Fayda ID, bank, KYC) |
| Service-identity certs | **cert-manager 1.17** §4 | → **mTLS** (ADR-PC-015), **TLS 1.3** floor 1.2 |
| Observability | **OTel Collector → Mimir / Loki / Tempo / Grafana**; **Prometheus** remote_write §6 | traces/metrics/logs; no PII in logs (ADR-PC-016/017) |
| Edge | **Kong** behind **Traefik** §5 | north-south BFF (`pc-gateway`) |
| OCI registry + CD | **Zot** (+ Trivy + cosign) §9; GitHub Actions; **Argo CD** GitOps (ADR-PC-023) | images build only from signed-off `main`; deploys pin the digest |

- **mTLS by default** (ADR-PC-015 zero-trust) — real service identities, so "unauthorized service
  identity" (IS-IQP-001 §6, X-2) is a testable negative, not a mock.
- **Bootstrap is scripted + idempotent** — seed tenants, OpenBao Transit keys, Kafka topics,
  and Apicurio subjects; a redeploy converges to the same state (`env/bootstrap/`).

### On External Secrets Operator (ESO)

The integration environment uses **OpenBao directly** and **does not run ESO**. ESO's role
(sync OpenBao → k8s `Secret` objects) is a *production* k8s convenience, not required to
integrate flows; excluding it keeps the environment lean and consistent with the OpenBao-direct,
permissive-license posture (IS-TECH-REG-001 §4). If a prod-parity ESO path is later wanted, it is
a separate ADR — not part of the integration environment definition.

## Consequences

- The environment definition (compose/Helm + idempotent bootstrap) lives in
  **`pc-integration-tests/env/`**; the shared stack images/charts come from **`pc-infra`**.
- Per-PR seam/contract tests run on the **ephemeral** env; **flow/saga e2e (IG4)** run only on the
  **integration server** against deployed artifacts.
- This ADR is the **authoritative integration-environment stack** — it reconciles the former stub
  and the IS-IQP-001 §5.1 backbone (closes `pc-integration-tests` spec open decision **IV-2**).
- Any change to this stack (add/drop a component, swap a technology) is an **ADR change**, not an
  `env/` edit — the environment mirrors IS-TECH-REG-001 and moves only with it.

## Alternatives considered

- **Single shared environment only** — rejected: too slow/contended to gate every contract/seam PR.
- **`testcontainers-go` only** — rejected: cannot prove deployed-artifact flows, real mTLS
  identity, or the CD promotion path (IG3/IG4 need a deployed producer).
- **HashiCorp Vault + External Secrets Operator** — rejected: OpenBao is the platform secrets
  choice (LF-governed, MPL-2.0; IS-TECH-REG-001 §4), used directly in this env (see above).
- **Redis (AGPLv3)** — rejected in favour of **Valkey** (BSD-3), consistent with §3.1.
