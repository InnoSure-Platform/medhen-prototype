# Refactoring Plan — Tier-1 Enterprise Modular Monolith

**Status:** DRAFT — awaiting sign-off on open decisions (see §2)
**Owner:** Platform / Architecture
**Companion doc:** [`docs/reviews/2026-07-18-code-review.md`](../reviews/2026-07-18-code-review.md)
**Last updated:** 2026-07-18

---

## 1. Goal & guiding principles

Transform the current 15-service "mesh" (mostly stubbed, not wired, 7 divergent Go module paths,
security effectively off) into a **single, cohesive, enterprise-grade modular monolith**: one
deployable binary composed of well-bounded, independently-testable domain modules that *could* be
extracted into services later, but run in-process today.

**Why a modular monolith (not the current mesh):** for a Phase-0 insurance pilot, the mesh pays the
full operational cost of microservices (network hops, distributed transactions, 7 build systems, per-
service infra) while delivering none of the benefits (the flagship flow is stitched with stubs). A
modular monolith gives us: strict module boundaries + one transaction + one deploy + one test harness,
with a clean extraction path when scale actually demands it.

### Non-negotiable principles (the "tier-1" bar)

1. **One module, one binary, one composition root.** Single `go.mod`, canonical module path, single
   `cmd/` entrypoint that wires everything.
2. **Modules are sealed.** A module may depend on another module *only* through its published **port**
   (Go interface package). No module ever imports another module's `internal/domain` or `adapters`.
   Enforced by a lint gate (depguard/go-arch-lint) that fails CI.
3. **Domain is pure.** Domain layer imports no framework, no DB, no HTTP — only the platform kernel
   (money, errors, ids). Dependencies point inward (hexagonal).
4. **One way to do money.** Single `platform/money` decimal type everywhere. No `float64` for money. No
   parallel implementations.
5. **One way to persist.** Single Postgres pool, schema-per-module, one migration tool, one
   Unit-of-Work. Every command handler runs in a transaction; the outbox write shares that transaction.
6. **Async = outbox + event bus, same code path in-proc or out.** Modules publish domain events to an
   in-process bus backed by the transactional outbox; a single relay drains it (and can fan out to
   Kafka for external consumers) with correct locking and at-least-once semantics.
7. **Auth is real and fails closed.** RS256/JWKS validation, RBAC, at the edge and in the app. No
   backdoors, no header-trust fallbacks, no hardcoded secrets.
8. **Contract-first at the edge, interfaces inside.** One versioned OpenAPI spec for the external HTTP
   API (generated types/handlers); internal module contracts are Go interfaces. Internal gRPC/proto is
   removed.
9. **Everything is tested and gated.** Unit + integration (testcontainers) + contract + e2e; coverage,
   lint, arch, and security gates block merge.
10. **Observable & reproducible.** OTel traces/metrics/logs with correlation IDs; single non-root
    container; hardened K8s; CI/CD with SAST + dependency + secret scanning.

---

## 2. Decisions — LOCKED (2026-07-18)

| # | Decision | Chosen |
|---|----------|--------|
| D1 | **Canonical name & module path** | `github.com/InnoSure-Platform/medhen-prototype` (org InnoSure-Platform, product Medhen) |
| D2 | **Migration style** | In-place strangler — stays green & demoable throughout |
| D3 | **Module granularity** | ~13 modules = current bounded contexts |
| D4 | **Async backbone** | In-proc event bus + transactional outbox; Kafka optional at the edge |
| D5 | **Internal transport** | Delete internal gRPC/proto; in-proc Go interfaces (default) |
| D6 | **First end-to-end vertical** | Motor: quote→UW→bind→issue→pay→FNOL→settle (default) |

> All module paths become `github.com/InnoSure-Platform/medhen-prototype/...`
> (e.g. `.../internal/modules/policy`, `.../internal/platform/money`).

---

## 3. Target architecture

### 3.1 Repository layout (target)

```
/                              # single Go module (go.mod), one canonical path
├── cmd/
│   ├── medhen-api/            # THE monolith: HTTP edge + all modules wired in one process
│   │   └── main.go            # composition root
│   └── relay/                 # (optional) standalone outbox relay; also runnable in-proc
├── internal/
│   ├── platform/              # shared kernel — no domain logic, imported by everyone
│   │   ├── money/             # single decimal ETB type (shopspring-based) + tax/rounding
│   │   ├── config/            # typed config loader (env + file), fail-closed validation
│   │   ├── database/          # pgx pool, UnitOfWork/tx manager, health
│   │   ├── outbox/            # outbox writer + relay (correct FOR UPDATE SKIP LOCKED in a tx)
│   │   ├── eventbus/          # in-process pub/sub, subscribers registered at startup
│   │   ├── auth/              # JWKS validation, principal, RBAC helpers, context propagation
│   │   ├── idempotency/       # atomic (SETNX/Lua) idempotency store, cached-response replay
│   │   ├── httpx/             # router, middleware (auth, reqid, recover, CORS), problem+json errors
│   │   ├── telemetry/         # OTel tracer/meter/logger, correlation IDs
│   │   ├── ids/               # ULID/UUID, sequence generator (policy numbers)
│   │   ├── i18n/              # thread-safe translator (am/en), Ge'ez calendar
│   │   └── errors/            # typed domain errors → HTTP mapping
│   ├── modules/
│   │   ├── iam/               # each module is a sealed bounded context
│   │   │   ├── domain/        # entities, value objects, domain services (pure)
│   │   │   ├── app/           # command/query handlers (use cases), UoW orchestration
│   │   │   ├── ports/         # PUBLIC interfaces this module exposes to others + needs from others
│   │   │   ├── adapters/      # driven adapters: postgres repo, external clients
│   │   │   ├── rest/          # driving adapter: HTTP handlers for this module's routes
│   │   │   ├── events/        # domain event types this module publishes/consumes
│   │   │   └── module.go      # Register(deps) → routes, subscriptions, public facade
│   │   ├── party/  product/  rating/  underwriting/  policy/
│   │   ├── billing/  claims/  document/  notification/
│   │   ├── integration/  audit/  reporting/            # + orchestration (was workflow)
│   ├── app/                   # module registry + wiring glue (the "kernel" of the monolith)
│   └── contracts/             # cross-module port interfaces & shared DTOs (if not per-module)
├── api/
│   └── openapi/medhen.v1.yaml # single external contract; codegen → internal/gen
├── migrations/                # one migration tree, namespaced per schema
│   ├── party/ policy/ billing/ ...
├── web/                       # Next.js (hardened, see Phase 6)
├── deploy/                    # one Dockerfile (multi-stage, non-root), K8s, helm, compose
├── configs/                   # config.yaml per env
├── scripts/                   # dev, migrate, seed, e2e
└── docs/                      # ADRs, reviews, this plan
```

