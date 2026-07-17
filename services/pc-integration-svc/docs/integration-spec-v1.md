# svc-18: Integration & Anti-Corruption Layer Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-18-v1 |
| **Service ID** | `svc-18` |
| **Service Name** | Integration Service (INT) |
| **Bounded Context** | `BC-MDH-18` — Integration & Anti-Corruption Layer |
| **Version** | 1.0 |
| **Status** | Draft |
| **Date** | 2026-07-16 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-0 |
| **Deploy Mode** | Microservice (`pc-integration-svc`) |
| **Target Repo** | `Platform Core/dev/services/pc-integration-svc` |
| **Phase** | Phase 1 (Core MVP) |
| **PRD Anchor** | [Platform Core PRD](../../../../../../docs/prd/Medhen-Platform-PRD.md) (`REQ-INT-*`) |
| **Methodologies** | DDD · Hexagonal · Anti-Corruption Layer (ACL) · CQRS · Circuit Breaker · Webhook Ingestion |
| **Companion Specs** | `svc-06` Billing · `svc-01` Party Mgmt · `svc-09` Notification |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-16 | Initial Tier-0 specification covering Sections 1-13. Drafted for standardizing integrations across Telebirr, CBE, Amole, ERP, Fayda, and SMS gateways. |

---

## Document Structure Overview

1. **Service Overview**
2. **Technology Stack & Architecture**
3. **Functional Requirements & State Machines**
4. **Domain Model & Events (Tactical DDD)**
5. **API Specifications (REST & gRPC)**
6. **Event Schemas & Contracts (Avro)**
7. **Behaviour-Driven Scenarios (BDD)**
8. **Data Ownership & Persistence Schema**
9. **Integration & Dependency Contracts**
10. **Non-Functional Requirements & SLOs**
11. **Observability Specification**
12. **Operational Runbooks**
13. **Engineering Definition of Done (DoD)**

---

## 1. Service Overview

### 1.1 Mission Statement

`svc-18` Integration Service (`BC-MDH-18`) operates as the universal **Anti-Corruption Layer (ACL)** and edge gateway for all third-party external system communications in the Medhen Platform. It isolates internal domain microservices (like Billing, Claims, and Party) from the volatile, legacy, inconsistent, or proprietary APIs of external partners.

By acting as a standardized translation layer, the Integration Service ensures that the core platform's ubiquitous language is never polluted by external domain models. It normalizes operations for **Payment Gateways (Telebirr, CBE, Amole)**, **National ID Verification (Fayda)**, **Accounting (ERP)**, and **Communications (SMS/Email Gateways)**. 

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem Space** | External systems often have brittle APIs, poor uptime, varied authentication mechanisms (SOAP, XML-RPC, REST, custom crypto), and unpredictable latency. Allowing core services like Billing to talk directly to a bank gateway creates massive systemic risk and tight coupling. |
| **Value Delivered** | Provides a highly resilient, standardized interface for internal systems. Implements circuit breakers, bulkheads, and automated retries. Normalizes webhook ingestion from external partners (e.g., async payment success callbacks) into clean, internal Kafka events. Enables hot-swapping of third-party vendors without touching core business logic. |
| **Stakeholders** | Finance (Reconciliation), Compliance (KYC/Fayda), Platform Engineering, External Vendor Tech Teams. |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Primary REQ | Phase |
|:---|:---|:---|:---|
| `CAP-INT-001` | **Payment Gateway Abstraction:** Standardizes `InitiatePayment` and `RefundPayment` operations across Telebirr, CBE, Amole, and standard card networks. Handles partner-specific signature generation and payload encryption. | `REQ-INT-001` - `004` | 1 |
| `CAP-INT-002` | **Webhook Ingestion Engine:** Exposes public, secure endpoints to receive asynchronous callbacks (e.g., Telebirr payment confirmation). Verifies signatures, enforces idempotency, and translates to internal Kafka events (`PaymentSettledEvent`). | `REQ-INT-010` - `015` | 1 |
| `CAP-INT-003` | **Fayda ID Proxy:** Acts as a synchronous ACL for the National ID (Fayda) system. Handles biometric template transmission, token rotation, and normalizes responses into the internal `KYCProfile` format. | `REQ-INT-020` - `023` | 1 |
| `CAP-INT-004` | **Omnichannel Dispatch:** Translates internal `pc.notification.send` commands into provider-specific SMS API calls (e.g., Ethio Telecom) or SMTP commands, handling rate limiting and failovers. | `REQ-INT-030` - `032` | 1 |
| `CAP-INT-005` | **ERP Ledger Sync:** Batches internal financial events (premium paid, claim settled) and pushes them as General Ledger (GL) entries to the corporate ERP via scheduled sagas. | `REQ-INT-040` - `042` | 2 |

