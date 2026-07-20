# Medhen Platform — Modular Monolith

**Ethiopian Insurance Corporation (EIC)** · End-to-end Motor insurance automation
**መድህን** — Protector / Insurer

Quote → STP underwriting → Telebirr payment → bilingual issuance → FNOL → fast-track settlement,
delivered as a single **modular-monolith** Go service (`medhen-api`) with sealed bounded-context modules,
an in-process event bus, and a transactional outbox.

Authoritative docs: [refactor plan](docs/refactor/modular-monolith-plan.md) ·
[PRD](docs/prd/Medhen-Platform-PRD.md) · [Capability Document](docs/prd/Medhen-Platform-Capability-Document.md)

---

## Quick start

```bash
# Stateless only (rating/underwriting) — no dependencies:
make build && ./bin/medhen-api

# Full platform (all 13 modules) — needs Postgres:
make infra-up                                  # Postgres, Valkey, Kafka, Keycloak
export DATABASE_URL="postgres://medhen:medhen@localhost:5432/medhen?sslmode=disable"
export TELEBIRR_WEBHOOK_SECRET="dev-secret"    # enables Telebirr webhook verification
make api                                       # build + run on :8080
```

Auth is **fail-closed**: it activates only when `KEYCLOAK_URL`/`KEYCLOAK_REALM` are set, and never falls
back to an insecure mode. Without Keycloak (dev), the tenant is taken from the `X-Tenant-ID` header.

## End-to-end lifecycle (dev, no auth)

```bash
B=http://localhost:8080; H=(-H Content-Type:application/json -H X-Tenant-ID:eic)
# 1. Register a party
PID=$(curl -s $B/party/parties "${H[@]}" -d '{"full_name":"Abebe Bikila","phone_e164":"+251911000000","national_id":"E1","address":{"region":"Addis Ababa","zone":"Bole","woreda":"03"}}' | jq -r .id)
# 2. Quote (real rating from the product catalog)
QID=$(curl -s $B/policy/quotes "${H[@]}" -d "{\"party_id\":\"$PID\",\"product_code\":\"MOT\",\"coverages\":[\"OD\",\"TPL\"],\"risk_dimensions\":{\"age_band\":\"adult\"}}" | jq -r .ID)
# 3. Bind → underwrite → issue (atomic); billing raises an invoice, document mints the COI, SMS is queued
POL=$(curl -s $B/policy/quotes/$QID/bind "${H[@]}" | jq -r .PolicyNumber)
# 4. FNOL + fast-track settle
CID=$(curl -s $B/claims/claims "${H[@]}" -d "{\"policy_id\":\"...\",\"reserve_minor\":4000000}" | jq -r .ID)
# 5. Read KPIs (real loss ratio) and the immutable audit trail
curl -s $B/reporting/kpis "${H[@]}"; curl -s "$B/audit/logs?limit=20" "${H[@]}"
```

## Architecture

A single process (`cmd/medhen-api`) composes the platform kernel and all bounded-context modules behind
one HTTP edge. Modules are **sealed**: a module may depend only on the shared platform kernel and other
modules' `ports` packages — enforced by [`.go-arch-lint.yml`](.go-arch-lint.yml). Cross-module calls are
in-process; every state change commits its aggregate and a domain event in one transaction (outbox), and
the relay publishes events to the in-process bus (Kafka-ready at the edge).

```
cmd/medhen-api            composition root (registry, HTTP edge, outbox relay)
internal/app              module contract (Module/BackgroundModule) + kernel + registry
internal/platform         kernel: money, eventbus, outbox, database (UoW), idempotency, ids, auth, i18n, config
internal/modules/<bc>     bounded contexts, each: domain / ports / app / adapters / rest / module.go
```

### Bounded-context modules

| Module | Responsibility | Notable |
|---|---|---|
| `rating` | premium calculation | `Calculator` port; VAT + stamp duty; banker's rounding |
| `product` | product/coverage/factor catalog | implements rating's `RateTableProvider` |
| `underwriting` | STP decision | `Decider` port (accept / refer / decline) |
| `party` | individuals & organizations | Amharic names, Region→Kebele address; `Reader` |
| `policy` | quote → bind → issue | atomic issuance; policy no. `EIC/MOT/{year}/{seq}` |
| `billing` | invoices & payments | subscribes `policy.issued`; **Telebirr HMAC webhook** |
| `claims` | FNOL → fast-track settle | validates cover via `policy.Reader` |
| `document` | Certificate of Insurance | generated on `policy.issued` |
| `notification` | SMS/email | background dispatcher via `integration` |
| `integration` | outbound ACL | SMS/email/Telebirr provider ports |
| `reporting` | KPIs | CQRS projection → real loss/combined ratio |
| `audit` | immutable trail | subscribes to **all** events |
| `iam` | app users/roles | `Reader` for authz (tokens verified in `platform/auth`) |

## Web portal

```bash
cp web/.env.local.example web/.env.local   # set KEYCLOAK_*, NEXTAUTH_* (see file)
cd web && npm install && npm run dev        # http://localhost:3000
```

## Infra backbone

`make infra-up` runs Postgres, Valkey, Kafka, and Keycloak (`medhen` realm) via
[`infra/docker-compose.yml`](infra/docker-compose.yml). Keycloak admin: http://localhost:8081 (admin/admin).

## Testing

```bash
make test              # unit tests (-short -race)
make test-integration  # + testcontainers integration tests (requires Docker)
```

## Layout

| Path | Role |
|---|---|
| `cmd/medhen-api/` | composition root / entrypoint |
| `internal/app/` | module contract, kernel, registry |
| `internal/platform/` | shared kernel packages |
| `internal/modules/` | the 13 bounded contexts |
| `web/` | Next.js customer/agent/staff portal |
| `infra/` | Docker Compose backbone |
| `docs/refactor/` | modular-monolith plan |

## License / classification

Confidential — InnoSphere Technologies / EIC design-partner prototype.