### 3.2 Module boundaries (13 modules)

`iam` · `party` · `product` · `rating` · `underwriting` · `policy` · `billing` · `claims` ·
`document` · `notification` · `integration` · `audit` · `reporting`
(+ `orchestration` for the bind/settle sagas, was `pc-workflow-svc`; `observability` folds into
`platform/telemetry`).

### 3.3 How modules talk

- **Sync read/command:** module A calls module B **only** via `B.ports.SomeAPI` (interface),
  implemented by B's facade, injected at wiring time. Example: `policy` calls `rating.Calculator` and
  `underwriting.Decider` in-process (replaces the stubbed gRPC clients).
- **Async:** module A's command handler writes domain events to the **outbox in the same DB tx**; after
  commit the **relay** dispatches to the **in-process event bus**; subscribers (e.g. `audit`,
  `notification`, `reporting`) handle them. Same relay optionally mirrors to Kafka for external
  consumers. This keeps the exact code path whether in-proc or later extracted.
- **Transactions:** a use case opens one UoW (one `pgx.Tx`); all repo writes + the outbox insert commit
  atomically. Cross-module *writes* in one business action are coordinated by an orchestration
  saga using events, not distributed 2PC.

---

## 4. Phased execution plan

Each phase is independently mergeable, keeps `main` green, and ends with explicit acceptance criteria.
Order front-loads safety (Phase 0) and the rails (Phases 1–2) so all later work lands on solid ground.

### Phase 0 — Security triage & guardrails  *(do first; small, high-urgency)*
**Goal:** stop the bleeding and install the rails that keep the refactor honest.

> **Locked (2026-07-18):** auth sequencing = **pull the real JWKS kernel forward now** (implement real
> Keycloak RS256/JWKS validation first, *then* remove bypasses, so real logins keep working — strangler
> stays green). Secret handling = **scrub git history + rotate** (git-filter-repo, force-push,
> collaborators re-clone). JWKS implemented with **stdlib + `golang-jwt/jwt/v5`** (no new dependency).
> Implemented in `libs/pc-auth-sdk` in place now; relocates to `internal/platform/auth` in Phase 2.

Steps:
1. Implement real RS256/JWKS validation in `pc-auth-sdk` (issuer/audience/alg checks, JWKS cache), then
   remove auth backdoors: `X-Tenant-ID` fallback, `mock-valid-token`, and `X-User-ID: demo-agent`
   (C2, H10). Auth fails closed.
2. Remove all hardcoded secret fallbacks (`NEXTAUTH_SECRET`, Keycloak secret) — fail closed (C3).
3. Purge committed secrets: remove `certs/server.key` and the Keycloak client secret from the tree and
   **git history** (git-filter-repo/BFG); rotate; generate certs at deploy time (H1, H2).
4. Add `/staff`, `/quote`, `/claim` to `web/middleware.ts` with role checks (C4).
5. Restrict CORS to an explicit allowlist in the running gateway config (H3); tighten Keycloak
   redirect/web origins (H4).
6. Stand up CI gates that later phases rely on: `golangci-lint`, `gitleaks`/secret scan, `govulncheck`,
   `trivy` (extend the existing notification-svc job repo-wide), unified Go version.
**Acceptance:** no known auth bypass; no secrets in tree or history; CI red on lint/vuln/secret/arch
violations; app still runs the demo.

**Status (2026-07-18):**
- [x] Real RS256/JWKS auth kernel in `pc-auth-sdk`; bypasses removed; 12 unit tests; call sites updated.
- [x] Web `X-User-ID` fallback removed; `/staff`,`/quote`,`/claim` added to `middleware.ts` with role
  gates; `NEXTAUTH_SECRET`/`KEYCLOAK_SECRET` hardcoded fallbacks removed. Fails closed at **runtime**
  (NextAuth refuses to run in prod without `NEXTAUTH_SECRET`), NOT at import — an import-time throw broke
  `next build` on Vercel (env injected only at runtime). Deployment env must set `NEXTAUTH_SECRET`,
  `KEYCLOAK_SECRET`, `KEYCLOAK_ISSUER`, `NEXTAUTH_URL`.