### 1.4 Bounded Context Responsibilities (`BC-MDH-18`)

| Owns | Exposes | Produces (via Outbox/Kafka) | Invariants |
|:---|:---|:---|:---|
| Integration state & credentials | Internal gRPC (Translation API) | `pc.integration.payment_success.v1` | Never leak external partner data models to internal services. |
| Third-party rate limits | External Webhooks (REST) | `pc.integration.payment_failed.v1` | All external webhooks must be verified cryptographically. |
| Idempotency keys & retries | Prometheus metrics (Partner health)| `pc.integration.fayda_verified.v1` | Failed downstream HTTP calls must not crash the upstream caller. |

### 1.5 Context Map

```mermaid
flowchart TB
    subgraph External["External Partners"]
        TB["Telebirr API"]
        CBE["CBE API"]
        FAY["Fayda National ID"]
        SMS["Ethio Telecom SMS"]
        ERP["Corporate ERP"]
    end

    subgraph Core["svc-18 Integration Service (ACL)"]
        PAY_ACL["Payment Gateway ACL"]
        ID_ACL["Identity ACL (Fayda)"]
        NOT_ACL["Notification Dispatcher"]
        ERP_ACL["ERP Sync Engine"]
        WEBHOOK["Webhook Receiver"]
        
        WEBHOOK --> PAY_ACL
    end

    subgraph Internal["Internal Platform Services"]
        BIL["pc-billing-svc (BC-06)"]
        PTY["pc-party-mgmt-svc (BC-01)"]
        NOT["pc-notification-svc (BC-09)"]
        FIN["pc-fincrime-svc (BC-14)"]
    end

    BIL -->|Initiate Payment (gRPC)| PAY_ACL
    PAY_ACL -->|Format & Encrypt| TB
    PAY_ACL -->|Format & Encrypt| CBE
    
    TB -->|Callback/Webhook| WEBHOOK
    CBE -->|Callback/Webhook| WEBHOOK
    
    PTY -->|Verify ID (gRPC)| ID_ACL
    ID_ACL -->|Secure Transport| FAY
    
    NOT -->|Send SMS (gRPC)| NOT_ACL
    NOT_ACL -->|SMPP / HTTP| SMS
    
    BIL -->|Emit Ledger Events (Kafka)| ERP_ACL
    ERP_ACL -->|Batch API| ERP
```

---

## 2. Technology Stack & Architecture

### 2.0 Operations-Plane Architecture Narrative

`svc-18` must be the most resilient component in the architecture because it interfaces with systems over which Medhen has zero control. 

To guarantee resilience, the architecture utilizes:
1. **Circuit Breakers & Bulkheads (Resilience4j / Go equivalent):** If the CBE gateway experiences an outage, the circuit breaker opens, immediately failing fast to the internal `pc-billing-svc` rather than tying up threads waiting for network timeouts.
2. **Idempotency Engine:** Webhooks from payment providers are notoriously unreliable (duplicate deliveries, out-of-order delivery). `svc-18` uses an atomic PostgreSQL table to guarantee exactly-once processing of incoming vendor transaction IDs.
3. **Temporal (or Asynq) for Workflow Execution:** Long-running, asynchronous partner integrations (like ERP batching or retry backoffs for SMS) are orchestrated using durable workflow engines to survive pod restarts.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / runtime | **Go 1.25.x** | Perfect for concurrent network I/O; excellent HTTP/gRPC multiplexing. |
| API — external | **REST/JSON & XML** | Webhook ingress for third-party systems. |
| API — internal | **gRPC / Protobuf** | Sub-millisecond latency for internal core services. |
| Primary Store (State) | **PostgreSQL 18.x** | Idempotency keys, webhook payload audit logs, API key rotation state. |
| Resiliency Lib | **sony/gobreaker** | Implementation of Circuit Breaker, Retries, and Timeouts. |
| Event Backbone | **Kafka (KRaft) + franz-go** | Emitting translated domain events (e.g. `PaymentSuccess`). |
| Secret Management | **OpenBao 2.x** | Dynamic retrieval of vendor API keys, certificates, and MTLS keys. |
| Cache & Rate Limits | **Valkey 8.x** | Enforcing outbound rate limits to prevent vendor throttling (e.g. max 50 SMS/sec). |

### 2.2 Core Configuration References

