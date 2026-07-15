# Enhancement Plan: Tier-0 Motor Platform

This implementation plan outlines the engineering effort required to elevate the current Motor prototype into a **Tier-0, future-proof, enterprise-grade platform** that perfectly aligns with the Medhen Capability Document.

Per your instruction, we are **halting expansion to new product lines (like Life)**. Instead, we will aggressively harden the **Motor Insurance** vertical to prove the premium architectural foundations (EDA, Zero Trust, InnoGuard Fincrime, Immutable Audit).

## Goal Description
Transform the existing monolithic Motor prototype into a distributed, event-driven, and highly secure microservices platform.

## User Review Required
> [!IMPORTANT]
> **Architectural Shift:** This plan involves dismantling the `medhen-api` monolith into separate running Go binaries, requiring us to introduce Apache Kafka (for events) and Keycloak (for IAM) into the local docker-compose environment.
> **InnoGuard Integration:** We will stub the integration for the InnoGuard Financial Crime platform. Please confirm if you have a specific OpenAPI spec for InnoGuard or if we should define the `pc-fincrime-svc` boundary contract ourselves.

## Proposed Changes

### 1. True Microservice Decomposition (The Shared Core)
We will extract the `usecase.Motor` monolith into physically isolated bounded contexts communicating via gRPC.
#### [NEW] `platform/cmd/pc-party-mgmt-svc/main.go`
#### [NEW] `platform/cmd/pc-policy-svc/main.go`
#### [NEW] `platform/cmd/pc-billing-svc/main.go`
#### [NEW] `platform/cmd/pc-claims-svc/main.go`
- **Impact:** Ensures the platform is truly modular. The "Motor" specific components will be expressed strictly as configuration injected into these core services.

### 2. Event-Driven Architecture (EDA) & Outbox Pattern
We will replace synchronous cross-domain calls with asynchronous Kafka events.
#### [MODIFY] `platform/internal/store/postgres.go`
- Add an `outbox_events` table to the schema creation logic.
- Implement a transactional outbox writer so domain events (e.g., `pc.policy.bound`) are saved in the same transaction as the state change.
#### [NEW] `platform/internal/events/relay.go`
- Implement an outbox relay to read from the outbox table and publish to Kafka.

### 3. Governance, Audit, & Fincrime (InnoGuard)
We will introduce the critical Tier-0 governance modules.
#### [NEW] `platform/cmd/pc-audit-svc/main.go`
- Subscribes to all domain events and constructs an immutable, hash-chained audit trail.
#### [NEW] `platform/cmd/pc-fincrime-svc/main.go`
- Exposes a gRPC hook invoked by `pc-party-mgmt-svc` during onboarding to perform Sanctions/PEP screening via the external **InnoGuard** platform.
- Exposes a hook for `pc-claims-svc` to score fraud probability on FNOL.

### 4. Zero-Trust Security & IAM
#### [MODIFY] `platform/internal/httphandlers/api.go`
- Integrate OIDC middleware to validate JWT tokens issued by Keycloak.
- Enforce strict RBAC (Role-Based Access Control) on all endpoints instead of mocking the tenant context.

### 5. OpenTelemetry & Observability
#### [MODIFY] `platform/internal/svcboot/boot.go`
- Inject OpenTelemetry (OTel) instrumentation.
- Configure Jaeger exporter for distributed tracing across the new microservices.

## Verification Plan

### Automated Tests
- Run integration tests validating that a policy bound in `pc-policy-svc` successfully drops an event into Kafka, which is consumed by `pc-billing-svc` to generate an invoice.
- Verify cryptographic hashing in the `pc-audit-svc` storage.

### Manual Verification
- Spin up the new distributed `docker-compose` environment.
- Execute the Quote-to-Bind Motor flow in the UI.
- Verify in Jaeger that traces span across the Gateway -> Policy -> Rating -> Fincrime services.
- Verify in the InnoGuard stub logs that screening was requested.
