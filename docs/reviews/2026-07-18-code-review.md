# Medhen / InnoSure Platform — Code Review

**Date:** 2026-07-18
**Scope:** Full prototype — 15 Go microservices (`services/`), shared libs (`libs/`), Next.js frontend (`web/`), infra/deploy (`infra/`, `deploy/`, `Dockerfile`, `.github/`).
**Reviewers:** Automated multi-agent review (Go services, web frontend, infra/security) + direct inspection of auth/money/wiring.

> **Branding note:** the repo is inconsistently branded — README says "Medhen Platform"; Go module
> paths mix `github.com/medhen/…`, `medhen.com/platform/…`, and `github.com/InnoSure-Platform/…`.
> Settle on one canonical name/module prefix.

---

## Executive summary

The **scaffolding and patterns are genuinely strong** — DDD/hexagonal layering, CQRS, transactional
outbox, a `shopspring/decimal` money pipeline with banker's rounding, circuit breakers. But depth is
**very uneven**: a few flows are real (billing ledger, rating engine, claims settlement domain), the
flagship quote→bind→issue→pay path is stitched together with stubs, and the security layer is
**effectively non-functional end-to-end** despite documentation claiming otherwise.

**Verdict:** demo-grade prototype, **not** the "Tier-0" the README claims. Must not leave an isolated
sandbox until the Critical authentication findings are fixed.

---

## Severity index

| ID | Severity | Area | Summary |
|----|----------|------|---------|
| C1 | Critical | Infra | Live Envoy gateway has no JWT filter — all routes unauthenticated |
| C2 | Critical | Auth SDK | `X-Tenant-ID` fallback + `mock-valid-token` backdoor + HMAC-vs-RS256 mismatch |
| C3 | Critical | Web | Hardcoded `NEXTAUTH_SECRET` / Keycloak secret fallbacks → forgeable sessions |
| C4 | Critical | Web | `/staff` (most privileged UI) has no server-side protection |
| C5 | Critical | Web | Access + id tokens exposed to the browser |
| C6 | Critical | Policy | Non-atomic policy issuance (quote + policy in separate txns) |
| C7 | Critical | Claims | Outbox relay lock is a no-op → duplicate event publishing |
| C8 | Critical | Policy | `float64` money in the policy aggregate |
| H1 | High | Security | Committed TLS private key (`certs/server.key`) |
| H2 | High | Security | Keycloak confidential client secret committed |
| H3 | High | Infra | Wildcard CORS on live gateway (`.*` + `Authorization`) |
| H4 | High | Infra | Wildcard `redirectUris`/`webOrigins` on Keycloak clients |
| H5 | High | Infra | Hardcoded DB/admin creds in K8s manifests + Terraform |
| H6 | High | Billing | Telebirr callback unreachable (`InvoiceID: nil`) + signature not verified |
| H7 | High | SDK | Idempotency SDK racy (GET-then-SET, TTL 0 = never expires) |
| H8 | High | Web | id token leaked via logout URL query string; CSRF-able |
| H9 | High | Web | Dead password-grant endpoint returns refresh token to browser |
| H10 | High | Web | `X-User-ID: demo-agent` sent when no token present |
| M1 | Medium | Infra | Dockerfile can't build (apt/shell on distroless base) |
| M2 | Medium | Policy | Stamp duty never computed; cancellation refund hardcoded |
| M3 | Medium | Reporting | KPI `LossRatio`/`CombinedRatio` are dummy constants |
| M4 | Medium | Docs | README "in-memory fallback" false — services panic without DB |
| M5 | Medium | Policy | Rating client stubbed — returns hardcoded `500.00` (float) |
| M6 | Medium | Policy | Endorsement history corruption (`ON CONFLICT DO NOTHING`) |
| M7 | Medium | Web | Pervasive `any`; money fields → `NaN`; `prompt()`-driven money ops |
| M8 | Medium | Web | Backend error bodies rendered to users; no CSP/security headers |
| M9 | Medium | Infra | No K8s resource limits / securityContext; `:latest` images |
| M10 | Medium | Infra | Plaintext Kafka + `sslmode=disable` everywhere |
| M11 | Medium | Policy | Double JSON-encoding in policy-local outbox (dead code) |
| M12 | Medium | i18n | `InMemoryTranslator` documented thread-safe but has no lock |
| L1–L4 | Low | Various | See below |