| Config Key | Default | Purpose |
|:---|:---|:---|
| `circuit.telebirr.error_threshold` | `50%` | Error percentage within the rolling window to trip the circuit breaker. |
| `circuit.telebirr.timeout_ms` | `5000` | Max wait time before abandoning an outbound HTTP request. |
| `webhook.idempotency_ttl_days`| `30` | Duration to retain processed external transaction IDs. |
| `rate_limit.sms.tps` | `100` | Token bucket capacity for outbound SMS dispatch. |

---

## 3. Functional Requirements & State Machines

### 3.1 Detailed Requirement Catalog (RFC 2119)

#### 3.1.1 Payment Abstraction (`FR-INT-PAY-*`)
- **FR-INT-PAY-1 — Unified Initiation:** The service SHALL expose an internal gRPC `InitiatePayment` RPC. The payload MUST include `amount`, `currency`, `provider_enum`, and `callback_reference`.
- **FR-INT-PAY-2 — Vendor Translation:** Upon receiving a payment request, the ACL MUST map the request to the specific vendor's schema (e.g., XML for legacy banks, JSON for Telebirr), inject the correct credentials, and perform MTLS or payload signing.
- **FR-INT-PAY-3 — Circuit Breaking:** The service SHALL wrap all outbound provider calls in a Circuit Breaker. If OPEN, it MUST immediately return an `Unavailable` error to internal callers.

#### 3.1.2 Webhook Ingestion (`FR-INT-WEB-*`)
- **FR-INT-WEB-1 — Signature Verification:** External endpoints (`/v1/webhooks/telebirr`, `/v1/webhooks/cbe`) SHALL strictly verify incoming payload signatures (HMAC-SHA256, RSA) using the provider's public key.
- **FR-INT-WEB-2 — Idempotency:** The service MUST hash the incoming payload or extract the vendor `transaction_id` and execute a `pg_try_advisory_xact_lock` (or equivalent `INSERT ... ON CONFLICT`) to prevent duplicate processing.
- **FR-INT-WEB-3 — Translation & Emission:** Validated webhooks SHALL be translated into an internal `pc.integration.payment_settled.v1` Avro event and dispatched via the Transactional Outbox.

#### 3.1.3 Fayda Identity ACL (`FR-INT-ID-*`)
- **FR-INT-ID-1 — Secure Proxy:** The service SHALL expose `VerifyNationalId` to `pc-party-mgmt-svc`. It will proxy the request to the Fayda system, managing the required OAuth2 client credentials and session tokens.
- **FR-INT-ID-2 — Error Normalization:** Fayda specific errors (e.g., `E-5001 Biometric Mismatch`) MUST be translated into standard internal gRPC status codes and localized error enumerations.

### 3.2 State Machine Definition (Payment Intent Tracking)

Even though Billing owns the invoice, the Integration service must track the lifecycle of the network transaction to handle retries and webhook correlation.

| Current State | Trigger | Target State | Action / Consequence |
|:---|:---|:---|:---|
| `—` | Internal `InitiatePayment` | `INITIATED` | Store internal reference ID. |
| `INITIATED` | Provider 200 OK | `PENDING_PARTNER`| Await async webhook callback. |
| `PENDING_PARTNER`| Webhook (Success) | `SUCCESS` | Emit `payment_settled` Kafka event. |
| `PENDING_PARTNER`| Webhook (Failed) | `FAILED` | Emit `payment_failed` Kafka event. |
| `PENDING_PARTNER`| Timeout (7 days) | `EXPIRED` | Auto-fail orphaned transactions. |

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Bounded Context Boundary (`BC-MDH-18`)

The Integration Domain is devoid of deep business logic. It does not know what an "Insurance Policy" or a "Premium" is. It only understands "Payment Intents", "Notification Payloads", and "Identity Verification Requests".

### 4.2 Aggregate Roots

| Aggregate Root | Definition & Invariants | Emitted Events |
|:---|:---|:---|
| **`IntegrationTransaction`** | Represents a distinct interaction with an external system. Used primarily for auditing, non-repudiation, and webhook correlation.<br><br>**Invariants:**<br>• Must record raw request/response payloads (encrypted) for compliance. | `ProviderCallbackReceived`<br>`IntegrationTransactionFailed` |
| **`WebhookReceipt`** | Represents a delivered webhook from a vendor.<br><br>**Invariants:**<br>• Vendor transaction ID must be unique per vendor. | `TranslatedEventDispatched` |

### 4.3 Command Catalog (Internal)

