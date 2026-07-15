# ADR-PC-007: Idempotency of mutating operations

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

At-least-once delivery and client retries mean mutating operations can be received more than once; duplicates must not double-charge, double-issue, or duplicate side effects.

## Decision

Every mutating command carries an idempotency key. Handlers are replay-safe: a replay returns the original result and starts no new transaction and repeats no side effects. Idempotency is realized by the Kernel (`pc-shared-go`, BC-MDH-20).

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
