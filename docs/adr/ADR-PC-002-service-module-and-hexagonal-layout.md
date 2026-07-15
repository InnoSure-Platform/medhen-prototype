# ADR-PC-002: Service module & hexagonal layout

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Medhen is full multi-repo microservices from inception (Capability Doc §5.5): each Bounded Context is an independently deployable Go service in its own repo with its own database, integrating only via published ports and events.

## Decision

Each service repo is a Go module `github.com/InnoSure-Platform/<repo>` (repo names end `-svc`). Internal structure is hexagonal: `internal/{domain,application,api,infrastructure}`. Simple in-process services MAY use top-level `api/app/infra/domain` packages without `internal/`. The domain layer imports no other layer; all SQL/pgx/redis/kafka live in infrastructure.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