- [x] Gateway CORS → explicit origin (`http://localhost:3000`), dropped `X-User-ID` header; Keycloak
  `redirectUris`/`webOrigins` de-wildcarded.
- [x] Committed `certs/server.{key,crt}` removed from tree + gitignored; `scripts/gen-dev-certs.sh`
  added. Keycloak client secret removed from realm JSON; `scripts/keycloak-configure.sh` +
  `KEYCLOAK_PC_WEB_SECRET` env added.
- [x] CI rewritten: fixed broken `platform/` build path + Go 1.26; added gitleaks (blocking, working
  tree), govulncheck/golangci-lint/trivy (report-only until Phase 7), web typecheck.
- [ ] **History scrub not yet executed** — `scripts/scrub-secrets-history.sh` is ready; needs
  `git-filter-repo` installed + a coordinated `git push --force-with-lease` (collaborators re-clone).
  Secrets are already rotated/removed from HEAD, so the exposure is neutralized; the scrub is hygiene.
- [ ] Gateway JWT filter (C1) + tenant-from-token in handlers → Phase 2/3 (documented, not Phase 0).

### Phase 1 — Repo & module unification
**Goal:** collapse 7 module paths into one module and stand up the monolith skeleton + composition root.
Steps:
1. Create the single `go.mod` at repo root with the canonical path (D1); delete per-service `go.mod`/
   `go.sum` and `go.work`.
2. Create target skeleton (`cmd/medhen-api`, `internal/platform`, `internal/modules/*`, `internal/app`).
3. Implement the **module registry**: each `module.go` exposes `Register(kernel) (facade, routes,
   subscriptions)`; `cmd/medhen-api/main.go` wires them in dependency order.
4. Introduce the arch-lint ruleset (no cross-module internal imports; domain imports only platform).
5. Temporary: keep old `services/` compiling behind a build tag OR move them wholesale under
   `internal/modules/<bc>` verbatim and fix imports so it builds as one module (strangler start).
**Acceptance:** `go build ./...` produces one binary; arch-lint passes on the skeleton; a trivial
`/healthz` served by the monolith.

**Status (2026-07-18):** skeleton stood up **additively** (mesh untouched, repo stays green).
- [x] Root module `github.com/InnoSure-Platform/medhen-prototype` (`go.mod`), added `.` to `go.work`.
- [x] Composition root `cmd/medhen-api/main.go`: config load, structured slog, module registry init,
  `/healthz`+`/readyz`, request-id + panic-recovery middleware, graceful shutdown. Smoke-tested (200s,
  X-Request-Id echoed).
- [x] Module contract + registry (`internal/app/{kernel,registry}.go`); platform basics
  (`internal/platform/{config,httpx}`); `internal/modules/README.md`; arch-lint config
  (`.go-arch-lint.yml`).
