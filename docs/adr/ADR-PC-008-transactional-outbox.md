# ADR-PC-008: Transactional outbox

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

A state change and the events announcing it must be consistent; a dual-write to DB and Kafka can lose or duplicate events on failure.

## Decision

Domain state, child rows, and emitted events are persisted in one database transaction via an outbox table. A relay publishes outbox rows to Kafka at-least-once. No service writes to Kafka outside the outbox.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