| Command | Aggregate | Pre-conditions | Post-conditions (Success) | Domain Exception |
|:---|:---|:---|:---|:---|
| `InitiatePayment` | `IntegrationTransaction` | Valid provider enum. | Request sent, transaction `PENDING_PARTNER`. | `ProviderUnavailable`, `ConfigurationMissing` |
| `IngestWebhook` | `WebhookReceipt` | Valid signature, novel transaction ID. | Translated event dropped in Outbox. | `InvalidSignature`, `DuplicateWebhook` |
| `SendSMS` | `IntegrationTransaction` | Rate limit bucket has capacity. | Message dispatched to SMPP/HTTP. | `RateLimitExceeded`, `ProviderRejected` |

---

## 5. API Specifications (REST & gRPC)

### 5.1 External Webhook API (REST)

**Base path:** `/api/pc-integration/v1/webhooks`

| Method | Endpoint | Purpose | Security |
|:---|:---|:---|:---|
| `POST` | `/telebirr/callback` | Receive async payment status | IP Whitelist + RSA Signature |
| `POST` | `/cbe/notification` | Receive CBE direct deposit status | IP Whitelist + HMAC |
| `POST` | `/fayda/status` | Receive async biometric validation | MTLS |

### 5.2 Internal gRPC API

**Service Definition:** `medhen.platform.integration.v1.IntegrationGatewayService`

| RPC | Request | Response | SLA (P95) |
|:---|:---|:---|:---|
| `InitiatePayment` | `PaymentRequest` (Amount, Provider) | `PaymentResponse` (Checkout URL / Txn ID) | < 50ms |
| `VerifyNationalId`| `FaydaVerificationReq` | `FaydaVerificationRes` | < 200ms (Proxy) |
| `DispatchMessage` | `MessageDispatchReq` (Channel, Body)| `DispatchAck` | < 10ms (Queued) |

### 5.3 Error Taxonomy (RFC 7807 Problem Details)

| Domain Exception | HTTP/gRPC Code | Error Code | Client Action / Resolution |
|:---|:---|:---|:---|
| `ProviderUnavailable` | `503 Unavailable` | `INT-1001` | Circuit breaker is open. Internal services must fallback (e.g. offer different payment method). |
| `InvalidSignature` | `401 Unauthorized` | `INT-1002` | Reject webhook. Security alert triggered. |
| `DuplicateWebhook` | `200 OK` (Idempotent)| `INT-1003` | Ignore payload, return 200 to satisfy vendor retry mechanism. |

---

## 6. Event Schemas & Contracts (Avro)

`svc-18` acts as an event producer, translating raw vendor webhooks into clean internal domain events.

### 6.1 Topic Mapping

| Event | Topic | Partition Key | Schema ID |
|:---|:---|:---|:---|
| `PaymentSettledEvent` | `platform.integration.payment.settled.v1` | `internal_reference_id` | `PaymentSettledEvent` |
| `PaymentFailedEvent` | `platform.integration.payment.failed.v1` | `internal_reference_id` | `PaymentFailedEvent` |

### 6.2 Avro Schema: `PaymentSettledEvent`

```json
{
  "namespace": "medhen.platform.integration.v1",
  "type": "record",
  "name": "PaymentSettledEvent",
  "fields": [
    {"name": "event_id", "type": "string", "logicalType": "uuid"},
    {"name": "internal_reference_id", "type": "string", "doc": "Matches the ID sent during InitiatePayment"},
    {"name": "provider", "type": "string", "doc": "e.g., TELEBIRR, CBE"},
    {"name": "provider_transaction_id", "type": "string", "doc": "The vendor's network ID"},
    {"name": "amount_settled", "type": "double"},
    {"name": "currency", "type": "string"},
    {"name": "settled_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
  ]
}
```
*Impact:* `pc-billing-svc` consumes this event to mark an Invoice as `PAID` and subsequently trigger policy binding.

---

## 7. Behaviour-Driven Scenarios (BDD)

### 7.1 Webhook Idempotency

**Scenario: INT-BDD-01 | Handle duplicate Telebirr callbacks**
* **Given** a Telebirr webhook payload with `transaction_id = "TB-999888"` is successfully processed
* **When** Telebirr sends the exact same payload 5 seconds later due to a network stutter
* **Then** the Integration Service detects the duplicate in the `webhook_receipts` table
* **And** no new Kafka event is emitted
* **And** the service responds with HTTP 200 OK to acknowledge receipt.

### 7.2 Circuit Breaker Activation

