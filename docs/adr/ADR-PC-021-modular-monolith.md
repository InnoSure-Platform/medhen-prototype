# ADR-PC-021: Modular monolith, sealed module boundaries, and the event/outbox backbone

- **Status:** Accepted (2026-07-20)
- **Applies to:** the whole platform (`cmd/medhen-api`, `internal/**`)
- **Supersedes:** ADR-PC-002 (repo-per-service hexagonal layout), ADR-PC-004
  (synchronous inter-service communication); amends ADR-PC-005 (Kafka is now
  optional at the edge, not the in-process backbone). ADR-PC-007 (idempotency),
  ADR-PC-008 (transactional outbox), ADR-PC-016 (audit) and ADR-PC-015
  (identity/tenancy) remain in force, realised inside the monolith.

## Context

The prototype began as a 15-service Go mesh (repo-per-service, gRPC between
services, Kafka/Avro mandatory, per-service databases). For an EIC Motor pilot
this imposed distributed-systems cost — network hops, partial failure, stubbed
cross-service calls (e.g. rating returning a hardcoded `500.00`), non-atomic
cross-service workflows (bind wrote policy and invoice in separate transactions),
and 20 `go.mod` files — without the scale that justifies it. The
[refactor plan](../refactor/modular-monolith-plan.md) consolidates to a **tier-1
modular monolith** while preserving a clean path to re-extract a module later.

## Decision

1. **Single deployable, single Go module.** One binary, `cmd/medhen-api`, is the
   composition root. Module path `github.com/InnoSure-Platform/medhen-prototype`.

2. **Sealed bounded-context modules.** Each context lives under
   `internal/modules/<bc>` as `domain / ports / app / adapters / rest / module.go`.
   A module may import only the shared platform kernel (`internal/platform/**`)
   and **other modules' `ports` packages** — never their `domain`, `app`, or
   `adapters`. This is enforced by `go-arch-lint` in CI (`.go-arch-lint.yml`).

3. **Cross-module calls are in-process ports.** The consumer defines the port it
   needs (e.g. `rating.Calculator`, `party.Reader`); the provider implements it;
   the composition root wires them. No gRPC between contexts.

4. **Events via a transactional outbox + in-process bus.** State-changing use
   cases write their aggregate and a domain event in **one** DB transaction
   (`database.WithinTx`); a relay publishes committed events to an in-process
   `eventbus`. Consumers (audit, billing, notification, reporting, document)
   subscribe. Because publish is a seam, a module can later be extracted by
   pointing the relay at Kafka — the handler code is unchanged (amends ADR-PC-005:
   Kafka becomes an edge/extraction concern, not a runtime dependency).

5. **One database, RLS tenant isolation.** A single Postgres with a versioned
   migration set; row-level security + a least-privilege app role enforce
   `tenant_id` isolation (ADR-PC-015). No cross-schema/cross-module FKs.

6. **Auth at the edge.** RS256/JWKS validation + server-side RBAC in one edge
   middleware; the tenant comes from the token, not a client header.

## Consequences

- **Simpler ops and correctness:** atomic issuance, real in-process rating, one
  image to deploy, one DB to run. Whole Motor lifecycle
  (quote→bind→issue→invoice→pay→FNOL→settle) runs and is audited end-to-end.
- **Boundaries stay honest** only because CI enforces them (arch-lint + the
  ports-only rule); without the gate the monolith would rot into a big ball of mud.
- **Extraction path preserved:** a module with a `ports` contract and outbox
  events can be lifted to its own service by swapping the relay target and adding
  a network adapter behind the same port.
- The older ADRs describing repo-per-service and synchronous inter-service calls
  are historical; this ADR is the current source of truth for structure.
