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
  gates; `NEXTAUTH_SECRET`/`KEYCLOAK_SECRET` hardcoded fallbacks removed (fail closed via `requireEnv`).
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
- [ ] Old `services/pc-*-svc` (rating, party, product, underwriting, policy) kept until **cutover**.
- [ ] Remaining 6 modules: `iam`, `billing`, `claims`, `document`, `notification`, `integration`,
  `audit`, `reporting`.

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

### Phase 7 — Testing & quality gates
**Goal:** the "everything is tested and gated" bar.
Steps:
1. Unit tests for every module's domain/app (target ≥80% domain coverage).
2. Integration tests with **testcontainers** (real Postgres/Redis) for repos, outbox, idempotency.
3. Contract tests against the OpenAPI spec; golden-file tests for documents (PDF/COI).
4. One e2e suite driving the Motor spine + failure paths (payment fail, UW referral, endorsement).
5. CI gates: build, lint, arch-lint, coverage threshold, govulncheck, trivy, gitleaks — all blocking.
**Acceptance:** CI enforces every gate; coverage threshold met; flaky-free e2e in CI.

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