---

## Critical

### C1 — Live Envoy gateway has no JWT validation
`infra/pc-gateway/envoy.yaml:61-67` — the config actually run locally has only `cors` + `router` HTTP
filters; no `envoy.filters.http.jwt_authn`. Every route (claims, billing, policy, party, IAM) is
proxied with zero token verification. `.env.example:12` ("gateway validates Bearer tokens") is false
for this config. The K8s `SecurityPolicy` (`infra/pc-gateway/k8s/security-policy.yaml`) has a correct
JWT provider — but that stack isn't what runs.

### C2 — Auth SDK can be trivially bypassed
`libs/pc-auth-sdk/auth.go`:
- Lines 40-47: no `Authorization` header → request accepted if `X-Tenant-ID` set (tenant impersonation).
- Lines 61-64: HMAC `SecretKey` validation against tokens Keycloak signs with RS256; no `alg` check,
  no JWKS — cannot validate real Keycloak tokens.
- Lines 68-71: hardcoded backdoor — literal `mock-valid-token` accepted, mapped to `tenant-test-123`,
  compiled into production binaries.

### C3 — Hardcoded secret fallbacks in web auth
`web/app/api/auth/[...nextauth]/route.ts:99` — `secret: process.env.NEXTAUTH_SECRET || "supersecret123"`.
If the env var is ever unset, session JWTs are signed with a public string → anyone can forge a
`role: "admin"` session. Same anti-pattern for the Keycloak secret (`route.ts:13,43-44`).

### C4 — `/staff` has no server-side protection
`web/middleware.ts:28-34` matcher covers only `/customer`, `/broker`, `/admin`. `web/app/staff/**`
(underwriting approve/decline, claim settlement, reserve adjustment, EOD reconciliation, policy
cancel/endorse/renew) is gated only by a client-side redirect in `components/AuthProvider.tsx:51-56`,
trivially bypassed. `/quote` and `/claim` are also unprotected server-side.

### C5 — Access + id tokens exposed to the browser
`web/app/api/auth/[...nextauth]/route.ts:87-96` copies `accessToken`/`idToken` into the
client-readable session; `AuthProvider.tsx:41-47` + `Shell.tsx:25-27` + `lib/api.ts:15-17` make every
backend call from the browser with `Authorization: Bearer <token>`. Raw token available to any script
/ XSS.

### C6 — Non-atomic policy issuance
`services/pc-policy-svc/internal/application/command/bind_policy.go:59-68` saves quote and policy in
two separate transactions (each opens its own tx in `postgres_policy.go:23-35`). Comment: "save
transactionally (Here we just save sequentially for brevity)". A failed policy insert leaves an
orphaned BOUND quote; no outbox event emitted (line 70 is a comment).

### C7 — Outbox relay locking is a no-op
`services/pc-claims-svc/internal/infrastructure/outbox/relay.go:48-79` runs
`SELECT ... FOR UPDATE SKIP LOCKED` via `pool.Query` with **no explicit transaction**. The implicit tx
commits on `rows.Close()` (line 79) — before publishing — so locks release immediately and two workers
can publish the same events. The `UPDATE ... processed=TRUE` runs on a different pooled connection.

### C8 — Float money in the policy aggregate
`services/pc-policy-svc/internal/domain/policy/policy.go:47` stores `TotalPremium float64`; `NewPolicy`,
`Endorse`, `Cancel` all do float arithmetic. The rating engine returns `shopspring/decimal`, downcast
to float on every bind. Three money representations coexist repo-wide (decimal, float64, `money.ETB`).

---

## High

- **H1 — Committed TLS private key.** `certs/server.key` tracked in git (added `790b0f4`). Self-signed
  localhost cert so bounded impact, but must be scrubbed from history and rotated; generate at deploy
  time (K8s stack does via cert-manager).
- **H2 — Keycloak client secret committed.** `infra/keycloak/realm-medhen.json:26` —
  `"secret": "supersecret-keycloak-client"` for confidential client `pc-web`.
