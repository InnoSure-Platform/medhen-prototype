# ADR-PC-016: Immutable audit trail

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Insurance is regulated; who-did-what-when on financial and policy actions must be reconstructable and tamper-evident.

## Decision

Every state-changing action emits an immutable audit event to `pc-audit-svc` (BC-MDH-17) through the outbox (ADR-PC-008). The audit store is append-only; audit emission is part of the Definition of Done for any mutating handler.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
