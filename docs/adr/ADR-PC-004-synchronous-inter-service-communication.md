# ADR-PC-004: Synchronous inter-service communication (gRPC / REST)

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Services have hard runtime (sync) dependencies (e.g. policy→party) and must expose stable, typed interfaces; the edge needs browser/mobile-friendly APIs.

## Decision

Service-to-service synchronous calls use gRPC. REST/JSON is exposed only at the edge via `pc-gateway`. All interface definitions live in `pc-contracts` (see ADR-PC-019).

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