**Scenario: INT-BDD-02 | CBE Gateway Degraded**
* **Given** the CBE API is experiencing 500 Server Errors
* **When** `pc-billing-svc` attempts 10 `InitiatePayment` calls within 10 seconds
* **Then** the first 5 calls fail with `ProviderUnavailable` and increment the error threshold
* **And** the circuit breaker transitions to `OPEN`
* **And** the 6th through 10th calls fail *immediately* without attempting a network connection
* **And** a Prometheus alert `CircuitBreakerOpen{provider="CBE"}` is triggered.

---

## 8. Data Ownership & Persistence Schema

### 8.1 PostgreSQL Transactional Schema (Simplified)

```sql
-- Webhook Idempotency & Audit Log
CREATE TABLE webhook_receipts (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    provider_transaction_id VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- PROCESSED, IGNORED_DUPLICATE, FAILED_SIGNATURE
    raw_payload_encrypted BYTEA, -- Retained for audit/disputes
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_transaction_id)
);

-- Transaction Ledger (Initiations)
CREATE TABLE integration_transactions (
    internal_reference_id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- PAYMENT, REFUND, ID_VERIFY
    amount NUMERIC(15, 2),
    currency VARCHAR(3),
    state VARCHAR(20) NOT NULL, -- INITIATED, PENDING_PARTNER, SUCCESS, FAILED
    circuit_breaker_status VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## 9. Integration & Dependency Contracts

| External Service | Contract / Protocol | Coupling | Resilience / Fallback |
|:---|:---|:---|:---|
| **Telebirr Gateway** | REST / JSON | Sync (Initiate) / Async (Webhook) | Circuit Breaker. Graceful degradation to other payment methods. |
| **CBE API** | SOAP / XML | Sync / Async | Strict timeouts (SOAP can hang indefinitely). |
| **Ethio Telecom SMS** | SMPP / REST | Async via queue | Token bucket rate limiting to prevent blocking. Retries via Temporal. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | SLO | Consequence of Breach | Measurement / Triggers |
|:---|:---|:---|:---|
| **Availability (gRPC)** | 99.99% | Internal services cannot process payments or verify IDs. | Prometheus `grpc_server_handled_total`. Alert if < 99.99%. |
| **Webhook Processing** | P95 < 200ms | Vendor retry storms if webhooks timeout. | OpenTelemetry REST Spans on `/webhooks/*`. |
| **Idempotency Guarantee** | 100% | Double-crediting invoices, leading to massive financial loss. | Audited via daily reconciliation scripts. |

---

## 11. Observability Specification

### 11.1 Golden Signals (Prometheus)

- **Traffic:** `webhook_ingest_total{provider="telebirr"}`
- **Errors:** `provider_http_requests_total{provider="fayda", status=~"5.."}`
- **Circuit Breakers:** `circuit_breaker_state{name="cbe_payment"}` (0=Closed, 1=Open, 2=Half-Open)

### 11.2 Custom Domain Metrics

- `payment_initiation_total{provider="telebirr", status="success|failed"}`
- `webhook_signature_failures_total{provider="telebirr"}` (Crucial security metric)

---

## 12. Operational Runbooks

### 12.1 Circuit Breaker Remains OPEN
**Symptom**: `CircuitBreakerOpen{provider="CBE"}` alert fires for > 15 minutes.
**Action**: 
1. Check CBE partner status dashboard.
2. If CBE is confirmed healthy but circuit is stuck, force transition to half-open:
   `curl -X POST localhost:8080/debug/circuitbreaker/cbe/half-open`
3. Monitor `provider_http_requests_total` to ensure traffic flows cleanly.

### 12.2 Webhook Signature Rotation
**Symptom**: Telebirr informs Medhen of an emergency RSA key rotation. Webhooks start failing signature validation (`INT-1002`).
**Action**:
1. Retrieve new public key from vendor.
2. Update OpenBao: `bao kv put secret/medhen/integration/telebirr public_key="..."`
3. Trigger service config hot-reload: `kubectl rollout restart deploy/pc-integration-svc`

---

## 13. Engineering Definition of Done (DoD)

Before `svc-18` can be deployed to the `staging` environment for Phase 1, the following quality gates MUST be passed:

1. **Test Coverage**: Idempotency logic, signature verification, and circuit breaker transitions MUST have 100% unit test coverage.
2. **Chaos Engineering**: Must inject network faults (TCP drops, 500s, 30s delays) into external partner mocks to prove the Circuit Breaker correctly opens and protects internal callers.
3. **Security Audits**: The webhook validation algorithms must be explicitly reviewed by AppSec. Webhook endpoints must be immune to replay attacks.
4. **Idempotency Proof**: Load test pushing 1,000 identical webhooks concurrently must result in exactly 1 Kafka event emitted.
