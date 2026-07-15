# Medhen Platform — Phase 0 Prototype

**Ethiopian Insurance Corporation (EIC)** · End-to-end Motor insurance automation  
**መድህን** — Protector / Insurer

Enterprise-grade Phase 0 pilot: quote → STP underwriting → Telebirr payment → bilingual issuance → FNOL → fast-track settlement.

Authoritative docs: [`docs/prd/implementation_plan.md`](docs/prd/implementation_plan.md) · [PRD](docs/prd/Medhen-Platform-PRD.md) · [Capability Document](docs/prd/Medhen-Platform-Capability-Document.md) · [Service Registry](docs/prd/service-registry.md)

---

## Quick start (monolith — no Docker)

```bash
make build-all
make api          # or: MEDHEN_DOCS_DIR=./data/docs ./bin/medhen-api
./scripts/demo-e2e.sh
```

API: `http://localhost:8080` · OpenAPI: `contracts/openapi/medhen.v1.yaml`

## Production mesh (Postgres + Kafka + microservices)

```bash
cp .env.example .env   # fill TELEBIRR_* when sandbox credentials arrive
make infra-up          # Postgres, Valkey, Redpanda, MinIO, Keycloak
make mesh              # builds + starts pc-*-svc + pc-gateway on :8080
./scripts/demo-e2e.sh
make mesh-down
```

| Process | Port | Role |
|---|---|---|
| `pc-gateway` | 8080 | BFF · Keycloak JWT · routes to mesh |
| `pc-party-mgmt-svc` | 8101 | Party BC |
| `pc-policy-svc` | 8103 | Quote / policy / product / KPIs |
| `pc-billing-svc` | 8107 | Invoices / Telebirr pay |
| `pc-claims-svc` | 8106 | FNOL / fast-track settle |
| `pc-audit-svc` | 8117 | Audit query + Kafka consumer |
| `pc-integration-svc` | 8118 | Telebirr sandbox ACL · SMS |

Set `DATABASE_URL`, `KAFKA_BROKERS`, `REDIS_URL` (see `.env.example`). Without them, services fall back to in-memory store.

**Keycloak JWT:** set `KEYCLOAK_URL=http://localhost:8081` — gateway requires `Authorization: Bearer <token>`. Demo users: `demo-agent` / `medhen-demo`.

**Telebirr sandbox:** set `TELEBIRR_APP_ID`, `TELEBIRR_APP_SECRET`, `TELEBIRR_SHORT_CODE`, `TELEBIRR_BASE_URL` — otherwise mock adapter is used.

**Documents:** bilingual PDF schedule, COI, and QR windshield sticker under `data/docs/*.pdf`.

## Infra backbone

```bash
cd infra && docker compose up -d
```

Brings up Postgres, Valkey, Redpanda (Kafka API), MinIO, Keycloak (`medhen` realm), Jaeger, Mailhog.

| Service | URL |
|---|---|
| Keycloak | http://localhost:8081 (admin/admin) |
| MinIO console | http://localhost:9001 (medhen/medhenmedhen) |
| Jaeger | http://localhost:16686 |
| Mailhog | http://localhost:8025 |

## Web portal

```bash
cp web/.env.local.example web/.env.local
cd web && npm install && npm run dev
```

Open http://localhost:3000 — sign in with `demo-agent` / `medhen-demo` (Keycloak).

## Demo rehearsal (stakeholders)

```bash
make mesh-smoke          # automated: infra + mesh + JWT E2E + Postgres/PDF checks
make demo-rehearse       # same + pointer to runbook
```

Facilitator guide: [`docs/demo/DEMO-RUNBOOK.md`](docs/demo/DEMO-RUNBOOK.md) (15 min)  
Storyboard script: [`docs/demo/STORYBOARD.md`](docs/demo/STORYBOARD.md)  
Seed personas: [`seeds/demo-personas.json`](seeds/demo-personas.json)

**Telebirr live sandbox** (when EIC provides credentials):

```bash
# add TELEBIRR_* to .env, restart mesh, then:
make telebirr-prove
```

## Architecture

**Monolith (dev):** `medhen-api` — all BCs in one process, in-memory or Postgres.

**Production mesh:** separate `pc-*-svc` binaries + `pc-gateway` BFF (`MEDHEN_MESH=1`).

```
pc-web → pc-gateway (:8080, JWT)
           ├─ pc-party-mgmt-svc (:8101)
           ├─ pc-policy-svc (:8103)
           ├─ pc-billing-svc (:8107) ──► pc-integration-svc (Telebirr)
           ├─ pc-claims-svc (:8106)
           └─ pc-audit-svc (:8117) ◄── Kafka (outbox relay)
shared-go: ETB · Ge'ez calendar · i18n · outbox · idempotency · auth · kafka
Postgres (pc_medhen schema) · Valkey · Redpanda · MinIO · Keycloak
```

`medhen-api` remains available as a single-process fallback for local demos.

## Demo storyboard

1. Register party (Amharic name + Region→Kebele address)
2. Seeded Motor product + risk schema (shared-core teaser)
3. Quote with itemized premium (base, factors, VAT 15%, stamp duty)
4. STP auto-accept → bind → Telebirr mock pay
5. Issue `EIC/MOT/{year}/{seq}` + bilingual **PDF** schedule / COI / QR sticker
6. Mobile FNOL (GPS + photos) → fast-track settle + SMS mock
7. Immutable audit trail + KPI tile (GWP / policies in-force)

## Quality bar (Tier-0 patterns)

- Transactional outbox → Kafka relay (when `KAFKA_BROKERS` set)
- Keycloak JWT at gateway (when `KEYCLOAK_URL` set)
- Postgres persistence (when `DATABASE_URL` set)
- Valkey idempotency (when `REDIS_URL` set)
- Tenant isolation (`tenant_id=eic`)
- Audit append on every state change
- Configuration-driven product/rating/UW (not hard-coded LOB forks)
- Contract-first OpenAPI + protobuf registry under `contracts/`

## Layout

| Path | Role |
|---|---|
| `contracts/` | `pc-contracts` — proto, OpenAPI, Avro topics |
| `shared/go/` | `pc-shared-go` — Kernel (BC-MDH-20) |
| `platform/` | Bounded-context packages + `medhen-api` |
| `infra/` | Docker Compose backbone |
| `web/` | `pc-web` customer/agent/staff portals |
| `seeds/` | Motor product seed + demo personas |
| `scripts/demo-e2e.sh` | Facilitator storyboard |
| `scripts/mesh-smoke.sh` | Full mesh smoke (Postgres/Kafka/JWT) |
| `docs/demo/` | 15-min runbook + storyboard |

## License / classification

Confidential — InnoSphere Technologies / EIC design-partner prototype.
