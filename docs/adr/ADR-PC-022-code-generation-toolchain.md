# ADR-PC-022: Code generation toolchain

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-10
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · `pc-contracts`

## Context

gRPC stubs and event types must be generated consistently across every repo from the contracts in `pc-contracts` — for both contract technologies (ADR-PC-019).

## Decision

- **Protobuf → `buf`** generates Go gRPC service stubs from `proto/` (with `buf lint` + `buf breaking`).
- **Avro → an Avro codegen step** (e.g. `hamba/avro` or `gogen-avro`) generates Go event types from `avro/*.avsc`, with an Avro BACKWARD-compat check against the registered prior subject.

Both outputs land under `pc-contracts/gen/go/...`, are published as a versioned module, and are **never hand-edited** (treated as read-only, excluded from contract-hygiene). Services consume them via `go get`; they do not run codegen locally for cross-service contracts.

## Consequences

Enforced by the quality system (IS-AQG-001) and the shared test/CI scaffold. This is a stub capturing the platform's stated convention — expand before ratifying.
