# svc-01: Party & Customer Management Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-01-v1 |
| **Service ID** | `svc-01` |
| **Service Name** | Party Management (PTY) |
| **Bounded Context** | `BC-MDH-01` — Party & Customer Management |
| **Version** | 1.0 |
| **Status** | Draft |
| **Date** | 2026-07-16 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-0 |
| **Deploy Mode** | Microservice (`pc-party-mgmt-svc`) |
| **Target Repo** | `Platform Core/dev/services/pc-party-mgmt-svc` |
| **Phase** | Phase 1 (Core MVP) |
| **PRD Anchor** | [Platform Core PRD](../../../../../../docs/prd/Medhen-Platform-PRD.md) (`REQ-PTY-*`) |
| **Capability Anchor** | [Capability Doc BC-MDH-01](../../../../../../docs/prd/Medhen-Platform-Capability-Document.md#bc-mdh-01--party--customer-management-pc-party-mgmt-svc) |
| **Capabilities** | `CAP-PARTY-001` to `CAP-PARTY-A3` |
| **Methodologies** | DDD · Hexagonal · EDA · CQRS · Transactional Outbox |
| **Companion Specs** | `svc-16` IAM · `svc-18` Integration (Fayda) · `svc-13` Policy · `svc-14` FinCrime |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-16 | Initial Tier-0 specification covering Sections 1-13. Drafted against PRD capabilities (`REQ-PTY-001` through `062`). Radically expanded for comprehensive implementation clarity. |

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

`svc-01` Party Management (`BC-MDH-01`) is the **central, foundational registry** of every individual and organization that interacts with the Medhen Platform. It represents the "Customer Master Data". This includes policyholders, insureds, beneficiaries, agents, brokers, adjusters, third-party service providers, and reinsurers. 

The service adheres strictly to the **ACORD party model**, which defines a singular `Party` record that can possess many `Roles`. It is completely product-agnostic; the same underlying identity record represents a party regardless of whether they hold a Motor, Life, or Property relationship. It acts as the ultimate authority for Identity Verification (KYC), sanctions screening state, and serves the **Customer-360 view** consumed by all other operational platform modules.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem Space** | Legacy systems typically silo customer data per policy or line-of-business. This causes massive data duplication, invalid addresses, disjointed customer service, compliance failures (e.g., missing sanctions checks), and prevents cross-selling. |
| **Value Delivered** | Provides a highly normalized, centralized identity substrate. It eliminates duplicate data entry, dramatically speeds up policy binding by pre-verifying KYC, ensures compliance with Ethiopian Data Protection laws, and empowers agents with a 360-degree view of all customer interactions and holdings. |
| **Stakeholders** | Underwriters, Claims Adjusters, Compliance & Risk Officers, Brokers/Agents, Customer Service Representatives, Legal (Data Privacy). |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Primary REQ | Phase |
|:---|:---|:---|:---|
| `CAP-PARTY-001` | **Party Registration:** Registration workflows for Individual and Organization types. Includes deep fuzzy matching, duplicate detection, and destructive/non-destructive merging logic. | `REQ-PTY-001` - `004` | 1 |
| `CAP-PARTY-002` | **Profile Management:** Management of complex Ethiopian hierarchical addresses, contact methods, encrypted bank accounts, and lifecycle statuses (e.g., Active, Suspended, Blacklisted). | `REQ-PTY-010` - `015` | 1 |
| `CAP-PARTY-003` | **KYC Processing:** Orchestrates identity document uploads, manual verification sagas, expiry tracking, and aggregate KYC status resolution (blocking bind events if unverified). | `REQ-PTY-020` - `023` | 1 |
| `CAP-PARTY-004` | **Party Roles:** Management of distinct role assignments with varying JSON-based attribute schemas (e.g., Agent License # vs Broker Accreditation). | `REQ-PTY-030` - `032` | 1 |
| `CAP-PARTY-005` | **Fuzzy Search:** Elasticsearch-backed lookup across names, IDs, phones, and emails with sub-500ms P95 latency. | `REQ-PTY-040` - `042` | 1 |
| `CAP-PARTY-006` | **Customer 360°:** A CQRS read-projection aggregating interactions, policies, claims, and billing records into a unified timeline and relationship map. | `REQ-PTY-050` - `052` | 1 |
| `CAP-PARTY-A1` | **Consent & Data Rights:** Versioned opt-in/opt-out for marketing and data processing; orchestrates PII anonymization per Ethiopian DPP. | `REQ-PTY-060` | 1 |
| `CAP-PARTY-A2` | **Vulnerable Customers:** Flagging capability for UK Consumer Duty equivalent care standards. | `REQ-PTY-061` | 2 |
| `CAP-PARTY-A3` | **Sanctions / PEP Hook:** Async delegation to the Financial Crime module at onboarding to gate policy inception on positive hits. | `REQ-PTY-062` | 2 |

### 1.4 Bounded Context Responsibilities (`BC-MDH-01`)

The Party domain boundary is strict. It does **not** store policy details, claims data, or payment ledgers. Instead, it maintains the canonical identity, and projects external relationships purely for the Customer 360 view.

| Owns | Exposes | Produces (via Outbox) | Invariants |
|:---|:---|:---|:---|
| `Party` aggregate (identity, profile) | REST CRUD & Search API | `pc.party.created.v1`, `pc.party.updated.v1` | Party IDs are UUIDv4, immutable, unique |
| KYC Profile & Status | KYC REST API | `pc.party.verified.v1`, `pc.party.kyc_rejected.v1` | Document expiry downgrades aggregate KYC status |
| Role associations | gRPC Read API | `pc.party.merged.v1` | A party can possess the same role only once simultaneously |
| Customer 360 Projections | REST Read API | — | Aggregated from downstream Kafka topics |

### 1.5 Context Map

```mermaid
flowchart TB
    subgraph Core["svc-01 Party Management (BC-MDH-01)"]
        PTY["Party Authoring API (Write)"]
        KYC["KYC & Verification"]
        C360["Customer 360 Projections (Read)"]
        ES["Elasticsearch (Search Index)"]
        
        PTY -->|CDC / Outbox| C360
        PTY -->|CDC / Outbox| ES
    end

    subgraph Core_Consumers["Downstream Consumers"]
        POL["pc-policy-svc (BC-03)"]
        CLM["pc-claims-svc (BC-07)"]
        BIL["pc-billing-svc (BC-06)"]
        COM["pc-commission-svc (BC-10)"]
    end

    subgraph Interoperability["Upstream / Integrations"]
        IAM["pc-iam-svc (BC-16)"]
        FAY["Fayda ACL (BC-18)"]
        FIN["pc-fincrime-svc (BC-14)"]
        NOT["pc-notification-svc (BC-09)"]
    end

    POL -->|Get Policyholder (gRPC)| PTY
    CLM -->|Get Claimant (gRPC)| PTY
    BIL -->|Get Bank Accounts (gRPC)| PTY
    COM -->|Get Producer Info (gRPC)| PTY
    
    PTY -->|Verify Fayda ID (REST)| FAY
    PTY -->|Sanctions Onboarding Hook (Kafka)| FIN
    IAM -->|Link Portal User (gRPC)| PTY
    KYC -->|Trigger Expiry Reminder (gRPC)| NOT
```

---

## 2. Technology Stack & Architecture

### 2.0 Operations-Plane Architecture Narrative

`svc-01` operates at the heart of the transactional core. Because Party data is queried continuously by almost every system (during quoting, rating, claim FNOL, and billing runs), read availability and latency are paramount. 

To achieve this, the architecture utilizes **CQRS (Command Query Responsibility Segregation)**:
1. **Write Model:** A highly normalized PostgreSQL database handles commands (`CreateParty`, `UpdateAddress`). It utilizes the Transactional Outbox pattern to atomically commit state changes and domain events.
2. **Search Model:** A Debezium CDC connector tails the PostgreSQL WAL (Write-Ahead Log) and pushes updates to a Kafka topic, which is consumed by a dedicated indexer that hydrates **Elasticsearch**. This enables instantaneous fuzzy matching (crucial for duplicate detection) without table-scanning Postgres.
3. **Customer 360 Read Model:** A projection engine listens to external events (`PolicyBound`, `ClaimOpened`, `InvoicePaid`) from Kafka and writes flattened, read-optimized JSON documents into a dedicated MongoDB/Postgres-JSONB projection table for sub-50ms retrieval of the full customer timeline.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / runtime | **Go 1.26.x** | Maximum concurrency for gRPC handling; low memory footprint. |
| API — external/UI | **REST/JSON**, OpenAPI 3.1 | Product Manager / Agent Portal authoring UI. |
| API — internal | **gRPC / Protobuf** | Sub-millisecond latency for internal resolution (e.g. `ResolveParty`). |
| Primary Store (Writes) | **PostgreSQL 18.x** | ACID guarantees, constraints, row-level locking for concurrent edits. |
| Search / Index | **Elasticsearch 8.x** | Advanced analyzers for Ethiopian name fuzziness and n-gram matching. |
| Event Backbone | **Kafka + Avro** | Durable domain events via Outbox pattern, schema registry compatibility. |
| Document Storage | **MinIO (S3 API)** | Storing KYC images, passports, and profile photos. |
| Cache | **Redis** | LRU caching of frequently accessed party profiles. |

### 2.2 Core Configuration References

| Config Key | Default | Purpose |
|:---|:---|:---|
| `search.duplicate_threshold` | `0.85` | Confidence score (0.0 to 1.0) required to trigger a `DUPLICATE_WARNING` during registration. |
| `kyc.expiry_warning_days` | `30` | Lead time to trigger KYC renewal notifications. |
| `kyc.auto_reject_fayda_mismatch`| `true` | If the provided name heavily diverges from the Fayda integration response, automatically set KYC to `REJECTED`. |
| `outbox.poll_interval_ms` | `100` | Rate at which the outbox relay sweeps for unpublished events. |

---

## 3. Functional Requirements & State Machines

### 3.1 Detailed Requirement Catalog (RFC 2119)

#### 3.1.1 Party Registration, Duplicate Detection, and Merging (`FR-PTY-REG-*`)
- **FR-PTY-REG-1 — Individual Reg:** The service SHALL expose `POST /v1/parties/individual` to register an individual, mandating `first_name`, `last_name`, `dob`, `gender`, `national_id_type` (Fayda/Kebele/Passport), and `national_id_number`.
- **FR-PTY-REG-2 — Organization Reg:** The service SHALL expose `POST /v1/parties/organization` to register an organization, mandating `legal_name`, `registration_number`, `tin`, and `industry_code`.
- **FR-PTY-REG-3 — Duplicate Detection Algorithm:** Before committing a registration, the service SHALL query Elasticsearch. 
  - **Exact match:** on `national_id_number`, `tin`, or `email`.
  - **Fuzzy match:** Levenshtein distance on (`first_name` + `last_name`) AND exact match on `dob`.
  - If a match > `search.duplicate_threshold` is found, the system MUST return a `409 Conflict` with `DuplicatePartyDetected`, yielding an array of `candidate_party_ids`. The client MUST re-submit with `override_duplicate_flag=true` to force creation.
- **FR-PTY-REG-4 — Party Merging:** The service SHALL expose `POST /v1/parties/merge`. It accepts `source_party_id` (duplicate) and `target_party_id` (survivor). The system SHALL:
  1. Set `source` status to `MERGED`.
  2. Set `source.surviving_party_id` to `target_party_id`.
  3. Emit `pc.party.merged.v1`. Downstream systems (Policy, Claims) MUST consume this to re-parent their foreign keys.

#### 3.1.2 Profile & Lifecycle Management (`FR-PTY-PRF-*`)
- **FR-PTY-PRF-1 — Admin Units:** The system SHALL manage Ethiopian addresses utilizing the standard hierarchy: `Region` → `Zone` → `Woreda` → `Kebele`. Invalid hierarchies MUST be rejected.
- **FR-PTY-PRF-2 — Contact Validation:** Email addresses SHALL conform to RFC-5322. Phone numbers SHALL conform to E.164 (e.g., `+251911234567`).
- **FR-PTY-PRF-3 — Bank Account Masking:** Bank Accounts (used for claims disbursement and premium refunds) SHALL be AES-256 encrypted at rest. API responses SHALL return masked strings (e.g., `******1234`) unless explicitly requested with elevated `finance.decrypt` scope.
- **FR-PTY-PRF-4 — PII Redaction / Erasure:** If a Data Subject Erasure request is approved, the service SHALL scramble `first_name`, `last_name`, `email`, and `phone` with a one-way hash, replacing the status with `ANONYMIZED`.

#### 3.1.3 KYC Verification (`FR-PTY-KYC-*`)
- **FR-PTY-KYC-1 — Document Upload:** The service SHALL accept `multipart/form-data` uploads of KYC images. These are pushed to MinIO. The returned `object_key` is stored against the party's KYC profile.
- **FR-PTY-KYC-2 — Aggregate Status:** A party's overall KYC status is derived dynamically:
  - `VERIFIED`: All required documents are present, manually approved, and `expiry_date` > `now()`.
  - `PENDING`: Documents uploaded but awaiting manual verification or Fayda async callback.
  - `EXPIRED`: One or more required documents have passed their expiry date.
  - `REJECTED`: Documents were manually flagged as fraudulent/illegible.
- **FR-PTY-KYC-3 — Bind Gate:** `pc-policy-svc` calls `VerifyKYCState(party_id)` during the binding flow. If the response is not `VERIFIED`, the policy CANNOT be bound.

### 3.2 State Machine Definition (Party Lifecycle)

| Current State | Trigger Command | Target State | Guards & Preconditions |
|:---|:---|:---|:---|
| `—` | `RegisterIndividual` / `Organization`| `ACTIVE` | Passes duplicate detection (or overridden). |
| `ACTIVE` | `SuspendParty` | `SUSPENDED` | Requires suspension reason. Used for suspected fraud. |
| `ACTIVE` | `BlacklistParty` | `BLACKLISTED` | Requires legal/compliance approval. Blocks all new policies. |
| `ACTIVE`/`SUSPENDED`| `MergeParty` | `MERGED` | Becomes a tombstone pointing to survivor. |
| `ACTIVE` | `AnonymizeParty` | `ANONYMIZED` | Must not have active financial relationships. Reversible hash applied. |

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Bounded Context Boundary (`BC-MDH-01`)

The Party domain acts as the upstream authority. Downstream contexts (`Policy`, `Claims`, `Billing`) do not redefine customer details; they reference `party_id` and fetch effective configuration dynamically.

### 4.2 Aggregate Roots

| Aggregate Root | Definition & Invariants | Emitted Events |
|:---|:---|:---|
| **`Party`** | Represents an individual or organization. Owns `Address`, `Contact`, `BankAccount`, and `ConsentRecord`. <br><br>**Invariants:**<br>• Must have exactly one PRIMARY address.<br>• IDs (`national_id`, `tin`) must be unique across the tenant unless the status is `MERGED`. | `PartyCreated`<br>`PartyUpdated`<br>`PartyStatusChanged`<br>`PartyMerged`<br>`PartyAnonymized` |
| **`KYCProfile`** | The aggregate capturing identity validation. Separated from `Party` to allow high-frequency document uploads and status changes without locking the core profile.<br><br>**Invariants:**<br>• Cannot approve a document without an expiry date (if applicable to type). | `KYCDocumentUploaded`<br>`KYCStatusEvaluated` |

### 4.3 Entities vs Value Objects

| Concept | Type | Justification |
|:---|:---|:---|
| `Party` | Entity (Root) | Possesses global identity (`party_id`). |
| `Address` | Entity (Local) | Identified by `address_id` within a `Party`; can be modified or deleted. |
| `BankAccount` | Entity (Local) | Secure entity requiring dedicated lifecycle (verification, masking). |
| `PartyRole` | Entity (Local) | Effective-dated assignment (`role_id`) tied to the parent `Party`. |
| `ConsentRecord` | Value Object | Entirely replaced when consent preferences are modified. No inherent identity. |
| `AdminUnit` | Value Object | Region/Zone structural mapping. |

### 4.4 Command Catalog

| Command | Aggregate | Pre-conditions | Post-conditions (Success) | Domain Exception |
|:---|:---|:---|:---|:---|
| `RegisterParty` | `Party` | ID constraints met, not a duplicate. | Party is `ACTIVE`, Outbox event queued. | `DuplicatePartyDetected`, `InvalidNationalId` |
| `UpdateAddress` | `Party` | Valid AdminUnit hierarchy. | Address updated, primary flags resolved. | `InvalidAdminUnitMapping` |
| `ApplyRole` | `Party` | Role doesn't already overlap in date. | Role added to party collection. | `RoleAlreadyActive` |
| `MergeParty` | `Party` | Source is not `MERGED`. | Source is `MERGED`, points to target. | `PartyAlreadyMerged`, `InvalidMergeTarget` |

### 4.5 Unit of Work (UoW) & Transaction Boundary

The UoW guarantees that exactly three things occur atomically in a single PostgreSQL transaction:
1. The domain state mutation (e.g., updating the `status` column in `parties`).
2. The domain event serialization into the `outbox` table (Avro binary).
3. The insertion of a record into the `audit_ledger` table (append-only).

Optimistic Locking is employed via a `version` integer column on every Aggregate Root. If two users edit the same profile simultaneously, the second yields an `OptimisticLockException` (HTTP 409).

---

## 5. API Specifications (REST & gRPC)

### 5.1 REST API (Authoring & UI)

**Base path:** `/api/pc-party-mgmt/v1`

| Method | Endpoint | Purpose | Security / RBAC |
|:---|:---|:---|:---|
| `POST` | `/parties/individual` | Register an individual | `party.create` |
| `POST` | `/parties/organization` | Register an organization | `party.create` |
| `PATCH`| `/parties/{id}` | Update core demographic fields | `party.update` |
| `POST` | `/parties/{id}/addresses` | Add a new address | `party.update` |
| `POST` | `/parties/{id}/kyc-documents` | Multipart upload ID image | `party.kyc.upload` |
| `POST` | `/parties/{id}/kyc-documents/{doc}/verify` | Approve/Reject document | `compliance.officer` |
| `POST` | `/parties/merge` | Execute party merge | `data.steward` |
| `GET`  | `/parties/search?q={query}` | Fuzzy Elasticsearch lookup | `party.read` |
| `GET`  | `/parties/{id}/360` | Fetch aggregated projection | `party.read.360` |

#### 5.1.1 Payload Example: `RegisterIndividual`
```json
{
  "first_name": "Abebe",
  "last_name": "Kebede",
  "date_of_birth": "1985-06-15",
  "gender": "MALE",
  "national_id_type": "FAYDA",
  "national_id_number": "FYD-9988776655",
  "tin": "TIN-1234567",
  "primary_contact": {
    "type": "MOBILE",
    "value": "+251911223344"
  },
  "override_duplicate_flag": false
}
```

### 5.2 gRPC API (High-Throughput Reads)

The gRPC API is heavily consumed by `pc-policy-svc` (for quoting/binding) and `pc-claims-svc` (for FNOL).

**Service Definition:** `medhen.platform.party.v1.PartyResolutionService`

| RPC | Request | Response | SLA (P95) |
|:---|:---|:---|:---|
| `ResolveParty` | `party_id` | `PartySummary` (Demographics, Primary Address) | < 5ms |
| `VerifyKYCState` | `party_id` | `KYCStatus` (Enum: VERIFIED, PENDING, REJECTED) | < 5ms |
| `GetProducerDetails` | `party_id` (Role = AGENT/BROKER) | `ProducerProfile` (License #, Territory) | < 5ms |
| `GetSettlementAccounts`| `party_id` | `BankAccountsList` (Decrypted, restricted caller) | < 10ms |

### 5.3 Error Taxonomy (RFC 7807 Problem Details)

| Domain Exception | HTTP Code | Error Code | Client Action / Resolution |
|:---|:---|:---|:---|
| `DuplicatePartyDetected` | `409 Conflict` | `PTY-1001` | Review `candidate_ids`. Prompt user to confirm merge or supply override flag. |
| `InvalidAdminUnitMapping`| `400 Bad Request` | `PTY-1002` | Ensure Woreda belongs to the specified Zone. Update dropdowns. |
| `PartyStatusInvalid` | `422 Unprocessable`| `PTY-1003` | Cannot execute requested action (e.g. merge) because party is BLACKLISTED. |
| `FaydaIntegrationTimeout`| `502 Bad Gateway`| `PTY-1004` | Async KYC mode activated. Proceed with PENDING status. |

---

## 6. Event Schemas & Contracts (Avro)

All domain events are published to Kafka topics utilizing the Apicurio Schema Registry with `BACKWARD` compatibility mode.

### 6.1 Topic Mapping

| Event | Topic | Partition Key | Schema ID |
|:---|:---|:---|:---|
| `PartyCreated`, `PartyUpdated`, `PartyStatusChanged` | `platform.party.lifecycle.v1` | `tenant_id:party_id` | `PartyLifecycleEvent` |
| `PartyMerged` | `platform.party.merged.v1` | `tenant_id:survivor_id` | `PartyMergedEvent` |
| `KYCStatusEvaluated` | `platform.party.kyc.v1` | `tenant_id:party_id` | `KYCStatusEvent` |

### 6.2 Avro Schema: `PartyMergedEvent`

```json
{
  "namespace": "medhen.platform.party.v1",
  "type": "record",
  "name": "PartyMergedEvent",
  "fields": [
    {"name": "event_id", "type": "string", "logicalType": "uuid"},
    {"name": "tenant_id", "type": "string"},
    {"name": "source_party_id", "type": "string", "doc": "The ID of the duplicate party being merged."},
    {"name": "target_party_id", "type": "string", "doc": "The ID of the surviving party."},
    {"name": "merged_by", "type": "string", "doc": "Actor who performed the merge."},
    {"name": "occurred_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
  ]
}
```
*Impact:* Upon receiving this event, `pc-policy-svc` MUST execute `UPDATE policies SET policyholder_id = target_party_id WHERE policyholder_id = source_party_id`.

---

## 7. Behaviour-Driven Scenarios (BDD)

### 7.1 Duplicate Detection & Mitigation

**Scenario: PTY-BDD-01 | Register duplicate individual (Fuzzy Name match)**
* **Given** an existing Party "Abebe Kebede" with DOB "1985-06-15" (ID: `P-111`)
* **When** a broker submits `RegisterIndividual` for "Abebaw Kebede" with DOB "1985-06-15"
* **Then** the Elasticsearch analyzer calculates a Levenshtein similarity > 0.85
* **And** the service rejects the request with HTTP 409 `DuplicatePartyDetected`
* **And** returns `candidate_ids: ["P-111"]` in the error payload.

**Scenario: PTY-BDD-02 | Force creation over duplicate warning**
* **Given** the rejected payload from `PTY-BDD-01`
* **When** the broker resubmits the exact payload with `override_duplicate_flag: true`
* **Then** the service processes the registration
* **And** a new Party `P-112` is created successfully.

### 7.2 KYC and Lifecycle Gating

**Scenario: PTY-BDD-03 | KYC Expiry downgrades status**
* **Given** a Party with KYC status `VERIFIED`
* **And** their sole Passport document expires today
* **When** the midnight cron evaluates document expiry
* **Then** the Passport is marked `EXPIRED`
* **And** the aggregate Party KYC status is downgraded to `EXPIRED`
* **And** a `KYCStatusEvaluated` event is emitted, blocking future policy binding.

### 7.3 Party Merging

**Scenario: PTY-BDD-04 | Valid Merge Execution**
* **Given** `Party A` (Duplicate) and `Party B` (Survivor)
* **When** a Data Steward submits a `MergeParty` command targeting A into B
* **Then** `Party A` status transitions to `MERGED`
* **And** `Party A`'s `surviving_party_id` is set to `Party B`
* **And** any attempt to read `Party A` directly yields a HTTP 301 Redirect to `Party B`
* **And** a `PartyMergedEvent` is published to Kafka.

---

## 8. Data Ownership & Persistence Schema

`svc-01` utilizes PostgreSQL for the core normalized source of truth, and Elasticsearch for denormalized query access.

### 8.1 PostgreSQL Transactional Schema (Simplified)

```sql
-- Core Party Table
CREATE TABLE parties (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL, -- INDIVIDUAL, ORGANIZATION
    status VARCHAR(20) NOT NULL, -- ACTIVE, SUSPENDED, BLACKLISTED, MERGED, ANONYMIZED
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    
    -- Individual specific
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    dob DATE,
    gender VARCHAR(10),
    national_id_type VARCHAR(50),
    national_id_number VARCHAR(100),
    
    -- Org specific
    legal_name VARCHAR(200),
    registration_number VARCHAR(100),
    industry_code VARCHAR(50),
    
    tin VARCHAR(100),
    surviving_party_id UUID REFERENCES parties(id),
    version INT NOT NULL DEFAULT 1, -- Optimistic Locking
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_parties_national_id ON parties(tenant_id, national_id_number) WHERE status != 'MERGED';

-- Address Hierarchy (Ethiopian Admin Units)
CREATE TABLE addresses (
    id UUID PRIMARY KEY,
    party_id UUID REFERENCES parties(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- RESIDENTIAL, MAILING, BUSINESS
    is_primary BOOLEAN DEFAULT false,
    region VARCHAR(100) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    woreda VARCHAR(100) NOT NULL,
    kebele VARCHAR(100),
    house_number VARCHAR(50),
    version INT NOT NULL DEFAULT 1
);

-- Party Roles (ACORD Model)
CREATE TABLE party_roles (
    id UUID PRIMARY KEY,
    party_id UUID REFERENCES parties(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL, -- POLICYHOLDER, BROKER, AGENT, ADJUSTER
    attributes JSONB, -- E.g. {"license_no": "L-123", "territory": "Addis Ababa"}
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    UNIQUE (party_id, role)
);
```

### 8.2 Elasticsearch Index Mapping (`parties_index`)

Driven asynchronously via Debezium CDC + Kafka Sink Connector.

```json
{
  "mappings": {
    "properties": {
      "party_id": { "type": "keyword" },
      "tenant_id": { "type": "keyword" },
      "status": { "type": "keyword" },
      "full_name_ngram": { 
        "type": "text", 
        "analyzer": "autocomplete_analyzer" -- Supports fuzzy partial matching
      },
      "national_id": { "type": "keyword" },
      "phone_numbers": { "type": "keyword" }
    }
  }
}
```

---

## 9. Integration & Dependency Contracts

| External Service | Contract / Protocol | Coupling | Resilience / Fallback |
|:---|:---|:---|:---|
| **Fayda API (National ID)** | REST API (via `BC-MDH-18` API Gateway) | Sync (Inbound) | If Fayda is unreachable (Circuit Breaker OPEN), registrations succeed but KYC defaults to `PENDING`. Manual verification workflow absorbs the load. |
| **`pc-iam-svc` (Auth)**| gRPC `VerifyToken` / `LinkUser` | Sync | Hard dependency for API calls. Cached JWT validation locally. |
| **`pc-fincrime-svc`** | Kafka `pc.party.created.v1` | Async | Non-blocking at creation. However, FinCrime emits a hit event that gates policy binding downstream if unresolved. |
| **`pc-document-mgmt-svc`** | MinIO S3 API | Sync | Required for KYC uploads. Uploads fail if object storage is offline. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | SLO | Consequence of Breach | Measurement / Triggers |
|:---|:---|:---|:---|
| **Availability (gRPC)** | 99.99% | Entire platform (Quoting, Billing, FNOL) halted. | Prometheus `grpc_server_handled_total`. Alert if < 99.99% over 15m. |
| **Latency (gRPC Reads)** | P95 < 5ms | Sluggish UI quoting flows, timeout cascades. | OpenTelemetry Span Duration. |
| **Search Latency (ES)** | P95 < 500ms | Poor operator experience during branch onboarding. | OpenTelemetry REST Spans. |
| **Outbox Relay Lag** | < 2 seconds | ES Index and 360 Projections are stale. | Max timestamp difference between Postgres `created_at` and `published_at`. |

---

## 11. Observability Specification

The service utilizes the standard `pc-telemetry-sdk` for OpenTelemetry instrumentation.

### 11.1 Golden Signals (Prometheus)

- **Traffic:** `grpc_server_handled_total{method="ResolveParty"}`
- **Latency:** `grpc_server_handling_seconds_bucket`
- **Errors:** `rest_http_requests_total{status=~"5.."}`
- **Saturation:** `pgxpool_acquired_conns`, `redis_memory_usage_bytes`

### 11.2 Custom Domain Metrics

- `party_registrations_total{type="individual|organization", channel="portal|branch"}`
- `kyc_verifications_total{status="approved|rejected|expired"}`
- `duplicate_detections_total{resolution="blocked|overridden"}`
- `party_merges_total`

### 11.3 Logging (Structured `slog`)

All logs must contain tracing fields contextually injected by the SDK:

```json
{"level":"INFO","time":"2026-07-16T12:00:00Z","msg":"Duplicate party override triggered","tenant_id":"t-123","party_id":"p-456","overridden_by_user":"u-999","trace_id":"...","span_id":"..."}
```

---

## 12. Operational Runbooks

### 12.1 Elasticsearch Sync Stalled (Search is stale)
**Symptom**: `OutboxRelayLag` alert fires. Newly created parties are returning 404 in search.
**Action**: 
1. Check Kafka Debezium connector status: `curl localhost:8083/connectors/party-cdc/status`
2. If failed, restart task: `curl -X POST localhost:8083/connectors/party-cdc/tasks/0/restart`
3. If data corruption occurred, trigger full resync script: `kubectl exec -it deploy/pc-party-mgmt-svc -- ./bin/cli ops resync-es`

### 12.2 Undoing an Accidental Merge
**Symptom**: A Data Steward merged Party A into Party B incorrectly.
**Action**:
Merge operations are inherently destructive to foreign keys downstream. A strict undo script must be executed by Platform Engineering:
1. `POST /v1/ops/parties/merge-undo?source=A&target=B` (Requires `system.admin` token).
2. This resets Party A status to `ACTIVE`, clears `surviving_party_id`, and emits a `PartyUnmergedEvent` to instruct downstream systems to revert foreign keys based on historical audit ledgers.

---

## 13. Engineering Definition of Done (DoD)

Before `svc-01` can be deployed to the `staging` environment for Phase 1, the following quality gates MUST be passed:

1. **Test Coverage**: Core domain logic (Fuzzy matching, State Machine, Merge Logic) MUST have > 90% unit test coverage.
2. **BDD Integration**: All Gherkin scenarios defined in §7 MUST pass against ephemeral Postgres/ES/Kafka Testcontainers.
3. **Idempotency**: All `POST/PATCH` endpoints MUST pass automated double-submit replay tests via the `pc-idempotency-sdk`.
4. **Resilience**: Chaos tests MUST prove that a simulated Fayda API outage correctly degrades KYC to `PENDING` without dropping the registration payload.
5. **Security**: Application MUST pass SonarQube gates with zero Critical/High vulnerabilities. Bank account masking logic MUST be explicitly verified in integration tests.
