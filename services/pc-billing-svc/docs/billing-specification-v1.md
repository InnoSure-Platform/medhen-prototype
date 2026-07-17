# svc-14: Billing & Payments Engine Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-14-v1 |
| **Service ID** | `svc-14` |
| **Service Name** | Billing & Payments Engine |
| **Bounded Context** | `BC-MDH-07` — Billing & Payments |
| **Version** | 1.0 |
| **Status** | Draft |
| **Date** | 2026-07-17 |
| **Classification** | Internal — Strictly Confidential (Financial Data) |
| **Tier** | Tier-0 (Mission Critical) |
| **Deploy Mode** | Microservice (`pc-billing-svc`) |
| **Target Repo** | `Platform Core/dev/pc-billing-svc` |
| **Phase** | Phase 1 (Core) |
| **PRD Anchor** | [Platform Core PRD](../../docs/prd/Medhen-Platform-PRD.md) (`REQ-BIL-*`) |
| **Capability Anchor** | [Capability Doc BC-MDH-07](../../docs/prd/Medhen-Platform-Capability-Document.md#bc-mdh-07--billing--payments-pc-billing-svc) |
| **Capabilities** | `CAP-BIL-001` to `CAP-BIL-A3` |
| **Methodologies** | DDD · Hexagonal · EDA · CQRS-lite · Transactional Outbox · Double-Entry Ledger |
| **Companion Specs** | `svc-13` Policy Management · `svc-18` Integration ACL |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-17 | Initial Tier-0 specification. Expanded to enterprise-grade depth, incorporating double-entry accounting constraints, MT940 reconciliation, rigorous BDDs, and strict Avro event schemas. |

---

## Document Structure Overview

1. **Service Overview**
2. **Technology Stack & Architecture**
3. **Comprehensive Functional Requirements**
4. **Domain Model & Events (Tactical DDD)**
5. **API Specifications (REST & Webhooks)**
6. **Event Schemas & Contracts (Avro)**
7. **Behaviour-Driven Scenarios (BDD)**
8. **Data Ownership & Persistence (PostgreSQL DDL)**
9. **Integration & Dependency Contracts**
10. **Non-Functional Requirements & SLOs**
11. **Observability & Audit Specification**
12. **Operational Runbooks**
13. **Engineering Definition of Done (DoD)**

---

## 1. Service Overview

### 1.1 Mission Statement

`svc-14` Billing & Payments (`BC-MDH-07`) is the **immutable, zero-loss transactional financial core** of the Medhen Platform. It abstracts all monetary movements away from insurance-specific domains (like Policy or Claims). It acts as a sub-ledger system employing **double-entry accounting principles** to track billing accounts, invoices, credit/debit notes, exact payment allocations, installment schedules, and refunds.

As a **Tier-0** component, this service must ensure absolute ACID compliance. It is the definitive source of truth determining whether a policy is "in force" due to paid premiums, or "cancelled" due to overdue balances. Furthermore, it natively integrates with external gateways (Telebirr, CBE Birr) via `svc-18` and bridges the gap to external enterprise ERPs (like the EIC General Ledger) via automated journal entry synchronization to satisfy IFRS 17 and local regulatory auditing.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem** | Traditional core insurance platforms entangle policy lifecycle logic with payment processing. This leads to reconciliation nightmares, untraceable manual journal entries, an inability to easily adopt new mobile money rails (e.g., Telebirr), and systemic audit risks. |
| **Value** | By isolating financial capabilities into a dedicated Tier-0 Bounded Context, Medhen ensures total ledger integrity. Complex scenarios—such as overpayments, partial payments across multiple invoices, short-rate cancellation refunds, and bulk bank statement reconciliations—are handled autonomously and transparently. |
| **Stakeholders** | Chief Financial Officer (CFO), Treasury, Revenue Accounting, External Auditors, NBE Regulators, Policyholders. |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Priority |
|:---|:---|:---|
| `CAP-BIL-001` | **Billing Account Management:** Auto-creation on bind, statement generation, aggregated multi-policy balances. | P0 |
| `CAP-BIL-002` | **Invoice & Note Management:** Itemized invoices, tax/VAT calculation, Credit/Debit notes for mid-term adjustments. | P0 |
| `CAP-BIL-003` | **Payment Processing:** Multi-rail processing (Telebirr, CBE, Bank, Cash, Check), with strict webhook idempotency. | P0 |
| `CAP-BIL-004` | **Installment Billing:** Configurable payment plans, down payments, dynamic recalculation on endorsements, late fees. | P0 |
| `CAP-BIL-005` | **Refunds Management:** Pro-rata / short-rate return premium calculations, maker-checker approval workflows. | P1 |
| `CAP-BIL-006` | **Reconciliation:** Auto-matching, MT940 / CSV bank statement ingestion, Suspense Account (Unallocated funds) management. | P1 |
| `CAP-BIL-007` | **ERP GL Integration:** Daily batching of journal entries into external corporate ERP systems. | P1 |
| `CAP-BIL-A1` | **Auto-Debit (Phase 3):** Tokenized recurring collections via wallet/card mandates. | P3 |
| `CAP-BIL-A2` | **Dunning Intelligence (Phase 4):** ML-optimized reminder cadence based on payer behaviour. | P4 |

---

## 2. Technology Stack & Architecture

### 2.0 Architectural Narrative

`svc-14` implements a **CQRS-lite architecture backed by a Relational Double-Entry Ledger**. 
Because financial consistency is non-negotiable, all domain mutations occur within a highly restrictive Unit of Work (UoW) orchestrated over PostgreSQL. We enforce idempotency at the network edge using Redis, meaning a duplicated webhook from a payment gateway will be rejected in milliseconds before touching the database.

To broadcast state (e.g., "Invoice Paid") to the Policy Engine (`svc-13`), we employ the **Transactional Outbox Pattern**. The outbox guarantees that the financial state mutation and the event emission to Kafka are committed atomically.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / Runtime | **Go 1.26.x** | High concurrency for webhook storms; exact decimal arithmetic (via specialized decimal libs, avoiding float64). |
| Primary Store | **PostgreSQL 18.x** | Indispensable ACID compliance. Extensive use of `NUMERIC(15,2)` constraints. |
| Event Backbone | **Kafka** + **Avro** | Guaranteed exactly-once delivery semantics for financial events. |
| Edge Idempotency | **Redis** | Sub-millisecond duplicate request rejection. |
| Bank Statement Parsing| **Custom MT940 Parser** | Native parsing of SWIFT MT940 standard for daily bank reconciliation. |

---

## 3. Comprehensive Functional Requirements

Requirements are classified according to RFC 2119.

### 3.1 Billing Accounts & Invoicing (`FR-BIL-INV`)
- **FR-BIL-INV-01 (Account Auto-Creation):** The service SHALL autonomously provision a `BillingAccount` when consuming a `pc.policy.bound.v1` event, keyed by the `customer_id`.
- **FR-BIL-INV-02 (Itemized Invoices):** The service SHALL generate an `Invoice` containing distinct `InvoiceLineItems` for Base Premium, Riders, Policy Fees, and VAT. 
- **FR-BIL-INV-03 (Credit/Debit Notes):** Upon consuming a `pc.policy.endorsed.v1` event, the service SHALL generate a `DebitNote` for additional premiums owed, or a `CreditNote` for return premiums.
- **FR-BIL-INV-04 (Taxation):** The service SHALL compute local VAT (e.g., 15%) and stamp taxes distinctly on the invoice line items, preparing them for statutory reporting.
- **FR-BIL-INV-05 (Statements):** The service SHALL expose an API to generate a `Statement of Account` containing all invoices, payments, and the net running balance for a given date range.

### 3.2 Payment Processing & Allocation (`FR-BIL-PAY`)
- **FR-BIL-PAY-01 (Omnichannel Initiation):** The service SHALL expose unified APIs to initiate payments via `svc-18 Integration ACL`, generating a unique `payment_reference` for gateway tracking.
- **FR-BIL-PAY-02 (Strict Idempotency):** Webhook endpoints MUST validate idempotency using the gateway's `transaction_id`. Duplicate callbacks SHALL return HTTP 200 without mutating the ledger.
- **FR-BIL-PAY-03 (Allocation Strategy):** When a payment is received without a specific `invoice_id`, the service SHALL auto-allocate funds based on the FIFO principle (oldest overdue invoice first).
- **FR-BIL-PAY-04 (Overpayments & Suspense):** If a payment exceeds the outstanding balance, the excess MUST be recorded as `Unallocated` and credited to a Suspense Liability Account tied to the `BillingAccount`.
- **FR-BIL-PAY-05 (Partial Payments):** The service SHALL permit partial payments, transitioning the invoice status to `PARTIALLY_PAID` and calculating the new `outstanding_amount`.

### 3.3 Installments & Dunning (`FR-BIL-INS`)
- **FR-BIL-INS-01 (Schedule Generation):** If an installment plan is selected, the service SHALL partition the total invoice into an `InstallmentSchedule` with distinct `due_dates` and `amounts`.
- **FR-BIL-INS-02 (Down Payment Enforcement):** The service SHALL enforce that Policy inception is blocked (via events) until the specific `DownPayment` installment is fully paid.
- **FR-BIL-INS-03 (Endorsement Recalculation):** If a mid-term endorsement increases the premium, the `InstallmentSchedule` SHALL dynamically recalculate, distributing the delta across the remaining future installments.
- **FR-BIL-INS-04 (Dunning Lifecycle):** The service SHALL track aging (Days Past Due) and emit `pc.billing.payment.overdue.v1` events at configured intervals (e.g., D+1, D+7, D+14) to trigger notifications or cancellation.

### 3.4 Refunds & Maker-Checker (`FR-BIL-REF`)
- **FR-BIL-REF-01 (Refund Calculation):** On policy cancellation (`pc.policy.cancelled.v1`), the service SHALL calculate the unearned premium. Depending on the cancellation reason, it applies either *Pro-Rata* or *Short-Rate* penalty scales.
- **FR-BIL-REF-02 (Approval Workflow):** Refunds exceeding an automated threshold (e.g., 5,000 ETB) MUST enter a `PENDING_APPROVAL` state, requiring a `FinanceAdmin` role to authorise the gateway disbursement via `svc-25` Workflow.

### 3.5 Bank Reconciliation (`FR-BIL-REC`)
- **FR-BIL-REC-01 (MT940 Ingestion):** The service SHALL support the uploading of MT940 formatted bank statements for batch processing.
- **FR-BIL-REC-02 (Auto-Matching Rules):** The reconciliation engine SHALL attempt to match bank credits to internal payments using exact amount, date proximity (+/- 1 day), and reference number extraction (fuzzy logic).
- **FR-BIL-REC-03 (Exception Queue):** Unmatched bank statement lines SHALL be routed to a manual reconciliation queue for Finance Operations to investigate.

### 3.6 ERP Journal Synchronization (`FR-BIL-ERP`)
- **FR-BIL-ERP-01 (Double-Entry Core):** Every financial action MUST generate balanced journal entries (Debits = Credits) across the internal Chart of Accounts (e.g., Dr Accounts Receivable, Cr Premium Revenue, Cr VAT Payable).
- **FR-BIL-ERP-02 (Batch Sync):** The service SHALL expose a cron-driven batch process to aggregate daily journal entries and push them to the external ERP (EIC GL) via the Integration ACL.

### 3.7 Negative Requirements
- **FR-BIL-NEG-01:** Financial records (`Invoices`, `Payments`, `JournalEntries`) SHALL NEVER be subjected to `DELETE` operations or `UPDATE` operations that alter monetary values. Corrections strictly require offsetting Credit/Debit Notes.
- **FR-BIL-NEG-02:** The system MUST NOT utilize floating-point data types (`float32`, `float64`) for any monetary representation, to avoid rounding errors. All math must utilize precise decimal libraries.

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Aggregate Roots

The domain relies on strict boundaries to ensure the UoW guarantees ACID updates across related entities.

| Aggregate Root | Responsibilities | Invariants | Domain Events Emitted |
|:---|:---|:---|:---|
| **`BillingAccount`** | Root entity for a customer's financial state. Aggregates the total outstanding balance and holds unallocated/suspense funds. | Balance >= 0 (unless in credit). | `BillingAccountCreated` |
| **`Invoice`** | Represents a formal request for payment. Composed of `InvoiceLineItem` entities. | `total_amount` must equal the sum of line items. Cannot change after `ISSUED`. | `InvoiceIssued`, `InvoiceSettled`, `InvoiceOverdue` |
| **`Payment`** | Captures funds received from a gateway or bank. Composed of `PaymentAllocation` entities pointing to invoices. | `total_amount` = `allocated_amount` + `unallocated_amount`. | `PaymentReceived`, `PaymentFailed` |
| **`InstallmentSchedule`**| Orchestrates the timeline of due dates for an Invoice. | Sum of `Installment` amounts MUST exactly equal the linked `Invoice.total_amount`. | `ScheduleGenerated`, `ScheduleRecalculated` |
| **`Refund`** | Governs the return of funds to a customer. | Cannot exceed the total historical `amount_paid` for the policy. | `RefundStaged`, `RefundProcessed` |
| **`LedgerTransaction`** | The Double-Entry source of truth. Contains multiple `JournalEntry` lines. | Sum(Debits) MUST exactly equal Sum(Credits). | `LedgerTransactionPosted` |

### 4.2 Chart of Accounts (Internal Representation)
`svc-14` maps insurance actions to an internal Chart of Accounts before syncing to the external ERP.
- `1000` - Accounts Receivable (A/R)
- `2000` - Premium Revenue
- `2100` - VAT Payable
- `2200` - Suspense / Unallocated Funds
- `1100` - Cash / Bank Clearing

### 4.3 Command Catalog & Unit of Work (UoW)

| Command | Aggregate Target | Operation | Example Post-Condition & Ledger Impact |
|:---|:---|:---|:---|
| `IssueInvoice` | `Invoice`, `LedgerTransaction` | Creates an invoice and records the revenue expectation. | Dr 1000 A/R<br>Cr 2000 Premium Rev<br>Cr 2100 VAT |
| `ProcessPaymentCallback`| `Payment`, `Invoice`, `BillingAccount` | Records funds, allocates to Invoice, updates Account balance. | Dr 1100 Cash Clearing<br>Cr 1000 A/R |
| `ProcessOverpayment` | `Payment`, `BillingAccount` | Payment > Invoice. Allocates remainder to Account Suspense. | Dr 1100 Cash Clearing<br>Cr 1000 A/R (Invoice amt)<br>Cr 2200 Suspense (Remainder) |
| `ApplySuspenseToInvoice`| `BillingAccount`, `Invoice` | Uses previously unallocated funds to pay a newly generated invoice. | Dr 2200 Suspense<br>Cr 1000 A/R |

---

## 5. API Specifications (REST & Webhooks)

Base path: `/api/pc-billing/v1`
Authentication: OAuth2 + RBAC.

### 5.1 Account & Invoice APIs

| Method | Endpoint | Description | Idempotency | Auth Role |
|:---|:---|:---|:---|:---|
| `GET` | `/accounts/{account_id}` | Retrieve comprehensive account balances. | — | `billing.read` |
| `GET` | `/accounts/{account_id}/statement`| Generate structured statement (JSON/PDF trigger). | — | `billing.read` |
| `GET` | `/invoices/{invoice_id}` | Fetch invoice and line items. | — | `billing.read` |

### 5.2 Payment APIs

| Method | Endpoint | Description | Idempotency | Auth Role |
|:---|:---|:---|:---|:---|
| `POST` | `/payments/initiate` | Generates a gateway checkout URL for the user. | Required | `billing.write` |
| `POST` | `/payments/cash` | Manual cashier operation to record cash receipt. | Required | `billing.cashier` |
| `POST` | `/payments/allocate` | Manually allocate suspense funds to an invoice. | Required | `billing.admin` |

#### Payload: `POST /payments/initiate`
```json
{
  "invoice_id": "inv-f8a7d2",
  "method": "TELEBIRR",
  "amount": 12500.00,
  "currency": "ETB",
  "success_return_url": "https://portal.medhen.com/payment/success",
  "metadata": {
    "device_ip": "197.156.90.1",
    "channel": "MOBILE_APP"
  }
}
```

### 5.3 Webhooks (Ingress)

| Method | Endpoint | Description | Auth |
|:---|:---|:---|:---|
| `POST` | `/webhooks/telebirr/callback` | Asynchronous confirmation of payment success/failure. | Signature |
| `POST` | `/webhooks/cbe/callback` | CBE Bank API callback. | Signature |

#### Idempotency Contract for Webhooks
1. Validate HMAC signature in headers.
2. Extract provider `transaction_id`.
3. Perform Redis `SETNX idempotency:telebirr:{transaction_id}`.
4. If exists, return `200 OK` instantly to acknowledge receipt without reprocessing.
5. If new, initiate UoW to process payment.

### 5.4 Error Taxonomy (RFC 7807)

| Error Code | HTTP | Message | Description / Client Action |
|:---|:---|:---|:---|
| `BIL-1001` | `404` | `AccountNotFound` | The specified Billing Account does not exist. |
| `BIL-1002` | `400` | `InvoiceAlreadySettled` | Cannot allocate funds to an invoice with zero balance. |
| `BIL-1003` | `422` | `InvalidAllocationAmount` | The requested allocation exceeds the payment amount. |
| `BIL-1004` | `401` | `InvalidWebhookSignature` | Cryptographic signature of webhook payload failed. |
| `BIL-2001` | `409` | `ConcurrentLedgerUpdate` | Optimistic locking failure. Retry the transaction. |

---

## 6. Event Schemas & Contracts (Avro)

`svc-14` is a vital publisher to the enterprise bus. Downstream services rely on these events to trigger policy activation, document generation, and external ERP syncs.

### 6.1 Topic Mapping

| Topic | Producer | Partition Key | Schema | Purpose |
|:---|:---|:---|:---|:---|
| `pc.billing.invoice.issued.v1` | `svc-14` | `tenant_id:invoice_id` | `InvoiceIssuedEvent` | Triggers Document Service to generate PDF. |
| `pc.billing.payment.received.v1` | `svc-14` | `tenant_id:policy_id` | `PaymentReceivedEvent` | Triggers Policy Service to transition `BOUND` -> `ACTIVE`. |
| `pc.billing.payment.overdue.v1` | `svc-14` | `tenant_id:policy_id` | `PaymentOverdueEvent` | Triggers Dunning / Notification Service. |
| `pc.billing.ledger.posted.v1` | `svc-14` | `tenant_id:ledger_tx_id`| `LedgerTransactionEvent` | Async sink for the Data Warehouse / EIC ERP. |

### 6.2 Avro Schema: `PaymentReceivedEvent`

```json
{
  "namespace": "medhen.platform.billing.v1",
  "type": "record",
  "name": "PaymentReceivedEvent",
  "doc": "Emitted when funds are successfully cleared and allocated.",
  "fields": [
    {"name": "event_id", "type": "string", "logicalType": "uuid"},
    {"name": "tenant_id", "type": "string"},
    {"name": "policy_id", "type": "string"},
    {"name": "billing_account_id", "type": "string"},
    {"name": "payment_id", "type": "string"},
    {"name": "gateway_reference", "type": "string"},
    {"name": "amount_cleared", "type": "double"},
    {"name": "currency", "type": "string", "default": "ETB"},
    {"name": "allocations", "type": {
      "type": "array",
      "items": {
        "type": "record",
        "name": "Allocation",
        "fields": [
          {"name": "invoice_id", "type": "string"},
          {"name": "amount_applied", "type": "double"}
        ]
      }
    }},
    {"name": "unallocated_amount", "type": "double"},
    {"name": "occurred_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
  ]
}
```

---

## 7. Behaviour-Driven Scenarios (BDD)

These scenarios form the Acceptance Criteria for the QA Engineering team.

### 7.1 Complex Payment Allocations & Overpayments
**Scenario: BIL-BDD-01 | Webhook triggers Overpayment to Suspense**
* **Given** an `Invoice` `inv-001` with an outstanding balance of `5,000 ETB`
* **And** a `BillingAccount` `acc-999` with `0.00 ETB` suspense balance
* **When** a Telebirr webhook confirms a payment of `6,000 ETB` for `inv-001`
* **Then** the service allocates `5,000 ETB` to `inv-001`
* **And** transitions `inv-001` status to `PAID`
* **And** allocates the remaining `1,000 ETB` to the `BillingAccount` suspense ledger
* **And** emits a `PaymentReceivedEvent` indicating `unallocated_amount: 1000`
* **And** generates a Double-Entry ledger transaction balancing the `1000 ETB` to account `2200`.

### 7.2 Installment Recalculation on Endorsement
**Scenario: BIL-BDD-02 | Mid-Term Endorsement increases premium**
* **Given** a Policy with an active `InstallmentSchedule` (Quarterly)
* **And** Installment 1 (Q1) is `PAID` (`3,000 ETB`)
* **And** Installments 2, 3, 4 are `PENDING` (`3,000 ETB` each)
* **When** a `pc.policy.endorsed.v1` event arrives with an additional premium of `1,500 ETB`
* **Then** a `DebitNote` is issued for `1,500 ETB`
* **And** the `InstallmentSchedule` is recalculated
* **And** Installments 2, 3, 4 are updated to `3,500 ETB` each (spreading the 1,500 over the 3 remaining periods).

### 7.3 Bank Reconciliation Edge Case
**Scenario: BIL-BDD-03 | MT940 Ingestion with Fuzzy Match**
* **Given** a `Payment` initiated via Bank Transfer for `15,000 ETB` with reference `REF-ABC-123`
* **When** an MT940 file is processed containing a credit of `15,000 ETB` and narrative text `TRANSFER FROM JOHN DOE REFABC123`
* **Then** the reconciliation engine uses fuzzy alphanumeric matching on the narrative
* **And** successfully links the bank statement line to the internal `Payment`
* **And** transitions the payment from `PENDING` to `SUCCESS`.

---

## 8. Data Ownership & Persistence (PostgreSQL DDL)

Strict typing and constraints are the backbone of `svc-14`.

### 8.1 Schema Implementation (Double-Entry Focus)

```sql
-- 1. Core Billing Structures
CREATE TABLE billing_accounts (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    customer_id UUID NOT NULL,
    suspense_balance NUMERIC(15, 2) DEFAULT 0.00 CHECK (suspense_balance >= 0),
    version INT NOT NULL DEFAULT 1, -- Optimistic locking
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, customer_id)
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    billing_account_id UUID REFERENCES billing_accounts(id),
    policy_id UUID NOT NULL,
    invoice_type VARCHAR(20) NOT NULL, -- NEW_BUSINESS, RENEWAL, DEBIT_NOTE, CREDIT_NOTE
    total_amount NUMERIC(15, 2) NOT NULL,
    amount_paid NUMERIC(15, 2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL, -- DRAFT, ISSUED, PARTIALLY_PAID, PAID, CANCELLED
    due_date TIMESTAMPTZ NOT NULL,
    CONSTRAINT check_amounts CHECK (amount_paid <= total_amount AND amount_paid >= 0)
);

CREATE TABLE invoice_line_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id UUID REFERENCES invoices(id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    tax_amount NUMERIC(15, 2) DEFAULT 0.00,
    gl_account_code VARCHAR(20) NOT NULL -- Maps to the Chart of Accounts
);

-- 2. Payment & Allocations
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    gateway_transaction_id VARCHAR(100),
    method VARCHAR(50) NOT NULL,
    total_amount NUMERIC(15, 2) NOT NULL CHECK (total_amount > 0),
    unallocated_amount NUMERIC(15, 2) NOT NULL,
    status VARCHAR(20) NOT NULL, -- PENDING, SUCCESS, FAILED
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (gateway_transaction_id)
);

CREATE TABLE payment_allocations (
    id BIGSERIAL PRIMARY KEY,
    payment_id UUID REFERENCES payments(id),
    invoice_id UUID REFERENCES invoices(id),
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- 3. Double Entry Ledger (Immutable Audit)
CREATE TABLE ledger_transactions (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    reference_id UUID NOT NULL, -- Ties to an Invoice or Payment ID
    reference_type VARCHAR(50) NOT NULL,
    posted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE journal_entries (
    id BIGSERIAL PRIMARY KEY,
    ledger_transaction_id UUID REFERENCES ledger_transactions(id),
    account_code VARCHAR(20) NOT NULL,
    debit_amount NUMERIC(15, 2) DEFAULT 0.00,
    credit_amount NUMERIC(15, 2) DEFAULT 0.00,
    CONSTRAINT single_side CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR 
        (credit_amount > 0 AND debit_amount = 0)
    )
);
```

### 8.2 Security & Data Privacy
- Payments containing PAN (Primary Account Numbers) are tokenised by `svc-18 Integration ACL`. `svc-14` ONLY stores secure reference tokens.
- DB level Encryption-at-Rest is strictly enforced for `pc_billing_db`.

---

## 9. Integration & Dependency Contracts

| Service | Protocol | Contract | Fallback Strategy |
|:---|:---|:---|:---|
| **`pc-policy-svc`** | Kafka (Async Inbound) | Consumes `policy.bound.v1`. | Kafka offsets handle temporary downtime seamlessly. |
| **`svc-18` ACL** | gRPC (Sync Outbound) | Payment Gateway Initiation. | If ACL is down, UI displays "Payment Gateway Unavailable"; retries via exponential backoff. |
| **`pc-document-mgmt-svc`** | Kafka (Async Outbound)| Consumes `invoice.issued.v1` to generate PDF. | Async generation; PDF link becomes available in UI once generated. |
| **External ERP** | SFTP / API via ACL | Daily GL Journal Sync. | Retains batch state; alerts Ops if sync fails for > 24 hours. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | Target SLO | Consequence of Breach | Measurement / APM |
|:---|:---|:---|:---|
| **Availability (Core APIs)** | 99.99% | Absolute blocker for revenue collection. | Prometheus: `sum(success) / sum(total)` |
| **Webhook Processing Latency** | P95 < 200ms | Gateways may timeout and retry aggressively, causing network storms. | OpenTelemetry span duration. |
| **Idempotency Accuracy** | 100% | Double-crediting accounts results in direct financial loss. | Automated Reconciliation Job drift reports. |
| **Database Transactions** | > 1000 TPS | Required for batch processing of large corporate payroll deductions. | Postgres `xact_commit` metrics. |

---

## 11. Observability & Audit Specification

The service utilizes the standard `pc-telemetry-sdk`.

### 11.1 Golden Signals & Financial Metrics (Prometheus)
- `billing_revenue_collected_total{method="telebirr", currency="ETB"}`
- `billing_refunds_processed_total{reason="cancellation"}`
- `billing_ledger_imbalance_count` (Must trigger P1 PagerDuty alert if > 0).

### 11.2 Audit Logging
All financial mutations generate an immutable log entry containing:
- `actor_id` (System, User, or Webhook Agent)
- `ip_address`
- `before_state` / `after_state` (JSON diff)
- `trace_id`

---

## 12. Operational Runbooks

### 12.1 Ledger Imbalance Alert (P1)
**Symptom:** PagerDuty alert `LedgerImbalanceDetected` fires. A UoW committed a `LedgerTransaction` where Debits != Credits.
**Action:**
1. Halt batch jobs.
2. Execute diagnostic script: `SELECT * FROM vw_unbalanced_journals WHERE date = CURRENT_DATE`.
3. Escalate to L3 Engineering to identify the precision/logic bug in Go decimal allocation code.
4. Issue manual correction scripts verified by the CFO.

### 12.2 Payment Gateway Webhook Storm
**Symptom:** Telebirr retries thousands of pending transactions simultaneously.
**Action:**
1. Ensure Redis cluster (handling Idempotency) memory and CPU are stable.
2. If PostgreSQL connection pool is saturated, scale the `pc-billing-svc` replicas via HPA up to maximum bounds. The system is designed to handle this via idempotency locks.

---

## 13. Engineering Definition of Done (DoD)

Prior to merging to `main` for `staging` deployment:
1. **Decimal Precision:** Codebase MUST NOT contain the strings `float32` or `float64` in any domain struct involving money. Uses `github.com/shopspring/decimal`.
2. **Double-Entry Tests:** Integration tests MUST assert that for every simulated BDD scenario, `SUM(debits) = SUM(credits)`.
3. **Idempotency Suite:** Concurrent load tests (100 parallel requests with the same `transaction_id`) MUST yield exactly 1 success and 99 cached 200s, with only 1 database INSERT.
4. **Outbox Chaos:** Chaos Mesh tests must verify that pod termination during a UoW does not result in orphaned DB state without a corresponding Kafka message.
