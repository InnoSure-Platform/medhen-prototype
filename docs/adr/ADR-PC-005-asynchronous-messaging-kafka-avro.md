# ADR-PC-005: Asynchronous messaging: Kafka + Avro

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Most cross-service integration is event-driven and must be decoupled and replayable; event payloads need a single schema definition, a run-time registry, and enforced backward compatibility. Synchronous request/response is a separate concern (Protobuf/gRPC — ADR-PC-004).

## Decision

The platform uses **both contract technologies, by transport**:

- **Synchronous** gRPC service contracts are **Protobuf** (ADR-PC-004, ADR-PC-019).
- **Asynchronous** Kafka event payloads are **Avro** (`.avsc`), defined in `pc-contracts` under `avro/<domain>/`, published to topics `pc.{domain}.{event}.v{n}` with `tenant_id` as the partition key.

Avro subjects are registered in a **run-time schema registry** at compatibility **BACKWARD**; evolution is additive-only (new fields carry defaults). A design-time CI gate (`pc-contracts`) and the run-time registry both enforce compatibility.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered (e.g. Protobuf-over-Kafka) and full rationale before ratifying.
