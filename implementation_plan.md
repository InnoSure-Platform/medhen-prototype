# Master Implementation Plan: Tier-0 Enterprise Platform

This master plan governs the end-to-end specification and implementation of the 18 target Bounded Contexts (excluding BC-MDH-14 FinCrime and BC-MDH-15 Commission). 

We are adopting a rigorous **Platform Engineering** philosophy: the platform core must be shared across all insurance products, while product-specific rules (like Motor) sit in their own implementations. Every service will be engineered to Tier-0/Tier-1 industry standards.

## User Review Required

> [!IMPORTANT]
> To execute this massive transformation, the AI Agent will operate in a **Service-by-Service** (Vertical Slice) progression. 
> For each Bounded Context, the agent will execute **Phase A (Specification)** followed by **Phase B (Code Implementation)** before moving to the next service.
> 
> Please review this strict engineering protocol. If approved, we will begin with **BC-MDH-02 (Product Definition Engine)**.

---

## The Standard Operating Procedure (SOP) per Service

For each of the 18 Bounded Contexts, the AI Agent SHALL follow this exact two-phase procedure:

### Phase A: Deep Domain Analysis & Specification (SSD)
Before writing a single line of code for a service, the AI Agent must conduct a deep domain analysis and author a comprehensive Software Specification Document (SSD). **There will be one SSD authored per microservice, operating under the strict mapping that 1 Bounded Context (BC) = 1 Microservice.**

1. **Format:** The SSD MUST strictly mirror the structure, rigor, and depth of the authoritative reference `docs/svc-14-case-management-spec-v2.md`.
2. **Sections Required:**
   - **Service Overview & Context Map:** Defining boundaries and producers/consumers.
   - **Technology Stack:** Strictly applying `docs/tech-stack-registry.md` (e.g., Go 1.26+, Postgres 18, `twmb/franz-go`, Chi, OpenTelemetry).
   - **Functional Requirements:** Normative `FR-` clauses mapped to PRD requirements.
   - **Domain Model (Tactical DDD):** Aggregates, Entities, Value Objects, pre/post conditions, invariants.
   - **Contracts:** Exhaustive REST/gRPC endpoints and Avro Event Schemas (Outbox/Kafka).
   - **BDD Scenarios:** Gherkin-style scenarios covering happy paths, edge cases, and degradation.
   - **Data Ownership:** PostgreSQL DDL with Outbox structure.
3. **Approval:** The AI Agent will pause and request user approval on the SSD artifact before proceeding to code.

### Phase B: Code Implementation (Clean Architecture)
Upon SSD approval, the AI Agent will implement the service using a hybrid architectural style:
1. **Horizontal / Hexagonal Domain Layer:** The core business logic (`domain/`) must be pure Go, completely isolated from infrastructure, containing Aggregates, Entities, and Domain Services.
2. **Vertical Slices for App/Infra:** Features (e.g., `features/createpolicy/`) will vertically slice through the Application (Command Handlers), API (HTTP/gRPC adapters), and Infrastructure (Postgres Repositories, Kafka Publishers) layers, ensuring high cohesion per use case.
3. **Event-Driven & Outbox:** Every state mutation must co-commit its domain data and an Outbox event (`platform.*` topics) in a single PostgreSQL transaction.
4. **Saga Pattern:** Distributed transactions (e.g., Quote → Underwriting → Billing → Bind) will be coordinated via Sagas, with strict compensation/rollback mechanisms.

---

## Execution Sequence (The Roadmap)

The AI Agent will apply the SOP above to the services in this precise order, establishing the foundational platform capabilities first before moving to complex orchestration.

### 1. Contract & Product Foundation
*Extracting the hardcoded Motor logic into true platform capabilities.*
1. **`BC-MDH-02` Product Definition Engine:** Dynamic versioned product schema registry.
2. **`BC-MDH-04` Rating & Premium Calculation:** Stateless premium calculation engine.
3. **`BC-MDH-05` Underwriting:** Standalone STP (Straight-Through Processing) rules engine.
4. **`BC-MDH-03` Policy Administration:** The Saga Orchestrator for the Quote-to-Bind journey.

### 2. Financial & Servicing Foundation
5. **`BC-MDH-07` Billing & Payments:** Idempotent invoicing and payment gateways.
6. **`BC-MDH-08` Document Management:** Async PDF generation (Schedules, COIs) via S3/MinIO.
7. **`BC-MDH-10` Notifications:** Async Kafka consumer for SMS/Email dispatch.
8. **`BC-MDH-09` Workflow & Approvals:** Maker-checker approval engine for referrals.

### 3. Claims & Advanced Capabilities
9. **`BC-MDH-06` Claims Management:** Async FNOL and settlement orchestration.
10. **`BC-MDH-12` Reinsurance & Coinsurance:** CQRS and async cessions.
11. **`BC-MDH-11` Reporting & Analytics:** CQRS data cubes and dashboards.
12. **`BC-MDH-13` Complaints & Disputes:** Complaint lifecycle management.

### 4. Governance & Shared Kernel
13. **`BC-MDH-01` Party & Customer Management:** Customer profiles and KYC events.
14. **`BC-MDH-16` Identity & Access Management:** RBAC management APIs.
15. **`BC-MDH-17` Audit & Compliance:** Immutable hash-chained audit ledger.
16. **`BC-MDH-18` Integration ACL:** Anti-corruption layer for Fayda/Telebirr.
17. **`BC-MDH-19` Observability:** Telemetry and SLO burn-rate configuration.
18. **`BC-MDH-20` Shared-Core Kernel:** Extracting the Outbox/Saga primitives into a reusable Go SDK.
