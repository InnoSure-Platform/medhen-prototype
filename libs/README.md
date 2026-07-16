# Medhen Platform Core Kernel (BC-MDH-20)

This directory contains the foundational, product-agnostic SDKs and primitives that form the Medhen Platform Kernel.

These libraries realize the specifications outlined in the **[Platform Kernel Specification](../docs/shared-go-spec-v1.md)**.

## Libraries

*   **`pc-auth-sdk`**: JWT validation and RBAC evaluation.
*   **`pc-contracts`**: Protobuf, gRPC, and Avro event schemas.
*   **`pc-idempotency-mgmt-sdk`**: Duplicate request detection and caching engine.
*   **`pc-telemetry-sdk`**: OpenTelemetry tracing, logging, and metrics instrumentation.

*See the [Service Registry](../docs/prd/service-registry.md) for context on how these libraries fit into the broader platform architecture.*
