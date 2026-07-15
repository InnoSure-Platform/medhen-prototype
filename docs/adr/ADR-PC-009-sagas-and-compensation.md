# ADR-PC-009: Sagas & compensation for cross-service workflows

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Business flows span services (e.g. policy issue ↔ billing, claims ↔ billing) and cannot use distributed two-phase commit.

## Decision

Cross-service workflows are choreographed sagas coordinated by events, each step defining an explicit compensating action. Failures resolve via compensation, never partial commit; degradation is fail-safe (deny/roll back), never fail-open.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
