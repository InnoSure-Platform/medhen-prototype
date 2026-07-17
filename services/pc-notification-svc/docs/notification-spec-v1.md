# svc-16: Notification Engine Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-16-v1 |
| **Service ID** | `svc-16` |
| **Service Name** | Notification Engine |
| **Bounded Context** | `BC-MDH-10` — Notifications |
| **Version** | 1.0 |
| **Status** | Draft |
| **Date** | 2026-07-17 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-2 |
| **Deploy Mode** | Microservice (`pc-notification-svc`) |
| **Target Repo** | `Platform Core/dev/pc-notification-svc` |
| **Phase** | Phase 1 |
| **PRD Anchor** | [Platform Core PRD](../prd/Medhen-Platform-PRD.md) |
| **Capability Anchor** | [Capability Doc BC-MDH-10](../../docs/prd/Medhen-Platform-Capability-Document.md#bc-mdh-10--notifications-pc-notification-svc) |
| **Capabilities** | `CAP-NOT-001` to `CAP-NOT-A2` |
| **Methodologies** | Event-Driven Architecture (EDA) · Hexagonal · Outbox / Inbox · Saga |
| **Companion Specs** | `svc-18` Integration ACL · `svc-01` Party Management |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-17 | Initial Tier-2 specification covering complete notification lifecycle, routing, and delivery tracking. |

---

## Document Structure Overview

1. **Service Overview**
2. **Technology Stack**
3. **Functional Requirements**
4. **Domain Model & Events (Tactical DDD)**
5. **API Specifications**
6. **Event Schemas & Contracts (Avro)**
7. **Behaviour-Driven Scenarios (BDD)**
8. **Data Ownership & Persistence**
9. **Integration & Dependency Contracts**
10. **Non-Functional Requirements & SLOs**
11. **Observability Specification**
12. **Operational Runbooks**
13. **Engineering Definition of Done (DoD)**

---

## 1. Service Overview

### 1.1 Mission Statement

`svc-16` Notification Engine (`BC-MDH-10`) is the event-driven, multi-channel **notification hub** for the Medhen Platform. It orchestrates the templating, delivery, routing, and tracking of all outbound communications (SMS, Email, In-App) to customers, agents, and internal staff.

The service is fundamentally asynchronous. It listens to domain events across the platform (e.g., `PolicyBound`, `ClaimSubmitted`, `PaymentOverdue`), renders the appropriate template based on the customer's locale and channel preferences, and dispatches the message via the Integration ACL. 

The service owns the following strict responsibilities:
1. **Template Management:** Storing and rendering dynamic, multi-lingual (Amharic/English) notification templates per channel and event type.
2. **Preference Routing:** Determining the correct channel for a message based on the customer's communication preferences (synced from Party Management).
3. **Dispatch Orchestration:** Sending payloads to the Integration ACL and managing retries, fallback channels, and Dead Letter Queues (DLQs).
4. **Delivery Tracking:** Maintaining the canonical timeline of sent, delivered, and failed notifications.
5. **Scheduled Notifications:** Processing future-dated communications like renewal reminders or cancellation warnings.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem** | Disjointed notification logic scattered across Policy, Billing, and Claims engines leads to inconsistent branding, redundant logic, and an inability to respect centralized customer communication preferences or regulatory disclosure timelines. |
| **Value** | Centralizes all outbound communication, ensuring strict adherence to branding, opt-out compliance (DPP), and delivery traceability. Allows business teams to manage messaging templates dynamically without codebase changes. |
| **Stakeholders** | Marketing, Customer Success, Compliance (for statutory notices), Integration Engineering. |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Status / Phase |
|:---|:---|:---|
| `CAP-NOT-001..004` | Multi-channel dispatch (SMS, Email, In-App). Push in Phase 4. | Phase 1 |
| `CAP-NOT-005` | Template engine supporting dynamic context, channel, and locale mapping. | Phase 1 |
| `CAP-NOT-006` | Event-driven automated dispatch based on platform domain events. | Phase 1 |
| `CAP-NOT-007` | Delivery tracking and status Webhook processing. | Phase 1 |
| `CAP-NOT-008` | Preference enforcement (opt-in/opt-out routing rules). | Phase 1 |
| `CAP-NOT-009` | Scheduled/delayed notification execution (e.g., D-30 renewal). | Phase 1 |
| `CAP-NOT-010` | Full notification history logging (7-year retention). | Phase 1 |

### 1.4 In-Scope / Out-of-Scope Responsibilities

**In-Scope:**
* Subscribing to platform events and triggering notifications.
* Rendering Handlebars/Go-templates with event payload context.
* Enforcing customer communication preferences before sending.
* Dispatching to the Integration ACL.
* Storing the historical log of sent/failed messages for 7 years.

**Out-of-Scope:**
* Managing external vendor APIs directly (e.g., Twilio, SendGrid, Ethio Telecom). This is abstracted by the Integration ACL (`BC-MDH-18`).
* Owning the User Interface for customer preferences. (`pc-party-mgmt-svc` owns the profile UI APIs and broadcasts preference updates).
* Generating PDF documents (e.g., Policy Schedules). The notification engine links to documents generated by `pc-document-mgmt-svc`.

---

## 2. Technology Stack

### 2.0 Architecture Narrative

`svc-16` is a high-throughput **event-consumer** and orchestration engine. Because notifications can spike (e.g., batch renewal runs), the architecture relies heavily on Kafka for buffering inbound triggers. The service employs an **Inbox Pattern** to ensure exactly-once processing of platform events, and a robust retry state machine for transient downstream integration failures. Templates are cached heavily in-memory/Redis to minimize database load during high-velocity dispatches.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / runtime | **Go 1.26.x** | High-concurrency throughput for asynchronous worker pools. |
| API — Internal | **gRPC** | Sync Ad-hoc dispatch and history fetching. |
| Event backbone | **Kafka** + **Avro** | Consumption of cross-platform triggers; publishing delivery statuses. |
| Primary store | **PostgreSQL 18.x** | ACID guarantees for notification status tracking and inbox idempotency. |
| Template Engine | **Go `text/template`** | High-performance, secure string templating. |
| Cache | **Redis** | Caching templates and user preference profiles. |
| Scheduler | **Temporal.io** / **PostgreSQL Job Queue** | Managing future-dated scheduled notifications (CAP-NOT-009). |

### 2.2 Configuration Reference

| Key | Default | Purpose |
|:---|:---|:---|
| `dispatch.retry.max_attempts` | `3` | Maximum attempts before marking notification as FAILED. |
| `dispatch.retry.backoff_ms` | `5000` | Exponential backoff base. |
| `template.cache.ttl` | `1h` | Time-to-live for compiled templates in memory. |
| `retention.history_years` | `7` | Regulatory retention threshold for sweeping old notification logs. |

---

## 3. Functional Requirements

### 3.1 Event-Driven Dispatch (`FR-NOT-EVT-*`)

- **FR-NOT-EVT-1 — Subscription.** The service SHALL subscribe to registered platform events (e.g., `PolicyBound`, `PaymentReceived`). Upon consumption, it MUST map the event type to an active Notification Template mapping.
- **FR-NOT-EVT-2 — Inbox Idempotency.** The service SHALL process inbound Kafka events using the Inbox pattern, guaranteeing exactly-once dispatch even if the same Kafka offset is read twice.
- **FR-NOT-EVT-3 — Payload Hydration.** If an inbound event lacks sufficient context for rendering (e.g., missing customer phone number), the service SHALL query the appropriate authoritative service (e.g., `pc-party-mgmt-svc`) via gRPC before dispatch.

### 3.2 Template Engine (`FR-NOT-TMP-*`)

- **FR-NOT-TMP-1 — Dimensions.** A template SHALL be resolved based on three dimensions: `EventType` + `Channel` + `Locale` (e.g., `PolicyBound` + `SMS` + `am-ET`).
- **FR-NOT-TMP-2 — Fallbacks.** If a specific locale template is unavailable, the service SHALL fallback to the default locale (English `en-US`) template.
- **FR-NOT-TMP-3 — Variables.** Templates SHALL support safe variable substitution from the triggering event's payload (e.g., `Hello {{ party.first_name }}, your policy {{ policy.number }} is active.`).

### 3.3 Preference Routing (`FR-NOT-PREF-*`)

- **FR-NOT-PREF-1 — Preference Enforcement.** Before dispatching, the service SHALL verify the customer's preferences. If a customer has explicitly opted out of a channel (e.g., "No SMS for Marketing"), the dispatch for that channel MUST be aborted and logged as `SUPPRESSED`.
- **FR-NOT-PREF-2 — Statutory Override.** Notification templates flagged as `STATUTORY` (e.g., Cancellation Notice, Renewal Disclosure) SHALL bypass marketing opt-out preferences and force delivery.
- **FR-NOT-PREF-3 — Local Caching.** The service SHALL maintain a read-optimised projection of customer communication preferences, populated by consuming `CustomerProfileUpdated` events from `BC-MDH-01`.

### 3.4 Delivery & Tracking (`FR-NOT-TRK-*`)

- **FR-NOT-TRK-1 — State Machine.** Every dispatched notification SHALL track its status: `PENDING` → `DISPATCHED` → `DELIVERED` / `FAILED`.
- **FR-NOT-TRK-2 — Webhook Ingestion.** The service SHALL expose an internal endpoint to receive delivery receipts (Delivery Reports) forwarded from the Integration ACL, updating the notification status accordingly.
- **FR-NOT-TRK-3 — Fallback Channels.** If a primary channel fails (e.g., Push Notification failed), the service SHALL support evaluating a routing rule to attempt a secondary channel (e.g., fallback to Email).

### 3.5 Retention (`FR-NOT-RET-*`)

- **FR-NOT-RET-1 — History Retention.** All notification records, templates, and delivery statuses SHALL be persisted for a minimum of 7 years to satisfy audit and regulatory conduct reviews.

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Aggregate Roots

| Aggregate Root | Definition & Invariants | Emitted Events |
|:---|:---|:---|
| **`Notification`** | Represents a single logical message intended for a recipient. Tracks the lifecycle from creation to delivery confirmation.<br><br>**Invariants:**<br>• Cannot transition from `DELIVERED` back to `PENDING`.<br>• Must have a valid Template reference. | `NotificationDispatched`<br>`NotificationDelivered`<br>`NotificationFailed` |
| **`NotificationTemplate`** | The content structure for a specific event/channel/locale combination. Versioned to allow historical auditing of what was exactly sent.<br><br>**Invariants:**<br>• Must compile successfully via the text/template engine upon saving. | `TemplateUpdated` |
| **`RoutingPreference`** | A read-projection of the party's communication opt-ins. | — |

### 4.2 Unit of Work (UoW) & Transaction Boundary

When `svc-16` consumes an event (e.g., `PolicyBound`), the UoW guarantees:
1. Insert of the `event_id` into the `inbox` table to prevent duplicates.
2. Creation of the `Notification` aggregate in PostgreSQL (`PENDING`).
3. (Async) A worker picks up the `PENDING` notification, formats it, calls the Integration ACL, and updates state to `DISPATCHED`.

---

## 5. API Specifications

### 5.1 Internal gRPC API (Sync / Read / Ad-Hoc)

**Service Definition:** `medhen.platform.notification.v1.NotificationService`

| RPC | Request | Response | Notes |
|:---|:---|:---|:---|
| `DispatchAdHoc` | `target_id`, `template_code`, `context_json` | `NotificationId` | Used by Admin UIs or jobs to force a send outside of domain events. |
| `GetNotificationHistory` | `party_id`, `limit`, `offset` | `List<NotificationView>` | Consumed by the Customer 360 UI in the Party context. |
| `UpdateDeliveryStatus` | `notification_id`, `status`, `vendor_msg` | `Empty` | Invoked by Integration ACL when vendor webhooks fire. |

---

## 6. Event Schemas & Contracts (Avro)

### 6.1 Published Events (Outbox)

`svc-16` publishes status updates for reporting and cross-BC reactivity.

| Event | Topic | Partition Key | Notes |
|:---|:---|:---|:---|
| `NotificationDelivered` | `platform.notification.status.v1` | `notification_id` | Used by Reporting / Analytics. |
| `NotificationFailed` | `platform.notification.status.v1` | `notification_id` | Can trigger ops alerts or fallback workflows in other BCs. |

### 6.2 Consumed Events (Inbox)

The service subscribes to multiple topics across the platform:
- `platform.policy.lifecycle.v1` (PolicyBound, PolicyCancelled, RenewalIssued)
- `platform.billing.lifecycle.v1` (PaymentReceived, PaymentOverdue)
- `platform.claim.lifecycle.v1` (ClaimSubmitted, ClaimSettled)
- `platform.party.profile.v1` (CustomerProfileUpdated - for preference syncing)

---

## 7. Behaviour-Driven Scenarios (BDD)

**Scenario: NOT-BDD-01 | Successful Event-Driven SMS**
* **Given** an active SMS template for `PaymentReceived` in `en-US`
* **And** a customer with no SMS opt-out preferences
* **When** a `PaymentReceived` event arrives on the Kafka topic
* **Then** the service resolves the template and hydrates it with payment context
* **And** a `Notification` is created in `PENDING` state
* **And** the payload is dispatched to the Integration ACL
* **And** the status updates to `DISPATCHED`

**Scenario: NOT-BDD-02 | Preference Suppression (Marketing Opt-Out)**
* **Given** a template for `PromoOffer` categorized as `MARKETING`
* **And** a customer who has explicitly opted out of `MARKETING` communications
* **When** an ad-hoc dispatch command is received for this customer
* **Then** the service logs the Notification as `SUPPRESSED`
* **And** no payload is sent to the Integration ACL

**Scenario: NOT-BDD-03 | Statutory Override**
* **Given** a template for `CancellationNotice` categorized as `STATUTORY`
* **And** a customer who has opted out of ALL emails
* **When** the `PolicyCancelled` event is consumed
* **Then** the service bypasses the opt-out preference
* **And** the email payload is dispatched to the Integration ACL

---

## 8. Data Ownership & Persistence

### 8.1 PostgreSQL Schema

#### 8.1.1 Notifications Table

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    party_id UUID NOT NULL,
    template_code VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- SMS, EMAIL, IN_APP
    category VARCHAR(50) NOT NULL, -- MARKETING, TRANSACTIONAL, STATUTORY
    status VARCHAR(30) NOT NULL, -- PENDING, DISPATCHED, DELIVERED, FAILED, SUPPRESSED
    recipient_address VARCHAR(255) NOT NULL,
    rendered_content TEXT, -- Stored for 7-year audit retention
    vendor_receipt_id VARCHAR(255),
    error_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notif_party ON notifications(party_id);
```

#### 8.1.2 Templates Table

```sql
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    category VARCHAR(50) NOT NULL,
    subject_template TEXT,
    body_template TEXT NOT NULL,
    version INT NOT NULL,
    UNIQUE(code, channel, locale, version)
);
```

### 8.2 Data Retention

To comply with enterprise standards and potential legal disputes regarding statutory disclosures, all records in the `notifications` table are retained for **7 years**. A PostgreSQL partition strategy (partitioned by `created_at` year) will be utilized to facilitate efficient querying and eventual lifecycle archival/sweeping.

---

## 9. Integration & Dependency Contracts

### 9.1 Upstream Dependencies

| Service | Contract | Coupling | Fallback/Degraded State |
|:---|:---|:---|:---|
| **`pc-party-mgmt-svc`** | gRPC `GetPartyContactInfo` | Sync | If Party is down, notification sits in `PENDING` retry queue. |
| **`pc-integration-acl`** | gRPC `DispatchMessage` | Sync | Maps internal request to Twilio/SendGrid/Ethio APIs. Handles circuit breaking to external vendors. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | SLO | Consequence of Breach | Measurement |
|:---|:---|:---|:---|
| **Dispatch Latency** | P95 < 5s | Customer confusion (e.g., waiting for OTP or payment receipt). | Time from Kafka consumption to ACL dispatch request. |
| **Availability (Processing)** | 99.5% | Delayed communications. Backlog buildup in Kafka. | Uptime of the asynchronous processor pods. |
| **Throughput (Sustained)** | > 50 msg/sec | System chokes during batch renewal/billing runs. | Prometheus throughput counters. |

---

## 11. Observability Specification

The service utilizes the standard `pc-telemetry-sdk`.

### 11.1 Custom Domain Metrics

- `notification_dispatch_total{channel="SMS|EMAIL", status="SUCCESS|FAILED|SUPPRESSED"}`
- `notification_template_render_errors_total`
- `inbox_event_processing_latency_seconds_bucket`

### 11.2 Distributed Tracing

Span Context must propagate from the inbound Kafka Event → Notification Processing → Integration ACL gRPC Call, ensuring a single Trace ID tracks the origin event to the external vendor dispatch.

---

## 12. Operational Runbooks

### 12.1 Handling External Vendor Outages
**Symptom:** Integration ACL returns `ResourceExhausted` or `Unavailable` consistently.
**Action:**
`svc-16` will automatically leverage exponential backoff. If the outage exceeds the `max_attempts`, messages will be routed to a `notifications_dlq` table. Once the vendor recovers, an operator can replay the DLQ:
```bash
grpc_cli call medhen.platform.notification.v1.AdminService/ReplayDLQ "channel: 'SMS'"
```

---

## 13. Engineering Definition of Done (DoD)

1. **Test Coverage:** Template rendering logic and preference enforcement state machine must have > 90% unit test coverage.
2. **BDD Scenarios:** All 3 scenarios in §7 pass in integration tests.
3. **Inbox Tests:** Kafka consumer tests prove that duplicate event offsets result in zero duplicate notifications sent.
4. **Load Testing:** Demonstrate sustained capability to process 10,000 notifications in a 10-minute window.