- **H3 — Wildcard CORS on live gateway.** `infra/pc-gateway/envoy.yaml:19-23` — `safe_regex: ".*"`
  reflecting `Authorization`, `max_age: 1728000`.
- **H4 — Wildcard redirect/web origins.** `realm-medhen.json:17-18,29` — `redirectUris: ["*"]` and
  `webOrigins: ["*"]` on public clients (open-redirect / code interception).
- **H5 — Hardcoded creds in manifests.** `infra/k8s/base/pc-iam-svc/deployment.yaml:32-37`
  (`admin123`, `postgres:postgres`), `infra/terraform/environments/dev/main.tf:28-29`,
  `deploy/helm/pc-notification-svc/values.yaml:14` (`medhen:medhen`).
- **H6 — Telebirr flow broken + unverified.** Webhook always passes `InvoiceID: nil`
  (`webhook/telebirr.go:62`) → handler returns "auto-allocation ... not implemented yet"
  (`process_payment_callback.go:114`); signature check only tests header non-empty (`telebirr.go:30-35`),
  no HMAC; `TenantID: "DEFAULT_TENANT"` hardcoded.
- **H7 — Idempotency SDK racy.** `libs/pc-idempotency-mgmt-sdk/idempotency.go:60-69` — non-atomic
  GET-then-SET (not SETNX/Lua), `Set(..., 0)` = never expires, returns 409 instead of cached response.
- **H8 — id token in logout URL.** `AuthProvider.tsx:63-70` →
  `web/app/api/auth/federated-logout/route.ts:5,14` puts the (unencoded) id token in a GET query
  string → browser history / logs / Referer; also CSRF-able.
- **H9 — Dead password-grant route.** `web/app/api/auth/login/route.ts:29` returns the full Keycloak
  token response (incl. `refresh_token`) to the browser; `lib/auth.ts:26` stores it in `localStorage`.
  Unused (import graph confirms) but live + unauthenticated + unrate-limited. Delete `lib/auth.ts` and
  `app/api/auth/login/`.
- **H10 — Impersonation fallback.** `web/lib/api.ts:18-20` sends `X-User-ID: demo-agent` when no token
  is present.

---

## Medium

- **M1 — Dockerfile can't build.** `Dockerfile:25-36` runs `apt-get` + shell `RUN echo` against a
  `distroless/static` base (no apt, no shell); `exec /app/bin/$SERVICE_NAME` runs an unvalidated
  env-named binary. Images tag-pinned but not digest-pinned.
- **M2 — Missing/fake premium logic.** Stamp duty never computed (only VAT 15% in
  `engine/pipeline.go:66-73`) despite spec; `policy.go:161-167` returns hardcoded 10%/5% cancellation
  refunds ignoring elapsed time.
- **M3 — Dummy KPIs.** `services/pc-reporting-svc/internal/infrastructure/clickhouse/query_repo.go:55-56`
  — `LossRatio = 0.65`, `CombinedRatio = 0.95` hardcoded.
- **M4 — README "in-memory fallback" false.** No `InMemory` repos exist; `pc-policy-svc` sets
  `pool = nil` and its own comment says repos "will panic on use". Services crash without Postgres.
- **M5 — Rating client stubbed.** `pc-policy-svc/internal/infrastructure/grpc_client/rating_client.go:34-52`
  returns hardcoded `500.00` as `float64`; never calls rating-calc.
- **M6 — Endorsement history corruption.** `policy.go:112` truncates prior version's `EffectiveTo` in
  memory, but `postgres_policy.go:53` inserts with `ON CONFLICT (id) DO NOTHING` → truncation never
  persisted → overlapping effective periods.
- **M7 — Frontend type/UX safety.** 84 `any`/unsafe casts (`lib/api.ts:97-121`, `staff/page.tsx:15-17`,
  etc.); `staff/page.tsx:228-230` reads `reserveMinor`/`recoveryMinor` unguarded → `NaN`; money ops via
  `prompt()`+`parseInt` (`staff/page.tsx:66-95`).
- **M8 — Info leak + missing headers.** `lib/api.ts:27-31` throws raw backend body rendered to users;
  `next.config.ts` sets no CSP/HSTS/X-Frame-Options.
