# ADR-PC-015: Identity, authorization & tenancy

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

The platform is multi-tenant and product-agnostic; access control and tenant isolation must be uniform and fail-safe.

## Decision

AuthN is OIDC via Keycloak fronting `pc-iam-svc` (BC-MDH-16). Tokens are validated at `pc-gateway` and re-checked per service. Every query, cache key, idempotency key, and event is scoped by `tenant_id`; a missing tenant predicate returns not-found, never cross-tenant data. Authorization decisions fail closed (deny).

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
