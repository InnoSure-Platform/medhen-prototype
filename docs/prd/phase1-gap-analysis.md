# Medhen Platform: Phase 1 Gap Analysis

This document evaluates the current state of the Medhen Prototype (Phase 1) against the enterprise-grade, Tier-0 requirements defined in the **PRD** and **Capability Document (MDH-CAP-001)**. 

While the current Phase 1 implementation successfully demonstrates the "happy path" Motor workflow (KYC, Quote, Authority Matrix, Billing, Claims), it falls short of the stringent architectural, security, and governance standards required for a production-ready, tier-0 global platform.

## 1. Architectural Gaps

| Capability | Target (Capability Document) | Current Implementation | Gap Severity |
| :--- | :--- | :--- | :--- |
| **Microservice Isolation** | Full multi-repo/multi-process microservices communicating via gRPC (sync) and Kafka (async). | The platform runs as a **monolith** (`medhen-api`) using a shared `usecase.Motor` struct and `api.go` router. | **CRITICAL** |
| **Event-Driven Architecture (EDA)** | Domain events (`pc.policy.bound`, etc.) published via the **Outbox Pattern** to Kafka for resilient asynchronous processing. | Synchronous in-memory function calls; no Kafka; no Outbox pattern. | **CRITICAL** |
| **CQRS** | Read-optimized projections for Customer-360 and Reporting (`BC-MDH-11`). | Standard relational queries against the write database. | **HIGH** |
| **Data Plane / Control Plane Split** | Products versioned immutably via `pc-product-defn-svc` and loaded at runtime. | Hardcoded product definition in the monolith state. | **HIGH** |

## 2. Security & Governance Gaps

| Capability | Target (Capability Document) | Current Implementation | Gap Severity |
| :--- | :--- | :--- | :--- |
| **Zero-Trust & mTLS** | Strict mTLS between all microservices. | N/A (Monolith). | **HIGH** |
| **Identity & Access (IAM)** | OAuth2 / OIDC via Keycloak with Role-Based and Attribute-Based Access Control (RBAC/ABAC). | Mocked tenant parsing (`httpx.TenantFromRequest`). | **CRITICAL** |
| **PII Encryption** | AES-256 encryption at rest; masked logs. | Plaintext storage of national IDs and personal data. | **HIGH** |

## 3. Observability & Audit Gaps

| Capability | Target (Capability Document) | Current Implementation | Gap Severity |
| :--- | :--- | :--- | :--- |
| **Distributed Tracing** | OpenTelemetry (OTel) traces across all services (Jaeger/Tempo). | None. | **HIGH** |
| **Immutable Audit** | Cryptographically hash-chained, append-only audit log for regulatory defense (`BC-MDH-17`). | Simple `queryAudit` relational table lookup. | **CRITICAL** |
| **Bi-Temporal History** | Policy states reconstructable in both system-time and business-time. | Simple relational updates (data loss on overwrite). | **HIGH** |

## 4. Operational & Domain Gaps

| Capability | Target (Capability Document) | Current Implementation | Gap Severity |
| :--- | :--- | :--- | :--- |
| **Financial Crime (`BC-MDH-14`)** | Sanctions, PEP screening, and fraud detection integrated with **InnoGuard**. | Completely absent. | **CRITICAL** |
| **Reinsurance (`BC-MDH-12`)** | Treaty configuration and automated cessions. | Completely absent. | **HIGH** |
| **Reporting (`BC-MDH-11`)** | NBE statutory returns, executive dashboards. | A simple `/demo/kpis` endpoint. | **HIGH** |

---

## Conclusion
The current prototype is an excellent **Phase 0 Pilot MVP**. However, to meet the Tier-0 enterprise standards for Phase 1 (Production Motor), the monolith must be physically decomposed into the architecture specified by the Capability Document, and the governance domains (Security, Fincrime, Audit) must be instantiated as first-class services.
