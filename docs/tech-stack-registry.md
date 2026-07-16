# InnoGuard — Technology Stack Registry

**Document ID:** IG-TECH-REG-001 · **Version:** 1.1 · **Date:** 2026-06-05 · **Status:** Current
**Purpose:** Single source of truth for **every technology** across Platform Core, FIP, and
AML — what is **in code today** vs **standardized/planned** — so the dev-env (registry, CI,
base images, secrets, gateway) is built to serve the *actual* stack.

**v1.0 → v1.1 delta (this revision):** PostgreSQL 16→**18**; lakehouse Delta Lake+MinIO →
**Apache Iceberg + Apache Ozone**; schema registry → **Apicurio** (not Confluent); secrets
**OpenBao** (final, not HashiCorp Vault); **Grafana Mimir** added to the OTel observability
stack; Kafka client **recommendation = franz-go** (§2.1); base-image drift **resolved** to
distroless; all versions set to latest-stable / enterprise targets (§0). These supersede the
older Architecture Pack values where they differ — **registry-only, no spec/code edits**.

---

## 0. Version & currency policy (financial-crime-grade)

- **Target = latest *stable* GA**, never `latest`/edge. Pin **exact patch** versions at
  deploy; no floating tags. Versions below are the current stable line as of 2026-06 — treat
  them as the *floor to track*, not a frozen set.
- **Prefer LTS** where a project ships one (Flink LTS, K8s patch lines, Go stable).
- **Automate currency:** Renovate/Dependabot for Go modules + base images; Trivy gate in CI
  ([cicd-github-actions.md](./cicd-github-actions.md)) blocks known-CVE versions.
- **Licensing matters for a regulated platform:** favor permissive/foundation-governed OSS
  (Apache-2.0, LF) over source-available re-licenses — this is why **OpenBao** (← Vault) and
  the **Valkey** option (← Redis) are called out. Confirm license posture before adopting.

**Sources of truth**
- Code: `Platform Core/dev/*/go.mod` + Dockerfiles (5 implemented services)
- Standards: [Platform Core Architecture Pack §3, §6–§11](../../Platform%20Core/docs/architecture/Platform-Core-Architecture-Pack.md), Platform/AML/FIP ADRs
- Catalog: [InnoGuard Services Registry](../docs/InnoGuard-Services-Registry.md) (31 services, 5 built)
- Dev-env decisions: [enterprise-dev-env.md](./enterprise-dev-env.md) §1

### Status legend

| Tag | Meaning |
| --- | ------- |
| ✅ **USED** | Present in committed code (go.mod / Dockerfile) |
| 📐 **STANDARD** | Locked in the Architecture Pack / an ADR — authoritative, to be used |
| 🔧 **DEVOPS** | Decided in this dev-env initiative (not app code) |
| 🧪 **CANDIDATE** | Referenced/evaluated, not yet locked |
| ⛔ **DEFERRED** | Explicitly out of MVP scope |

> Only **5 of 31** services have code today (`svc-17` compliance-audit, `svc-20` telemetry-sdk,
> `svc-21` observability-config, `svc-24` idempotency-sdk, `svc-06` rules-mgmt). Everything
> tagged 📐/🧪 is "to be used" — captured here so the platform is provisioned ahead of need.

---

## 1. Languages & runtimes

| Tech | Version | Status | Where / notes |
| ---- | ------- | ------ | ------------- |
| **Go** | **1.25.x** (pin); 1.26 once GA (Feb 2026) | ✅📐 | Primary service language (ADR style §3). Standardize all services + Dockerfiles on one line; retire `golang:1.21-alpine` |
| **Python** | **3.12.x** (LTS-ish, stable) | 📐🧪 | ML serving only (ONNX / TF SavedModel / scikit pickle); separate stack from Go |
| Java / Rust | — | ⛔🧪 | "Future multi-language" note only; no code |
| **SQL (PL/pgSQL)** | **PostgreSQL 18** | 📐 | Recursive CTEs for UBO graph MVP (ADR-AML-005) |

---

## 2. Go application libraries (✅ in code today)

