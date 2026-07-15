# Platform Core — Architecture Decision Records

Cross-cutting architecture decisions for the Medhen Platform Core (`pc-*`) services.
IDs are stable (`ADR-PC-0NN`); cited across the quality system (IS-AQG-001) and the
shared test/CI scaffold. Entries below are **stubs pending Architect ratification** —
they capture the platform's stated conventions from the
[Service Registry](../prd/service-registry.md) and
[Capability Document](../prd/Medhen-Platform-Capability-Document.md); expand each with
alternatives and full rationale before marking Accepted.

| ADR | Title |
|-----|-------|
| [ADR-PC-002](./ADR-PC-002-service-module-and-hexagonal-layout.md) | Service module & hexagonal layout |
| [ADR-PC-004](./ADR-PC-004-synchronous-inter-service-communication.md) | Synchronous inter-service communication (gRPC / REST) |
| [ADR-PC-005](./ADR-PC-005-asynchronous-messaging-kafka-avro.md) | Asynchronous messaging: Kafka + Avro |
| [ADR-PC-007](./ADR-PC-007-idempotency-of-mutating-operations.md) | Idempotency of mutating operations |
| [ADR-PC-008](./ADR-PC-008-transactional-outbox.md) | Transactional outbox |
| [ADR-PC-009](./ADR-PC-009-sagas-and-compensation.md) | Sagas & compensation for cross-service workflows |
| [ADR-PC-015](./ADR-PC-015-identity-authorization-and-tenancy.md) | Identity, authorization & tenancy |
| [ADR-PC-016](./ADR-PC-016-audit-trail.md) | Immutable audit trail |
| [ADR-PC-017](./ADR-PC-017-observability.md) | Observability standard |
| [ADR-PC-019](./ADR-PC-019-contract-registry.md) | Contract registry: pc-contracts |
| [ADR-PC-020](./ADR-PC-020-integration-environment.md) | Shared integration environment |
| [ADR-PC-021](./ADR-PC-021-cicd-image-build-and-publish.md) | CI/CD & image build/publish |
| [ADR-PC-022](./ADR-PC-022-code-generation-toolchain.md) | Code generation toolchain |
| [ADR-PC-023](./ADR-PC-023-continuous-deployment-gitops.md) | Continuous deployment via GitOps |
| [ADR-PC-024](./ADR-PC-024-agentic-testing-strategy.md) | Agentic testing strategy — agent roles per test type |

> Numbering follows the IDs already referenced by the quality docs (inherited from the
> platform standard); gaps (001, 003, 006, …) are reserved for ADRs not yet cited.
