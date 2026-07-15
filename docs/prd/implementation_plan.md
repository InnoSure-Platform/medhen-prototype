# Medhen Platform — Phase 0 Implementation Plan

**Document ID:** MDH-IMPL-001 · **Version:** 1.0 · **Date:** July 2026 · **Status:** Active  
**Owner:** Lead Solutions Architect  
**Companions:** [PRD §34.1](./Medhen-Platform-PRD.md#341-phase-0--pilot-mvp-detail) · [Capability Document](./Medhen-Platform-Capability-Document.md) · [Service Registry](./service-registry.md) · [ADR Index](../adr/README.md)

---

## 1. Mission

Deliver a **demo-winning, enterprise-grade Phase 0 Motor vertical slice** for the **Ethiopian Insurance Corporation (EIC)** that proves:

1. **End-to-end automation** — quote → STP UW → Telebirr pay → bilingual issuance → FNOL → fast-track settlement  
2. **Ethiopian localization** — Amharic/EN, ETB, VAT 15% + stamp duty, Ethiopian calendar, Telebirr, EIC branding  
3. **Shared-core extensibility** — one seeded Motor product/config teaser; Life/Property = configuration, not fork  

Built on the **production architecture** (multi-service, DDD, hexagonal, EDA, outbox, saga, CQRS-where-justified) with **zero throwaway** — Phase 0 services thicken into Phase 1 Motor production.

**Duration:** ~10–12 weeks · **Team:** 4–6 engineers · **Target grade:** Tier-0 (regulatory/financial/hot-path patterns on every money & policy path)

---

## 2. Workspace topology

This prototype workspace is a **monorepo that mirrors the multi-repo layout** so each Bounded Context remains an independently buildable/deployable module (Go module per service). When promoting to multi-repo, each directory under `services/` and top-level library becomes its own git remote with no structural rewrite.

```
Medhen Prototype/
├── docs/prd/                 # Cap doc, PRD, registry, this plan
├── docs/adr/                 # Architecture decision records
├── contracts/                # pc-contracts (proto, OpenAPI, Avro)
├── shared/go/                # pc-shared-go (Kernel BC-MDH-20)
├── infra/                    # pc-infra (Compose + bootstrap)
├── services/                 # pc-*-svc runtimes
│   ├── gateway/
│   ├── iam-svc/
│   ├── audit-svc/
│   ├── integration-svc/
│   ├── party-mgmt-svc/
│   ├── product-defn-svc/
│   ├── rating-calc-svc/
│   ├── underwriting-svc/
│   ├── policy-svc/
│   ├── billing-svc/
│   ├── document-svc/
│   ├── notification-svc/
│   └── claims-svc/
├── web/                      # pc-web (Next.js)
├── seeds/                    # Demo personas, Motor product, rate table
├── scripts/                  # bootstrap, demo, codegen
├── go.work                   # Go workspace
└── README.md                 # Runbook
```

**Module path root:** `github.com/InnoSure-Platform/`

---

## 3. Technology stack (locked)

Per ADR-PC-020 / IS-TECH-REG-001 principles applied to Phase 0 local/demo:

| Concern | Technology |
|:---|:---|
| Language (services) | Go 1.24+ |
| Sync inter-service | gRPC (`pc-contracts` protos) |
| Edge / BFF | `pc-gateway` REST/JSON (OpenAPI) |
| Frontend | Next.js 15 (App Router) + TypeScript |
| OLTP | PostgreSQL 16+ (one DB schema per service: `pc_{domain}_db`) |
| Cache / idempotency | Valkey 8 (Redis-compatible) |
| Events | Kafka (KRaft) + Avro subjects |
| Schema registry | Apicurio Registry (ccompat) — Phase 0 may embed registry optional |
| Object store | MinIO (policy docs, FNOL photos) |
| Identity | Keycloak 26 · realm `medhen` · tenant `eic` |
| Observability | OpenTelemetry → Jaeger + Prometheus + Grafana |
| Secrets (local) | Compose env / files; OpenBao deferred to integration server |

---

## 4. Demo storyboard (acceptance script)

| Step | Actor | Action | Proof |
|:---:|:---|:---|:---|
| 1 | Agent/Customer | Motor quote (Amharic/EN) | Itemized premium: base + factors + VAT 15% + stamp duty |
| 2 | System | STP underwriting | Auto-accept standard risk |
| 3 | Customer | Telebirr sandbox / mock pay | Receipt number |
| 4 | System | Issue policy | Bilingual schedule + COI + QR sticker PDFs |
| 5 | Customer | Mobile FNOL (photo + GPS) | Fast-track settlement + SMS |
| 6 | Staff | Product-config / risk-schema | Shared-core teaser |
| 7 | Staff | Audit trail + KPI tile | GWP / policies in-force |

---

## 5. In-scope REQ slice (happy path)

| Domain | Pilot | REQ-IDs |
|:---|:---|:---|
| Party | Register individual + KYC upload | `REQ-PTY-001`, `-020` |
| Product | Seeded Motor product | `REQ-PRD-001/010/020` |
| Rating | Motor + VAT/stamp | `REQ-RAT-001/002/003/006` |
| UW | STP auto-accept | `REQ-UW-001/002` |
| Policy | Quote → bind → issue | `REQ-POL-001/002/003/020/021/030` |
| Billing | Telebirr single premium | `REQ-BIL-001/010/020/023` |
| Docs | Schedule + COI + QR sticker | `REQ-DOC-001/002/010/020` |
| Claims | FNOL → fast-track settle | `REQ-CLM-001/002/011/040/050` |
| Notify | Bind + settle SMS/email | `REQ-NOT-001/011` |
| Core | am/en, ETB, calendar, Motor schema | `REQ-CORE-002/006/007` |
| IAM/Audit | Basic roles + immutable trail | `REQ-IAM-002/003`, `REQ-AUD-001/003` |

**Out of Phase 0:** endorsements/renewals/cancellations, installments, ERP, full Fayda, UW referral, reserves/recovery, reinsurance, commission, complaints, fin-crime, NBE returns, multi-branch, deep ABAC.

---

## 6. Milestones

| ID | Weeks | Exit criteria |
|:---|:---|:---|
| **M0 Foundations** | 1–2 | Compose stack healthy; contracts + shared-go published; IAM + gateway + party walking skeleton; empty quote creatable |
| **M1 Buy journey** | 3–6 | Steps 1–4 of storyboard green E2E via gateway + web |
| **M2 Claim + story** | 7–10 | Steps 5–7 green; seed script; demo deploy |
| **M3 Polish** | 11–12 | Hardening, rehearsal runbook, SLO dashboards, buffer |

---

## 7. Build waves (execution order)

### W0 — Platform foundations
`contracts` → `shared/go` → `infra` → `observability (thin)` → `iam-svc` + Keycloak → `audit-svc` → `integration-svc` → `gateway` → `party-mgmt-svc`

### W1 — Contract core
`product-defn-svc` → `rating-calc-svc` → `underwriting-svc` → `policy-svc`

### W2 — Money & servicing
`billing-svc` → `document-svc` → `notification-svc`

### W3 — Claims
`claims-svc`

### Continuous
`web` — screens land against each wave’s OpenAPI

---

## 8. Cross-cutting patterns (Tier-0 bar)

Every mutating money/policy/claims path MUST implement:

| Pattern | ADR | Phase 0 bar |
|:---|:---|:---|
| Hexagonal layout | ADR-PC-002 | `internal/{domain,application,ports,adapters}` |
| Idempotency keys | ADR-PC-007 | Valkey or PG unique on pay/bind/FNOL/settle |
| Transactional outbox | ADR-PC-008 | Same TX as aggregate write |
| Saga / compensation | ADR-PC-009 | Bind ↔ Billing payment saga |
| Audit event emit | ADR-PC-016 | Append-only consumer + query API |
| OTel traces | ADR-PC-017 | Trace-ID end-to-end via gateway |
| Tenant isolation | ADR-PC-015 | `tenant_id=eic` on every row & claim |
| Contract-first | ADR-PC-019 | No hand-copied cross-service schemas |

---

## 9. Service minimal APIs (Phase 0)

| Service | Key operations |
|:---|:---|
| party-mgmt | `RegisterIndividual`, `GetParty`, `UploadKYC` |
| product-defn | `GetProduct`, `ListCoverages`, `GetRiskSchema` |
| rating-calc | `CalculatePremium` |
| underwriting | `EvaluateRisk` (STP accept/reject) |
| policy | `CreateQuote`, `GetQuote`, `BindQuote`, `IssuePolicy`, `GetPolicy` |
| billing | `CreateInvoice`, `InitiatePayment`, `ConfirmPayment`, `GetReceipt` |
| document | `GeneratePack` (schedule/COI/sticker), `GetDocument` |
| notification | `Send` (triggered by events) |
| claims | `SubmitFNOL`, `GetClaim`, `SettleFastTrack` |
| audit | `Append`, `Query` |
| integration | `TelebirrCharge`, `SendSMS`, `VerifyFayda` (mock) |
| iam | Keycloak broker + role claims enrichment |
| gateway | REST façade for all of the above |

---

## 10. Saga: Quote → Bind → Pay → Issue

```
CreateQuote
  → CalculatePremium (sync)
  → EvaluateRisk STP (sync)
  → BindQuote (policy PENDING_PAYMENT)
  → CreateInvoice + InitiatePayment (billing)
  → Telebirr sandbox confirm (integration)
  → ConfirmPayment → PolicyIssued event
  → DocumentGenerate + Notify (async consumers)
On payment failure: compensate → QuoteExpired / BindCancelled
```

---

## 11. Demo environment & seed

- **Local:** `infra/docker-compose.yml` + `scripts/demo-up.sh`  
- **Seed:** `seeds/motor-product.json`, `seeds/demo-users.json`, EIC brand assets in `web/public/`  
- **Script:** 15-minute facilitator runbook in `README.md`  
- **Policy number:** `EIC/MOT/{YYYY}/{seq}`  

---

## 12. Quality gates (Phase 0)

| Gate | Requirement |
|:---|:---|
| Unit | Domain rules (rating, STP, money) ≥ 80% on critical packages |
| Contract | Gateway OpenAPI matches `contracts/openapi` |
| Seam | Party → Quote path against Compose stack |
| E2E | Storyboard 1–7 via `scripts/demo-e2e.sh` |
| Security | No secrets in git; JWT validation at gateway; tenant check on reads |

---

## 13. Risks & mitigations

| Risk | Mitigation |
|:---|:---|
| Repo sprawl before value | Cookiecutter template + shared library first 48h |
| Localization last | i18n keys + calendar in shared-go from day 1 |
| Scope creep | Storyboard is the hard gate |
| Kafka/outbox lag | Outbox poller in shared-go; reference impl = party |

---

## 14. Immediate execution checklist

- [x] Author this plan (`MDH-IMPL-001`)
- [x] Scaffold monorepo + `go.work`
- [x] `contracts` — pilot protos + OpenAPI + topics
- [x] `shared/go` — money, calendar, i18n, outbox, middleware
- [x] `infra` — Compose backbone
- [x] W0 services + gateway walking skeleton (`medhen-api` composition host)
- [x] W1–W3 services + web + seeds + demo script (happy-path)

---

## 15. Phase 0 → Production mesh (completed)

| # | Deliverable | Status |
|:---|:---|:---|
| 1 | Split BCs into `pc-*-svc` binaries + Postgres `Repository` | Done — `platform/cmd/pc-{party,policy,billing,claims,audit,integration,gateway}-svc` |
| 2 | Kafka outbox relay | Done — `shared/go/kafka` + `store/postgres` outbox table |
| 3 | Keycloak JWT at gateway | Done — `shared/go/auth` + `KEYCLOAK_URL` |
| 4 | Telebirr sandbox ACL | Done — `integration/telebirr_sandbox.go` + `pc-integration-svc` |
| 5 | PDF schedule / COI / QR sticker | Done — `platform/internal/pdf` (gofpdf + qrcode) |
| 6 | Mesh orchestration | Done — `scripts/mesh-up.sh`, `MEDHEN_MESH=1`, gateway proxy |

Run: `make mesh` after `make infra-up`. E2E verified with PDF issuance (`EIC-MOT-2026-000001-*.pdf`).

---

*End of Implementation Plan v1.0. Update milestone checkboxes as exits land; never renumber stable service/REQ IDs.*