| Concern | Library | Version(s) seen | Notes |
| ------- | ------- | --------------- | ----- |
| HTTP router | `go-chi/chi/v5` | 5.2.1–5.2.5 | REST/OpenAPI external surface |
| Postgres driver | `jackc/pgx/v5` | 5.7.2–5.9.1 | + `lib/pq` in observability-config (drift) |
| Migrations | `golang-migrate/migrate/v4` | 4.17–4.19 | Run via Helm Job |
| Redis client | `redis/go-redis/v9` | 9.7.0–9.18.0 | Cache + idempotency |
| **Kafka client** | `twmb/franz-go` **/** `IBM/sarama` **/** `confluentinc/confluent-kafka-go/v2` | 1.21 / 1.42 / 2.14 | ⚠️ three clients today → **standardize on `franz-go`** (§2.1) |
| Avro codec | `linkedin/goavro/v2` | 2.15.0 | Schema Registry payloads |
| gRPC | `google.golang.org/grpc` | 1.67–1.81 | Internal service-to-service |
| Protobuf | `google.golang.org/protobuf` | 1.36 | Contracts |
| Telemetry | `go.opentelemetry.io/otel` (+sdk, contrib, otlp, prometheus exporters), `otelslog` bridge | 1.41–1.44 | Feeds [pc-telemetry-sdk](../../Platform%20Core/dev/pc-telemetry-sdk) |
| Metrics client | `prometheus/client_golang` | 1.20–1.23 | |
| Logging | `rs/zerolog` (services) **/** `slog`+otelslog (telemetry SDK) | 1.33–1.35 | ⚠️ drift vs slog-only SDK decision — see §14 |
| Config | `spf13/viper` | 1.19 | rules-mgmt |
| Validation | `go-playground/validator/v10` | 10.x | |
| JWT | `golang-jwt/jwt/v5` | 5.2.1 | Token validation (pre-Keycloak) |
| Circuit breaker | `sony/gobreaker` | 0.5.0 & 1.0.0 | Resilience (§9 arch) |
| LRU cache | `hashicorp/golang-lru/v2` | 2.0.7 | |
| Cron | `robfig/cron/v3` | 3.0.1 | Scheduled jobs (audit retention) |
| Object storage SDK | `aws-sdk-go-v2` (config, credentials, **s3**) | s3 1.74 | Pack/object storage (rules-mgmt) — S3-compatible → **Apache Ozone S3 gateway** |
| UUID / YAML | `google/uuid` 1.6, `gopkg.in/yaml.v3` | | |

---

## 2.1 Kafka client — comparison & recommendation

Three Go clients are in use today; this is the single biggest standardization win. Context
that drives the choice: **Kafka 4.x (KRaft)**, **Apicurio Registry** (not Confluent),
**`CGO_ENABLED=0` distroless builds**, exactly-once/transactional needs for financial flows.

| Client | Impl | CGO? | Kafka 4.x / KRaft | Txns / EOS | Schema Registry fit | Maintenance | Verdict |
| ------ | ---- | ---- | ----------------- | ---------- | ------------------- | ----------- | ------- |
| **`twmb/franz-go`** | Pure Go | **No** | Full, incl. newest KIPs (KIP-848 group proto) | First-class, idempotent + txns | `pkg/sr` schema-registry client speaks the **Confluent-compatible API that Apicurio exposes (CCompat)** | Active, single strong maintainer | ✅ **Recommended** |
| `confluentinc/confluent-kafka-go/v2` | CGO (librdkafka) | **Yes** | Full | Yes | Confluent SR libs; works w/ Apicurio CCompat | Confluent-backed, well-funded | ❌ CGO breaks `CGO_ENABLED=0` static/distroless builds (§11) |
| `IBM/sarama` | Pure Go | No | Lags on newest protocol features | Txns supported, rougher ergonomics | No built-in SR; bolt-on | Community/IBM, historically uneven | ❌ Legacy; weakest for 4.x feature parity |
| `segmentio/kafka-go` | Pure Go | No | Good basics | Limited txn support | None built-in | Active but simpler scope | 🧪 Fine for simple producers, not the platform default |

**Recommendation: standardize on [`twmb/franz-go`](https://github.com/twmb/franz-go).**

- **No CGO** → fits the `CGO_ENABLED=0` distroless build standard (§11); confluent-kafka-go
  does not, and that conflict is already latent in the repo.
- **Best Kafka 4.x / KRaft coverage** and the strongest **transactional / exactly-once**
  support — required for at-least-once + idempotency (ADR-PC-007) on financial event paths.
- **Apicurio works via its sibling `pkg/sr`**, which targets the **Confluent-compatible
  (CCompat) REST API** Apicurio Registry serves — so you keep Avro + `BACKWARD` compat
  without pulling in Confluent libraries. Point `pkg/sr` at the Apicurio CCompat endpoint
  (`/apis/ccompat/v7`).
- Already the client in `svc-17`, so partial precedent exists — migrate `svc-21`
  (confluent) and `svc-24` (sarama) onto it.

> **Industry read:** for *new* pure-Go services on modern Kafka, franz-go is now the common
> enterprise default; confluent-kafka-go remains the pick only when teams want the exact
> librdkafka feature set and accept CGO. Given your distroless + Apicurio constraints,
> franz-go is the cleaner fit.

---

## 3. Data, storage & messaging

| Tech | Version | Status | Role / notes |
| ---- | ------- | ------ | ------------ |
| **PostgreSQL** | **18.x** | ✅📐 | Transactional state. Pin CNPG image to **18**; supersedes the older "16" standard |
| **Redis / Valkey** | **8.x** | ✅📐 | Online cache + idempotency store; OT-Redis operator in dev. **Recommended standard: Valkey** — see §3.1 |
| **Apache Kafka** | **4.x (KRaft)** | ✅📐 | Event backbone, RF 3 / min-ISR 2. Pin dev Strimzi to a 4.x line matching prod |
| **Apicurio Registry** | **3.x** | 📐 | Schema registry (**replaces Confluent SR**); Avro `BACKWARD` compat enforced in CI; serves Confluent-compatible **CCompat** API for franz-go `pkg/sr` (§2.1) |
| **Apache Iceberg** | **1.8.x** (table format v2) | 📐 | Open lakehouse table format (**replaces Delta Lake**); Bronze→Silver→Gold medallion (ADR-PC-006 — pack still says Delta; drift §13) |
| **Apache Ozone** | **1.4.x** | 📐 | Lakehouse object storage (**replaces MinIO**); S3-compatible gateway for Iceberg data/warm-cold tiers |
| **Apache Spark** | **3.5.x** (4.0 available) | 📐 | Batch compute writing Iceberg |
| **Apache Flink** | **1.20 LTS** (2.0 available) | 📐 | Stream-to-Iceberg processing |
| Iceberg catalog — REST / Nessie / Polaris | latest | 🧪 | Replaces Hive Metastore/Unity Catalog; choice not locked (lean REST catalog or Polaris) |
| **Debezium** | **3.x** | 📐 | CDC into Kafka |

---

## 3.1 Redis vs Valkey — clarification & recommendation

**The short version:** they are the *same engine* technically; the real difference is
**licensing and governance**, with a secondary edge in momentum/cost.

| | **Redis** | **Valkey** |
| --- | --------- | ---------- |
| Origin | Redis Inc. (commercial steward) | **Linux Foundation** fork of Redis 7.2.4 (Mar 2024), after Redis re-licensed |
| License | **AGPLv3** since Redis 8 (was source-available SSPL/RSAL in 7.4) | **BSD-3-Clause** (permissive — same as old Redis) |
| Backers | Redis Inc. | **AWS, Google, Oracle, Ericsson, Snap** + community |
| Compatibility | — | Wire/RDB/AOF/command **drop-in** with Redis 7.2; `go-redis` works unchanged |
| Perf (8.x) | Multi-threading, new query/vector features | Multi-threaded I/O, memory efficiency; tracks core feature set |
| Managed cloud | Redis Cloud | **AWS ElastiCache/MemoryDB default + ~20–33% cheaper**; GCP Memorystore |
| Operators | OT-Redis (Opstree), Redis Enterprise | Bitnami chart, valkey-operator, ElastiCache |

**Why the licensing reason is *softer* than in 2024, but still real:** Redis 8 (May 2025)
re-added an OSI-approved option (**AGPLv3**), so "Redis went proprietary" is no longer
strictly true. But **AGPLv3 is network-copyleft** — many financial-services legal teams
prefer to avoid it even for an internally-run datastore, whereas **Valkey's BSD-3 is
permissive** with no such questions.

**Recommendation: standardize on Valkey 8.x** for the platform, because it:

1. **Matches your OSS-governance stance** — you already chose **OpenBao over Vault** for the
   same licensing reason; Valkey-over-Redis is the consistent call. [[devenv-stack-decisions]]
2. Is a **true drop-in** — your use is cache + idempotency store (no Redis-8-exclusive
   feature), `go-redis/v9` is unchanged, RDB/AOF migrate cleanly.
3. Is **cheaper and default on AWS managed** — relevant with prod approaching.
4. Has the broader **vendor + community backing** and release cadence.

**Operational catch (the one real cost):** the **OT-Redis (Opstree) operator is
Redis-specific.** Moving to Valkey means swapping the dev operator for a Valkey path
(Bitnami Valkey chart / a Valkey operator / ElastiCache-for-Valkey in prod). It is **not
worth churning dev today** — keep OT-Redis (Redis 7.2-compatible) for now and adopt Valkey
when you provision the prod data tier. Clients and app code don't change either way.

> **Net:** Valkey = platform standard (license + cost + momentum); Redis-on-OT-operator =
> fine for dev until the prod cutover. Confirm and I'll lock it in §3 and the YAML.

---

## 4. Security & identity

| Tech | Status | Role / notes |
| ---- | ------ | ------------ |
| **Keycloak** 26.x | 📐 | OIDC/OAuth2 IAM provider (`svc-19`); also gates Kafka topic ACLs by service identity |
| **OpenBao** 2.x | 🔧📐 | Secrets, PKI, transit encryption — **final choice, NOT HashiCorp Vault** ([secrets-openbao.md](./secrets-openbao.md)). LF-governed, MPL-2.0. Arch Pack still says "Vault" → drift §13 |
| **cert-manager** 1.17.x | 🔧📐 | Service-identity certs → mTLS; Let's Encrypt for edge ([ingress-tls-hardening.md](./ingress-tls-hardening.md)) |
| **External Secrets Operator** 0.x (latest) | 🔧 | Syncs OpenBao → k8s Secrets |
| mTLS / Zero Trust | 📐 | All internal S2S (ADR-PC-015); **TLS 1.3** (1.2 floor) |
| RBAC + ABAC | 📐 | Roles + attribute-based fine-grained authz |
| Encryption | 📐 | AES-256 at rest (**OpenBao transit** envelope for PII); per-product retention classes |

---

## 5. API, networking & gateway

| Tech | Status | Role / notes |
| ---- | ------ | ------------ |
| **gRPC** (HTTP/2 + protobuf) | ✅📐 | Internal service-to-service (ADR-PC-004) |
| **REST / OpenAPI 3.1** | ✅📐 | Customer-facing product APIs (FIP & AML each expose their own) |
| **Kong** (API gateway) | 🔧 | North-south edge: per-customer API keys, rate-limit→HTTP 429, channel-header extraction, tracing — matches the Arch Pack "API Gateway" layer ([api-gateway-kong.md](./api-gateway-kong.md)) |
| **Traefik** (edge ingress) | 🔧 | TLS edge in front of Kong; dashboards/infra bypass Kong ([ingress-tls-hardening.md](./ingress-tls-hardening.md)) |
| **Istio / Linkerd** (service mesh) | 📐🧪 | East-west mTLS, traffic shaping, retry/timeout (ADR-PC-015). Complements Kong (north-south); choice not locked |

---

## 6. Observability

| Tech | Status | Role |
| ---- | ------ | ---- |
| **OpenTelemetry SDK** | ✅📐 | Metrics + traces + logs instrumentation; in code via [pc-telemetry-sdk](../../Platform%20Core/dev/pc-telemetry-sdk) |
| **OTel Collector** 0.116.x | 📐 | Central OTLP pipeline: receive → process (batch, tenant attrs) → export to Mimir/Loki/Tempo. The seam between apps and backends |
| **Prometheus** 3.x | ✅📐 | Scrape/agent + recording & SLO burn-rate rules; **`remote_write` to Mimir** for storage (short local retention only) |
| **Grafana Mimir** 2.15.x | 📐 | **Long-term, HA, horizontally-scalable, multi-tenant metrics store** (Prometheus-compatible). See note below — this is the metrics system of record |
| **Loki** 3.x | 📐 | Structured-JSON log store (trace_id/span_id/service/version mandatory) |
| **Tempo** 2.x | 📐 | Trace store (Jaeger/OTLP-compatible) |
| **Grafana** 11.x | 📐🔧 | Dashboards + unified alerting over Mimir/Loki/Tempo; seeded by `svc-21`; **edge-routed via Traefik**, bypasses Kong |

> [!NOTE]
> **What Mimir does in this stack (Q6).** Prometheus alone is a single-node, short-retention
> scraper — fine for live golden signals, weak for the platform's needs. **Mimir is the
> durable, scalable metrics backend behind it:**
> - **Long retention** — keeps metrics for months/years to satisfy financial-crime audit and
>   capacity/trend analysis, which Prometheus' local TSDB isn't built for.
> - **High availability + scale** — clustered, object-storage-backed (sits naturally on your
>   **Apache Ozone / S3** tier), survives node loss, and scales ingestion horizontally.
> - **Multi-tenancy** — native per-tenant isolation maps cleanly to **per-product namespaces
>   (FIP / AML / platform)**, with separate limits and access.
> - **Drop-in** — speaks PromQL and the Prometheus remote-write protocol, so Prometheus (or
>   the OTel Collector) `remote_write`s into it and Grafana queries it unchanged.
>
> Flow: **app → OTel SDK → OTel Collector → Mimir (metrics) / Loki (logs) / Tempo (traces) →
> Grafana**. Prometheus becomes a thin scraping/rules agent in front of Mimir, not the store.

---

## 7. ML / AI & analytics (📐🧪 — not in code)

| Tech | Status | Role |
| ---- | ------ | ---- |
| Python ML serving | 📐 | Separate from Go hot path; champion/challenger + fallback (ADR-PC-011) |
| ONNX / TF SavedModel / scikit-learn | 📐 | Standard model interchange formats |
| GenAI (async-only) | 📐 | STR narrative drafting etc. (ADR-PC-012, ADR-AML-007); async, never hot-path |
| Vector DB — pgvector / Weaviate / Milvus | 🧪 | Low signal; candidates only, none locked |
| ClickHouse | 🧪 | Single mention; analytics candidate |

---

## 8. Graph & search

| Tech | Status | Role |
| ---- | ------ | ---- |
| PostgreSQL recursive CTEs | 📐 | **UBO ownership graph MVP** (`aml-ubo-registry`, ADR-AML-005) |
| Neo4j / JanusGraph | ⛔ | Dedicated graph DB **explicitly deferred** past MVP ("not in platform standards"); Platform Graph Phase 3 |
| Elasticsearch / OpenSearch | 🧪 | Search candidate (23 refs); not locked |

---

## 9. Platform, orchestration & DevOps (🔧 dev-env initiative)

| Tech | Status | Role / notes |
| ---- | ------ | ------------ |
| **Kubernetes** | 📐 | **1.32.x** in prod (track patch line; 1.30 floor per ADR style §3) |
| **k3s** | 🔧 | Single-node Milan **dev** cluster ([server_setup_guide.md](./server_setup_guide.md)) |
| **Helm** | ✅🔧 | Unit of deploy; per-env values files |
| **CloudNativePG** | 🔧 | Postgres operator (dev data stack) |
| **OT-Redis (Opstree)** | 🔧 | Redis operator |
| **Strimzi** | 🔧 | Kafka operator (KRaft) |
| **Zot** | 🔧 | Self-hosted OCI registry + Trivy scan + cosign ([artifact-registry-zot.md](./artifact-registry-zot.md)) |
| **GitHub Actions** | 🔧 | CI/CD ([cicd-github-actions.md](./cicd-github-actions.md)) |
| **Trivy** | 🔧 | Image CVE scanning (CI gate + Zot) |
| **cosign / Sigstore** | 🔧 | Image signing (key-based) |
| **Kyverno** | 🔧🧪 | Optional admission: enforce signed images |
| **Vultr** (Object Storage, compute) | 🔧 | Host + S3-compatible backups ([backups-dr.md](./backups-dr.md)) |

---

## 10. Testing & quality (✅ in code)

| Tool | Purpose |
| ---- | ------- |
| `stretchr/testify` | Assertions / unit tests |
| `testcontainers-go` (+ kafka, redis, postgres, localstack modules) | Integration tests against real deps |
| `cucumber/godog` | BDD / acceptance specs |
| `leanovate/gopter` + `pgregory.net/rapid` | Property-based testing |
| `alicebob/miniredis` | In-memory Redis fake | 
| LocalStack (via testcontainers) | AWS/S3 emulation for rules-mgmt |

> Aligns with the staged test initiative (rapid + testcontainers locked).

---

## 11. Container & build baseline

**Standard (resolves the bookworm-slim vs alpine drift).** Ready-to-use template:
[`templates/Dockerfile`](./templates/Dockerfile) + [`.dockerignore`](./templates/.dockerignore)
([usage](./templates/README.md)).

| Item | **Standard** | Status | Notes |
| ---- | ------------ | ------ | ----- |
| Build stage | **`golang:1.25-bookworm`** (pin digest) | 📐 | Single Go line for all services; glibc matches the runtime below. Retire `golang:1.21-alpine` |
| Runtime stage | **`gcr.io/distroless/static-debian12:nonroot`** (pin digest) | 📐 | One standard for everyone. No shell/pkg-mgr → smallest CVE surface; runs **non-root** by default; ideal for static Go binaries |
| Build flags | `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`, multi-stage | ✅ | Keep. Static build is **required** by distroless-static and franz-go (§2.1) — and is why confluent-kafka-go (CGO) is rejected |
| Ports | 8080 (HTTP), 50053 (gRPC) | ✅ | |

**Why distroless over alpine:** alpine uses **musl** libc (subtle DNS/threading differences
vs glibc, occasional cgo/runtime surprises) and still ships a shell/apk — a larger attack
surface. **distroless `static`** has no libc dependency for `CGO_ENABLED=0` Go binaries, no
shell, and a non-root user, which is the industry default for hardened Go service images.
Avoid `debian:bookworm-slim` as a *runtime* base (full userland you don't need); keep
bookworm only as the *build* stage.

---

## 12. Per-service stack matrix (implemented services)

| Service | Lang | DB | Cache | Kafka client | Notable |
| ------- | ---- | -- | ----- | ------------ | ------- |
| `svc-17` compliance-audit | Go 1.25 | pgx | go-redis + miniredis | **franz-go** | chi, jwt, goavro, cron, gobreaker, gopter |
| `svc-20` pc-telemetry-sdk | Go 1.26 | — | — | — | OTel SDK (slog, OTLP exporters) — shared lib |
| `svc-21` observability-config | Go 1.25 | pgx + lib/pq | go-redis | **confluent-kafka-go** | golang-migrate, zerolog, OTLP |
| `svc-24` idempotency-sdk | Go 1.25 | — | go-redis | **IBM/sarama** | godog, gobreaker, goavro — shared lib |
| `svc-06` rules-mgmt | Go 1.26 | pgx | go-redis | franz/kafka (testcontainers) | viper, **aws-sdk-go-v2/s3**, rapid, localstack |

> The `fip-observability-svc` and `fip-idempotency-management-sdk` directories duplicate
> `pc-*` counterparts and several declare the wrong module path — see §14.

---

## 13. Reconciliation backlog (drift & risks a senior review should close)

| # | Finding | Impact | Recommendation |
| - | ------- | ------ | -------------- |
| 1 | **3 Kafka clients** (franz-go, IBM/sarama, confluent-kafka-go) | Inconsistent semantics, ops burden | **Decided: `franz-go`** (§2.1). Migrate `svc-21`, `svc-24` |
| 2 | **confluent-kafka-go needs CGO**, but Dockerfiles build `CGO_ENABLED=0` | Build breaks / fat images | Drop confluent-kafka-go (ties to #1) |
| 3 | Go version drift: 1.21 / 1.25 / 1.26 | CI matrix, base-image skew | Pin **1.25.x** everywhere now; move to 1.26 after GA. Retire `1.21-alpine` |
| 4 | Base-image drift: bookworm-slim vs alpine | Inconsistent CVE surface | **Resolved (§11): build `golang:1.25-bookworm` → runtime `distroless/static-debian12:nonroot`** |
| 5 | Logging drift: zerolog vs slog | SDK decision is **slog-only** | Migrate services to slog via telemetry SDK |
| 6 | Postgres version | Behavioral surprises | **Target PG 18** — pin CNPG image to 18 in dev + prod |
| 7 | Kafka 4.x (standard) vs older dev Strimzi | Feature/metadata mismatch | Pin dev Strimzi to the same 4.x line as prod |
| 8 | Arch Pack says **Vault**; final is **OpenBao** | Doc/impl mismatch | Update ADR-PC / Arch Pack §11 to OpenBao |
| 9 | Arch Pack says **Delta Lake + MinIO**; platform uses **Iceberg + Ozone** | Doc/impl mismatch | Update ADR-PC-006 / Arch Pack §6 to Iceberg + Ozone |
| 10 | Arch Pack says **Confluent Schema Registry**; platform uses **Apicurio** | Doc/impl mismatch | Update Arch Pack §7 to Apicurio (CCompat API) |
| 11 | go.mod **module-path copy-paste errors** (≥4 modules named `github.com/fip/observability-config-service`) | Import collisions, broken `go get` | Rename module paths per service |

---

## 14. Machine-readable summary

```yaml
languages:   { primary: go, go_version: 1.25.x, secondary: [python-3.12], deferred: [java, rust] }
data:        { rdbms: postgresql-18, cache: redis-8 (valkey-8 option),
               lakehouse_table_format: apache-iceberg-1.8, object_store: apache-ozone-1.4 }
messaging:   { broker: kafka-4.x-kraft, client: twmb-franz-go, schema: apicurio-registry-3.x (ccompat),
               format: [avro, protobuf], cdc: debezium-3.x }
compute:     { batch: spark-3.5, stream: flink-1.20-lts, catalog: [rest, nessie, polaris] (tbd) }
api:         { internal: grpc, external: rest-openapi-3.1, gateway: kong, mesh: [istio, linkerd] }
security:    { iam: keycloak-26, secrets: openbao-2 (not-vault), certs: cert-manager-1.17, model: zero-trust-mtls-1.3 }
observability: { sdk: opentelemetry, pipeline: otel-collector, metrics_store: grafana-mimir,
               scraper: prometheus-3 (remote_write), logs: loki-3, traces: tempo-2, ui: grafana-11 }
ml:          { serving: python, formats: [onnx, tf-savedmodel, sklearn], genai: async-only }
graph:       { mvp: postgresql-cte, deferred: [neo4j, janusgraph] }
build:       { build_image: golang-1.25-bookworm, runtime_image: distroless-static-debian12-nonroot, cgo: disabled }
devops:      { orchestrator_prod: kubernetes-1.32, orchestrator_dev: k3s, registry: zot, ci: github-actions,
               scan: trivy, sign: cosign, secrets_sync: external-secrets, package: helm }
status_counts: { services_total: 31, implemented: 5 }
```

---

*Keep this registry in lockstep with the [Services Registry](../docs/InnoGuard-Services-Registry.md)
and the [Architecture Pack](../../Platform%20Core/docs/architecture/Platform-Core-Architecture-Pack.md).
When a 📐/🧪 item ships into code, flip it to ✅ and record the version.*