- **M9 — K8s hardening gaps.** No `resources`/`securityContext` in `infra/k8s/deployment.yaml` &
  `pc-iam-svc/deployment.yaml`; `:latest` images with `imagePullPolicy: Always`.
- **M10 — Plaintext in transit.** Kafka `PLAINTEXT` (`docker-compose.yml:43-45`); `sslmode=disable`
  everywhere.
- **M11 — Double JSON-encoding.** `pc-policy-svc/internal/infrastructure/outbox/outbox.go:30` marshals
  an already-`[]byte` payload → base64 corruption (dead code; shared outbox is correct).
- **M12 — i18n race.** `libs/pc-shared-go-svc/i18n/i18n.go:12` documents thread-safety but `dict` has no
  mutex.
- **Postgres init** (`infra/postgres/init.sql`) is grant-only (`GRANT ALL ... TO medhen`), no tables/
  constraints/RLS, no least-privilege; schema-per-BC (`pc_party`, …) conflicts with `PG_SCHEMA=pc_medhen`
  used elsewhere.

---

## Low

- **L1 — Rating rounding inconsistency.** `engine/pipeline.go:68,73` — components kept at 6-dp internal
  precision but only gross rounded to currency; displayed parts can fail to sum.
- **L2 — Policy number not sequential.** `bind_policy.go:50` uses `EIC/MOT/2026/<8-hex-uuid>` with
  hardcoded year and no monotonic counter (spec: `EIC/MOT/{year}/{seq}`).
- **L3 — Silent audit drops.** `rating_service.go:52-80` fire-and-forget Kafka audit publish, error
  discarded (`_ =`).
- **L4 — `money.ETB` truncation.** `libs/pc-shared-go-svc/money/money.go:26,56` `FromFloat`/`Mul`
  truncate via `int64(x*100)` (used once / effectively dead; live path uses `shopspring/decimal`).
- Module-path inconsistency; `cmd/api` vs `cmd/server`; 3 services lack graceful shutdown; Go version
  drift (CI `1.22`, Dockerfile `1.26`, notification CI `1.25.x`); no SAST/dep/secret scanning in main CI.

---

## Testing

- ~18 test functions total; **6 of 15 services have zero tests** (incl. billing, underwriting, audit —
  which hold money logic).
- `pc-rating-calc-svc` is the only genuinely tested service (math/engine/models/handler).

---

## What's actually good

- Rating engine: `shopspring/decimal`, `RoundBank`, `SafeDiv`, internal-vs-final precision separation.
- Billing ledger (double-entry, UoW, overpayment→suspense) and claims settlement domain are well-built.
- **Parameterized SQL throughout — no SQL injection found** (incl. dynamic query builders,
  `search_policies.go:32`, `clickhouse/query_repo.go`).
- Real transactional outbox + relay in workflow/product-defn/audit/party-mgmt.
- No `dangerouslySetInnerHTML`/`eval` in frontend; `.env`/`.env.local` correctly gitignored (no
  populated secrets committed); CI has no secret exposure; `pc-notification-svc-ci.yml` runs Trivy.
- K8s stack has the right patterns (cert-manager TLS, JWT SecurityPolicy, explicit CORS allowlist,
  Coraza WAF) — they just aren't wired into the running path.

---

## Top priorities (before leaving an isolated sandbox)

1. **Fix authentication end-to-end** (C1–C5, H8, H10): enable gateway JWT, remove `X-Tenant-ID` /
   `mock-valid-token` / `X-User-ID` bypasses, switch SDK to RS256/JWKS, drop hardcoded secret
   fallbacks, gate `/staff` server-side, stop shipping tokens to the browser.
2. **Rotate & scrub** the committed private key (H1) and Keycloak client secret (H2); move creds to
   Secrets (H5).
3. **Make policy issuance atomic** + emit its outbox event (C6); eliminate `float64` money (C8); fix
   the outbox relay lock (C7).
4. **Implement or clearly mark** simulated inter-service calls (rating M5, Telebirr H6) so the demo
   isn't mistaken for a working pipeline.
</content>