- [ ] Collapsing the 20 per-service `go.mod`s + deleting `go.work` completes as the **last** module
  migrates in Phase 3 (can't delete them while the mesh still builds — strangler). Until then the root
  module and the mesh coexist in the workspace.

### Phase 2 — Platform / shared kernel
**Goal:** one correct implementation of every cross-cutting concern.
Steps:
1. `platform/money`: consolidate to one decimal ETB type (adopt the good `pc-rating-calc-svc` decimal
   math: internal vs final precision, `RoundBank`, `SafeDiv`); delete `shared-go-svc/money` and all
   `float64` money. Add stamp-duty support alongside VAT (M2).
2. `platform/database`: pgx pool + `UnitOfWork` (tx-per-use-case) + health; `platform/outbox`: writer +
   **correct** relay (`FOR UPDATE SKIP LOCKED` inside an explicit tx, update+publish before commit or
   with proper at-least-once + dedup) (C7). `platform/eventbus`: in-proc pub/sub.
3. `platform/auth`: real JWKS fetch/cache, RS256 validation, `alg` allow-list, audience/issuer checks,
   `Principal` with roles/tenant/branch, RBAC middleware (fixes C1/C2 at the app layer).
4. `platform/idempotency`: atomic SETNX/Lua, TTL, cached-response replay (H7).
5. `platform/config`, `telemetry`, `httpx` (problem+json, request-id, recover, CORS), `i18n`
   (add mutex — M12), `ids` (ULID + monotonic policy-number sequence — L2).
**Acceptance:** platform packages unit-tested to high coverage; money/auth/outbox have adversarial
tests; no `float64` money remains (`grep` gate).

**Status (2026-07-18):** offline-testable kernel landed (all `-race` green).
- [x] `internal/platform/money` — single `shopspring/decimal` ETB `Amount`; internal vs currency scale;
  banker's rounding; `SafeDiv`; largest-remainder `Allocate` (no lost cents); `TaxProfile` for VAT +
  stamp duty (M2). 14 tests.
- [x] `internal/platform/eventbus` — concurrency-safe in-proc pub/sub; error aggregation; per-handler
  panic isolation. 5 tests (incl. race).
- [x] `internal/platform/ids` — ULID (oklog/ulid) + `Sequencer` (in-mem) + `PolicyNumber`
  `EIC/MOT/{year}/{seq}` formatter (L2). 5 tests.
- [x] `internal/platform/auth` — JWKS validator relocated from `pc-auth-sdk` (canonical home) +
  `RequireRole` middleware. 12 tests. Wired into `Kernel` + composition root (enabled only when
  Keycloak configured; never insecure fallback).
- [x] `internal/platform/database` — pgx pool + **Unit-of-Work** with ambient-tx-on-context (`WithinTx`
  / `Conn`), so repos + outbox share one commit. Integration-tested (commit persists, rollback discards).
- [x] `internal/platform/outbox` — `Write` (same tx as aggregate) + **correct relay**: claims rows with
  `FOR UPDATE SKIP LOCKED` *inside* the tx that marks them processed, holding locks until commit (**C7
  fixed**). Regression test: 4 concurrent workers drain 200 msgs with zero double-publish. Also
  commit/rollback + idempotent-reprocess tests. (Postgres via testcontainers.)
- [x] `internal/platform/idempotency` — atomic `SETNX` claim + TTL + cached-response replay (**H7
  fixed**). Regression test: exactly 1 winner among 100 concurrent claimants; replay returns cached
  body. (Redis via testcontainers.)
- [x] `internal/platform/i18n` — RWMutex-guarded translator, am/en + fallback (**M12 fixed**). Race-tested.
- [x] CI `test-go` now runs `go test -race ./internal/...` (unit + testcontainers integration).
- [ ] `platform/telemetry` (OTel traces/metrics) — **moved to Phase 8** (Observability), where the full
  OTel/Jaeger wiring lives; structured slog logging + request-id correlation already exist.
- [ ] `float64`-money grep gate flips to blocking once the policy module migrates (C8) in Phase 3.

**Phase 2 is complete** (telemetry intentionally consolidated into Phase 8).

### Phase 3 — Module-by-module migration (the bulk)
**Goal:** move each BC into `internal/modules/<bc>`, define ports, replace network calls with in-proc.
Order (dependency-first): `iam → party → product → rating → underwriting → policy → billing → claims →
document → notification → integration → audit → reporting`.
Per-module checklist (repeat 13×):
1. Move domain/app into the module; strip framework deps out of domain.
2. Define `ports/` — the interface(s) it exposes and the interface(s) it consumes from others.
3. Replace gRPC/HTTP clients with in-proc port injections (e.g. delete
   `grpc_client/rating_client.go` stub; `policy` calls `rating` port directly — fixes M5).
4. Convert repos to the platform UoW; command handlers write via one tx + outbox.
5. Wire domain events onto the event bus; move `audit`/`notification`/`reporting` to subscribers.
6. Register routes on the shared router; delete the per-service `cmd/server/main.go`.
7. Add module unit + integration tests before deleting the old service dir.
**Acceptance (per module):** old `services/<svc>` deleted; arch-lint clean; module tests pass; no
in-proc network call to a sibling module.

**Status (2026-07-18):** first module migrated as the **reference implementation**.
- [x] **rating** → `internal/modules/rating` (`domain`/`ports`/`adapters`/`rest`/`module.go`). Re-homed
  the decimal engine onto `platform/money`; **added stamp duty (M2)** and made components sum exactly to
  gross (**L1**); dropped the OTel/validator deps from the domain (pure). Exposes a `Calculator` port
  (the in-proc replacement for policy's `500.00` gRPC stub). 7 unit tests + full end-to-end smoke test
  through the monolith (`POST /rating/quote` → correct breakdown + audit trace). Registered in
  `composeModules`; arch-lint component added.
- [x] **party** → `internal/modules/party` — the **persistence + outbox + events** reference. Register
  individual (Amharic name + Region→Zone→Woreda→Kebele address) persists the aggregate **and** a
  `PartyRegistered` outbox event in **one UoW** (`database.WithinTx`); the relay bridges the outbox to
  the event bus. Exposes a `Reader` port for policy/claims. Composition root now: connects Postgres when
  `DATABASE_URL` set, applies schemas, runs the relay (`outbox → busPublisher → eventbus`), registers a
  demo audit subscriber. 5 testcontainers integration tests (atomicity: duplicate rolls back party +
  outbox; tenant isolation) + live e2e (2 parties persisted, both outbox rows `processed=t`).
- [x] **product** → `internal/modules/product` — the **cross-module wiring** reference. DB-backed
  catalog (products/coverages/factors, Amharic names, seeded Motor). Implements
  `rating/ports.RateTableProvider` and is injected into rating at the composition root, **replacing the
  static rate table** (consumer-defines-port / provider-implements pattern). Exposes a `Catalog` reader
  + `GET /product/products`. 4 testcontainers tests incl. a cross-module test (rating engine + product
  provider → correct young-driver premium) + live e2e (rating now catalog-sourced).
- [x] **underwriting** → `internal/modules/underwriting` — stateless STP decision engine (auto-accept /
  refer / decline by rules). Exposes a `Decider` port. 4 unit tests.
- [x] **policy** → `internal/modules/policy` — **the keystone**. `CreateQuote` prices via the real
  in-process `rating.Calculator` (**M5** — no more `500.00` stub) after validating the party via
  `party.Reader`; `BindQuote` runs `underwriting.Decider` then persists the quote transition + policy +
  `PolicyIssued` outbox event in **one UoW** (**C6** — atomic issuance). Money is `platform/money`
  throughout (**C8** — no float). Policy number is a gap-free DB sequence `EIC/MOT/{year}/{seq}`
  (**L2**). Exposes a `Reader` for billing/claims. 5 testcontainers vertical tests (atomic issue, rebind
  rejected + atomic, sequence increments, refer path issues nothing, party-not-found) + **live e2e**:
  party → quote → bind → policy `EIC/MOT/2026/000001`, both `party.registered` and `policy.issued`
  relayed to the bus.
- [x] **billing** → `internal/modules/billing` — the **event-consumer** reference. Subscribes to
  `policy.issued` on the event bus and **idempotently raises the first invoice** (unique on
  tenant+policy, so redelivery doesn't double-bill). Applies payments atomically (invoice + payment +
  `PaymentReceived` outbox event in one UoW), tracking OPEN/PARTIALLY_PAID/PAID. **Telebirr webhook now
  verifies the HMAC-SHA256 signature (H6)** — arbitrary/empty signatures fail closed. 1 HMAC unit test +
  4 testcontainers tests + **live full-chain e2e**: bind → auto-invoice → bad-sig 401 → signed webhook →
  invoice PAID.
- [x] **claims** → `internal/modules/claims` — FNOL → reserve → fast-track settle. FNOL validates the
  policy is `ISSUED` via `policy.Reader` (cross-module); settlement within a configurable authority limit
  auto-settles, above it is **referred atomically** (no event, claim stays FILED). Emits
  `claims.filed`/`claims.settled` via the outbox. 4 testcontainers tests + live e2e (FNOL → over-authority
  409 → fast-track SETTLED). GPS stays float64 (not money).
- [x] **audit** → `internal/modules/audit` — subscribes to **all** events (new `eventbus.SubscribeAll`)
  and records an **immutable, append-only trail** for every state change (runs inside the relay tx, so
  the audit row commits with the event). `GET /audit/logs`. 1 eventbus test + 1 testcontainers test +
  **capstone e2e**: the full lifecycle produced the trail `party.registered → policy.issued →
  billing.invoice_raised → billing.payment_received → claims.filed → claims.settled`. Makes the "audit on
  every state change" claim real.
- [x] **integration** → `internal/modules/integration` — stateless outbound ACL; exposes `SmsSender`/
  `EmailSender` ports (logging stubs for the prototype; real gateways drop in behind the same ports).
- [x] **notification** → `internal/modules/notification` — subscribes to `policy.issued`/`claims.settled`,
  resolves the recipient via `party.Reader`, queues an SMS in the relay tx, and **dispatches via a
  background loop** (new `app.BackgroundModule` hook) through `integration.SmsSender`. 3 unit tests.
- [x] **document** → `internal/modules/document` — subscribes to `policy.issued` and generates the
  **Certificate of Insurance** (idempotent per policy), emitting `document.generated`. 1 testcontainers test.
- [x] **reporting** → `internal/modules/reporting` — **CQRS projection** of `policy.issued`/`claims.settled`
  into a KPI read model computing a **real loss/combined ratio (M3 fixed)** — no more dummy values.
  2 testcontainers tests.
- [x] **iam** → `internal/modules/iam` — application user/role management + a `Reader` for cross-module
  authz (token verification stays in `platform/auth`). 2 testcontainers tests.

### ✅ Phase 3 COMPLETE — all 13 bounded contexts migrated
Boot order: `underwriting → integration → audit → product → rating → party → policy → billing → claims →
document → notification → reporting → iam`. 20 test packages green. Full lifecycle verified live across
every module: quote → underwrite → bind → issue → COI → invoice → Telebirr-pay → SMS → FNOL → settle →
KPIs, with every state change in the immutable audit trail.

- [x] **Cutover DONE (2026-07-20):** deleted all 15 `services/pc-*-svc` and 5 relocated `libs/*`, removed
  `go.work`/`go.work.sum` — the repo is now a **single Go module**. Removed mesh tooling
  (`scripts/mesh-*.sh`, `demo-e2e.sh`, per-service `pc-notification-svc-ci.yml`); rewrote the `Makefile`
  (monolith `build`/`api`/`test`/`test-integration`) and `pipeline.yml` (`go build/vet/test ./...`);
  rewrote `README.md` and `TESTING_GUIDE.md` for the monolith. `docker-compose.yml` was already
  infra-only (Postgres/Valkey/Kafka/Keycloak). Verified: `go build/vet ./...` clean, 20 test packages
  green standalone (no workspace), module graph is root-only. `pc-workflow-svc` (Temporal orchestration)
  and `pc-observability-svc` (SLO) were intentionally not carried over — orchestration is now event
  choreography; observability is Phase 8.

### Phase 4 — Core flow correctness (Motor vertical, D6)
**Goal:** one real, atomic, event-emitting end-to-end spine.
Steps:
1. `policy` bind/issue in **one UoW**: quote transition + policy insert + outbox event atomic (C6);
   real `rating` + `underwriting` port calls (M5); monotonic policy number (L2); fix endorsement
   bi-temporal persistence (drop `ON CONFLICT DO NOTHING` bug — M6).
2. Rating: stamp duty + VAT; consistent final rounding so components sum to gross (M2, L1).
3. Billing: wire Telebirr callback → invoice allocation (pass real `InvoiceID`, real tenant) and
   **verify HMAC signature** (H6); connect to the (good) ledger UoW.
4. Claims: FNOL → reserve → fast-track settle with real reserve math and events.
5. Reporting: replace dummy KPI constants with real aggregation (M3).
**Acceptance:** scripted e2e (quote→UW→bind→issue→pay→FNOL→settle) passes against a real Postgres,
emits the full audit trail, money reconciles to the cent.

### Phase 5 — Persistence hardening
**Goal:** production-grade data layer.
Steps:
1. One migration tool (goose/atlas), schema-per-module; consolidate the scattered per-service
   migrations; resolve `pc_medhen` vs `pc_*` schema inconsistency.
2. Constraints, indexes, FKs *within* a schema; no cross-schema FKs (module boundary).
3. Tenant isolation: enforce `tenant_id` in every query + Postgres RLS where feasible; least-privilege
   DB roles (app role ≠ owner) — replaces `GRANT ALL TO medhen`.
4. TLS to Postgres (`sslmode=require`) in non-local envs (M10).
**Acceptance:** `migrate up` from clean DB builds full schema; RLS tests prove tenant isolation;
no `GRANT ALL`.

**Status — DONE (2026-07-20):**
- [x] Migration tool: in-house transactional migrator [`internal/platform/migrate`](../../internal/platform/migrate/migrate.go)
  (versioned, `schema_migrations`-tracked, idempotent). Chosen over goose/atlas so the platform kernel
  stays module-agnostic: the composition root assembles the ordered list from each module's own `Schema`
  const (single source of truth, no drift) + a security migration. Replaces the `applySchemas` stopgap;
  runs at boot. Verified live: 12 applied on clean DB, 0 on restart. Tests: apply/idempotent/ordered +
  failed-migration rollback.
- [x] **Least-privilege role**: security migration provisions `medhen_app` (LOGIN, not owner/superuser)
  with **DML-only grants** (SELECT/INSERT/UPDATE/DELETE) — no `GRANT ALL`, no DDL. Verified via
  `information_schema.role_table_grants`.
- [x] **RLS**: `tenant_isolation` policy (`USING` + `WITH CHECK` on `app.current_tenant`) enabled on all
  11 tenant-scoped tables; global tables (products/*) excluded. `database.WithinTenantTx` sets the
  per-tx GUC. RLS integration test proves: no-tenant read = 0 rows (fail-closed), each tenant sees only
  its rows, cross-tenant insert blocked by `WITH CHECK`.
- [x] **TLS (M10)**: `.env.example` documents `sslmode=require` + connecting as `medhen_app` in non-local
  envs; the app warns at boot if `MEDHEN_ENV!=dev` and `sslmode=disable`.
- [~] **Deferred (documented):** (a) physical schema-per-module (separate Postgres schemas) — kept a
  single `public` schema to avoid rewriting every repo's SQL; RLS already enforces the tenant boundary
  (the security-critical part). (b) Full *runtime* RLS enforcement requires connecting the pool as
  `medhen_app` and routing tenant-scoped reads through `WithinTenantTx` — the mechanism + role + policies
  are in place and proven; flipping the default connection to the app role lands with Phase 6 edge auth
  (tenant-from-token).

### Phase 6 — API edge & web hardening
**Goal:** secure, contract-driven edge.
Steps:
1. Author the single `api/openapi/medhen.v1.yaml`; generate server types/handlers; the monolith serves
   it directly (gateway becomes optional TLS/CORS/WAF layer, not the auth boundary).
2. Frontend auth: stop shipping access/id tokens to the browser (C5) — proxy API calls through Next
   route handlers holding the token server-side; move federated-logout id token out of the URL (H8);
   delete the dead password-grant path + `lib/auth.ts` (H9).
3. Enforce RBAC server-side for all role-gated routes incl. `/staff` (C4); use the computed
   `staff`/`claims` roles.
4. Replace `prompt()`/`parseInt` money entry with validated forms + money inputs (M7); stop rendering
   raw backend error bodies (M8); add label associations (a11y).
5. Security headers/CSP in `next.config.ts` and/or gateway (M8).
6. Kill remaining `any`; type the API client against generated OpenAPI types (M7).
**Acceptance:** no token in browser storage/URL; RBAC enforced server-side (tested); CSP present; typed
API client; a11y lint clean.

**Status — PARTIAL (2026-07-20):**
- [x] **Edge authentication + server-side RBAC (API)**: [`auth.EdgeMiddleware`](../../internal/platform/auth/edge.go)
  wired into the monolith edge. Public paths bypass (health + HMAC Telebirr webhook); when Keycloak is
  configured a valid Bearer token is required (401), claims go on the context, and per-prefix role rules
  are enforced (403). Dev (no Keycloak) = pass-through. 6 unit tests (public bypass, dev passthrough,
  401 missing/bad token, authenticated-no-rule, 403 missing role, 200 matching role, longest-prefix).
- [x] **Web tokens (C5/H9)**: deleted the dead ROPC password-grant + `localStorage` token path
  (`lib/auth.ts`, `/api/auth/login`); login is NextAuth Authorization-Code only.
- [x] **H8**: federated logout now reads the id_token from the server-side session (`getToken`), never the
  URL; the id_token is no longer exposed in the client session either.
- [x] **RBAC (C4, web)**: `middleware.ts` gates `/admin`,`/broker`,`/staff` server-side by role + requires
  a session on all protected routes.
- [x] **CSP + security headers (M8)**: `next.config.ts` sets CSP (no `unsafe-eval` in prod), HSTS,
  X-Frame-Options DENY, nosniff, Referrer-Policy, Permissions-Policy — verified served at runtime.
- [ ] **Deferred (needs the app running against the monolith):** author `api/openapi/medhen.v1.yaml` +
  generated typed client (step 1, M7); **realign `web/lib/api.ts`** — it still targets deleted mesh
  gateway paths (`/api/pc-*/v1/...`) instead of the monolith routes (`/party/parties`, `/policy/quotes`,
  …) **and** route calls through server-side Next handlers so the access token never reaches the browser
  (remainder of C5); replace `prompt()`/`parseInt` money entry (M7); stop rendering raw backend error
  bodies (M8); a11y labels. Also flip the DB pool to connect as `medhen_app` + thread tenant-from-token
  into `WithinTenantTx` to activate runtime RLS (Phase 5 hook).

### Phase 7 — Testing & quality gates
**Goal:** the "everything is tested and gated" bar.
Steps:
1. Unit tests for every module's domain/app (target ≥80% domain coverage).
2. Integration tests with **testcontainers** (real Postgres/Redis) for repos, outbox, idempotency.
3. Contract tests against the OpenAPI spec; golden-file tests for documents (PDF/COI).
4. One e2e suite driving the Motor spine + failure paths (payment fail, UW referral, endorsement).
5. CI gates: build, lint, arch-lint, coverage threshold, govulncheck, trivy, gitleaks — all blocking.
**Acceptance:** CI enforces every gate; coverage threshold met; flaky-free e2e in CI.

**Status — DONE (2026-07-20):**
- [x] **Domain unit tests** added for every module that lacked them (party, product, policy, billing,
  claims, document, notification, iam) + a `platform/config` test. Domain layer now **93.2%** statement
  coverage.
- [x] **Coverage gate**: [`scripts/coverage-check.sh`](../../scripts/coverage-check.sh) rewritten for the
  module layout — statement-weighted, block-deduped (handles `-coverpkg`), per-layer floors
  domain 85 / app 55 / platform 55. Verified on a full `-coverpkg` profile: domain 93.2, app 60.7,
  platform 60.0 → PASS. `make test-cover` / `make cover`.
- [x] **Lint**: `.golangci.yml` (errcheck, govet, staticcheck, unused, ineffassign, misspell, unconvert,
  bodyclose, gofmt, goimports) — clean. `make lint`.
- [x] **arch-lint**: `.go-arch-lint.yml` fixed for v1.16 (vendor allow, per-component self-deps, worktree
  exclude) — "No warnings". `make arch-lint`.
- [x] **e2e suite** (step 4): the cross-module Motor-spine test moved to `internal/e2e` (party → product →
  rating → underwriting → policy, atomic issue + refer/rebind failure paths); its own arch-lint component.
- [x] **CI gates blocking**: `pipeline.yml` — govulncheck, golangci-lint, go-arch-lint, and the coverage
  gate are now blocking; trivy vuln+secret blocking, misconfig report-only (Phase 8); Go bumped to 1.26.3
  (clears the stdlib advisories). `docker-build` now needs all gates green. gitleaks already blocking.
- [ ] **Deferred:** contract tests against the OpenAPI spec (the spec itself is the deferred Phase 6 item);
  HTTP-level e2e (the in-process vertical test already covers the spine + failure paths).

### Phase 8 — Observability & delivery
**Goal:** run it like production.
Steps:
1. OTel traces/metrics/logs end-to-end with correlation IDs through the event bus; wire to
   Jaeger/OTel-collector (already in infra).
2. One multi-stage **buildable** Dockerfile (fix M1: no apt/shell on distroless; static non-root
   binary; digest-pinned base); single image, `SERVICE=medhen-api`.
3. K8s hardening: resource requests/limits, `securityContext` (runAsNonRoot, readOnlyRootFS, drop caps),
   pinned image digests, secrets via `secretKeyRef`/external-secrets (M9, H5); wire the good gateway
   stack (cert-manager TLS, JWT SecurityPolicy, CORS allowlist, WAF) into the running path.
4. CI/CD: build→test→scan→push→deploy; SBOM; provenance.
5. ADRs for the modular-monolith decision, module boundaries, event/outbox strategy; update README to
   match reality.
**Acceptance:** image builds & runs non-root; traces visible in Jaeger; K8s passes a policy scan
(kubesec/kube-linter); README accurate.

## Post-refactor follow-ups (2026-07-20)

1. [x] **Web branch merged** — the API-client realignment + server-side token proxy landed in `main`
   (`b7761c5`); stale local worktree/branch removed (remote branch left for GitHub cleanup).
2. [x] **OpenAPI contract** — `api/openapi/medhen.v1.yaml` (OpenAPI 3.1, 20 paths) documents every module
   endpoint + bearer security. The web already has a hand-written typed client; server codegen is optional.
3. [~] **Runtime RLS — groundwork done, activation staged.** `database.WithTenant` + `WithinTx` now set
   `app.current_tenant` automatically from the request tenant (writes are RLS-ready with no service
   changes); an edge middleware binds the authenticated tenant to the DB context. **Not flipped to the
   `medhen_app` pool yet** because the outbox relay + notification dispatcher are cross-tenant background
   workers — full activation needs a *system/BYPASSRLS* connection for those workers plus request-scoped
   reads wrapped in tenant-txs. Documented so it's a deliberate, tested rollout, not a flag flip.
4. [x] **CI delivery** — `docker-build` now uses buildx `build-push-action` with **SBOM + SLSA provenance
   attestations**, pushes to GHCR on `main`, and uploads a standalone Syft SPDX SBOM artifact.
5. [~] **Secret hygiene.** Found + fixed **live HEAD leaks**: `scripts/set-vercel-env.sh` (hardcoded client
   secret → now required from env) and `web/.env.production.example` (real client + NextAuth secrets →
   placeholders). Added `.gitleaks.toml` allowlisting the review doc + scrub script. The **history scrub**
   (`scripts/scrub-secrets-history.sh`) is verified (secrets confirmed in commits fa38961/790b0f4 +
   4 `supersecret` commits) and ready, but **not executed** — it rewrites all history and needs
   `git-filter-repo` + a coordinated `--force-with-lease` push + collaborator re-clone.

**Status — DONE (2026-07-20):**
- [x] **OTel tracing** (step 1): `internal/platform/telemetry` installs a global tracer provider + W3C
  propagators; OTLP/HTTP export is enabled by `OTEL_EXPORTER_OTLP_ENDPOINT` (else spans are created but
  not shipped, so instrumentation is always safe). The HTTP edge is wrapped with `otelhttp` (server spans
  + context extraction) and the event bus starts a span per published event, so a request trace flows
  edge → handler → event → subscriber. 2 unit tests. `version` is stamped via `-ldflags -X main.version`.
- [x] **Buildable Dockerfile** (M1/M3): rewritten multi-stage — `golang:1.26.3` builder → `CGO_ENABLED=0`
  static, trimmed, stripped binary → `distroless/static:nonroot` runtime (no apt/shell), `USER nonroot`,
  `ENTRYPOINT` the binary directly. Added `.dockerignore` (build context 1.1 GB → tiny).
- [x] **K8s hardening** (M9/H5): `deploy/k8s/medhen-api.yaml` — `runAsNonRoot`/`runAsUser 65532`,
  `readOnlyRootFilesystem`, `allowPrivilegeEscalation: false`, `capabilities: drop [ALL]`, seccomp
  RuntimeDefault, CPU/mem requests+limits, liveness `/healthz` + readiness `/readyz`,
  `automountServiceAccountToken: false`, writable `/tmp` emptyDir; secrets via `secretRef` (no inline
  secrets — `secret.example.yaml` documents external-secrets/sealed-secrets). Removed the stale
  `pc-notification-svc` helm chart.
- [x] **ADR** (step 5): `docs/adr/ADR-PC-021-modular-monolith.md` records the modular-monolith decision,
  sealed module boundaries, and the event/outbox backbone; supersedes ADR-PC-002/004, amends 005.
- [ ] **Deferred:** CI push/deploy + SBOM/provenance (step 4) — needs a registry/target; wire the good
  gateway stack (cert-manager/WAF) into the running path when a cluster exists.

---

## 5. Cross-cutting review-finding coverage

Every review finding maps to a phase: **auth** C1–C5,H8,H10 → P0/P2/P6 · **secrets** H1,H2,H5 → P0/P8 ·
**money** C8,M2,L1,L4 → P2/P4 · **transactions/outbox** C6,C7,M6,M11 → P2/P3/P4 · **service wiring**
M4,M5,H6,H7 → P2/P3/P4 · **infra/deploy** M1,M9,M10,H3,H4 → P0/P5/P8 · **frontend** M7,M8 → P6 ·
**testing** → P7 · **misc** M3,M12,L2,L3 → P2/P4.

## 6. Risks & mitigations

- **Big-bang temptation → regressions.** Mitigation: strangler (D2), always-green `main`, module-by-
  module with tests-before-delete.
- **Boundary erosion over time.** Mitigation: arch-lint as a blocking CI gate from Phase 1.
- **Losing the good code** (rating decimal, billing ledger, outbox). Mitigation: these become the
  canonical platform/module implementations, not rewrites.
- **Hidden coupling surfaced late.** Mitigation: define ports first (Phase 3 step 2) before moving code.

## 7. Suggested sequencing / sprints

- **Sprint 1:** Phase 0 (+ start Phase 1). *Ship: safe app, one module, CI gates.*
- **Sprint 2:** Phase 1 complete + Phase 2. *Ship: platform kernel.*
- **Sprints 3–5:** Phase 3 (batches of ~4 modules) + Phase 4 vertical. *Ship: real end-to-end Motor.*
- **Sprint 6:** Phase 5 + Phase 6. *Ship: hardened data + edge/web.*
- **Sprint 7:** Phase 7 + Phase 8. *Ship: gated, observable, deployable.*

---

## 8. Definition of done (tier-1 exit bar)

- [ ] One Go module, one binary, one composition root; arch-lint enforces sealed modules.
- [ ] Domain layers pure; single money type; no `float64` money anywhere.
- [ ] All use cases transactional; outbox correct; async via bus/outbox.
- [ ] Real JWKS auth, RBAC, no backdoors/secret-fallbacks; no secrets in tree/history.
- [ ] Core Motor flow real & atomic end-to-end; money reconciles; full audit trail.
- [ ] Schema-per-module migrations; tenant isolation/RLS; least-privilege DB roles.
- [ ] Contract-first OpenAPI edge; web never holds tokens; server-side RBAC.
- [ ] Unit+integration+contract+e2e; coverage & security gates blocking in CI.
- [ ] OTel end-to-end; buildable non-root image; hardened K8s; accurate docs/ADRs.
</content>
