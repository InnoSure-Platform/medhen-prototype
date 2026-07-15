# ADR-PC-017: Observability standard

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md)

## Context

Operating 20+ services requires uniform logs, metrics, and traces, and must never leak sensitive data.

## Decision

Services emit structured `slog` logs, OpenTelemetry traces, and metrics, exported to the `pc-infra` observability stack (Prometheus / Grafana / Jaeger, BC-MDH-19). No secrets, PII, or financial data appear in logs.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand with alternatives considered and full rationale before ratifying.
