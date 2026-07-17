# ADR-PC-017: Observability standard and Control Plane

- **Status:** Approved
- **Date:** 2026-07-16
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md) · [Observability SSD](../../services/pc-observability-svc/docs/observability-spec-v1.md)

## Context

Operating 20+ services requires uniform logs, metrics, and traces, and must never leak sensitive data. Furthermore, as the platform scales to support multiple Lines of Business (LOB) and multi-tenancy, relying on static GitOps files (YAML) for Prometheus recording rules, SLOs, and Alertmanager routing becomes unmanageable and prone to "alert fatigue".

## Decision

1. **Telemetry Standard**: Services emit structured `slog` logs, OpenTelemetry traces, and metrics, exported to the `pc-infra` observability stack (Mimir / Loki / Tempo, BC-MDH-19). No secrets, PII, or financial data appear in logs.
2. **Active Control Plane**: We introduce `pc-observability-svc` (BC-MDH-21) as an active Tier-1 Control Plane. It exposes a programmatic REST/gRPC API for defining SLOs (Service Level Objectives) and provisioning tenants.
3. **Control Plane / Data Plane Separation**: 
   - *Control Plane*: `pc-observability-svc` backed by Postgres, managing the desired state of SLOs and alert routes.
   - *Data Plane*: The underlying Grafana Mimir/Loki/Tempo stack. The Control Plane translates SLOs into Prometheus recording rules and pushes them to the Mimir Ruler API dynamically.
4. **Telemetry Gateway**: To support authenticated frontend/mobile tracing, `pc-observability-svc` acts as an OTLP-compatible gateway, validating JWTs via `pc-iam-svc` before forwarding enriched telemetry to the internal OTel Collector.

## Consequences

- Product teams can define "SLO-as-Code" via API without modifying core infrastructure repositories.
- The platform achieves automated tenant isolation for metrics and logs.
- The architecture requires `pc-observability-svc` to be highly available (Tier-1) because, while its failure won't stop telemetry ingestion, it will halt new tenant provisioning and dynamic SLO updates.
- Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance.
