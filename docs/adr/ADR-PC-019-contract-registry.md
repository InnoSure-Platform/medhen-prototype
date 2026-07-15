# ADR-PC-019: Contract registry: pc-contracts

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · `pc-contracts/README.md`

## Context

Cross-service schemas duplicated in-tree drift apart and break integration; there must be a single source of truth for every contract technology.

## Decision

All cross-service contracts live only in `pc-contracts`, by kind:

- **gRPC service contracts → Protobuf** — `proto/<domain>/<name>.proto` (package `pc.<domain>.v1`).
- **Kafka event schemas → Avro** — `avro/<domain>/<subject>.v{n}.avsc` (subjects `pc.{domain}.{event}.v{n}`), plus a `topics/topics.yaml` producer/consumer matrix (ADR-PC-005).
- **Edge REST → OpenAPI** — `openapi/<domain>.v{n}.yaml`.

Services import generated Go stubs from `github.com/InnoSure-Platform/pc-contracts/gen/go/...` (Protobuf) and the generated Avro types; they never hand-copy or in-tree a cross-service schema (enforced by `scripts/contract-hygiene.sh`, which flags stray `.proto` and `.avsc`). Backward-compatibility is enforced by CI for both proto and Avro.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold; reviews and gates check conformance. This is a stub capturing the platform's stated convention — expand before ratifying.
