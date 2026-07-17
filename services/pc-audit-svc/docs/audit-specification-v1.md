# svc-17: Audit & Compliance Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-AUD-17-v1 |
| **Service ID** | `svc-17` |
| **Service Name** | Audit & Compliance (AUD) |
| **Bounded Context** | `BC-MDH-17` — Audit & Compliance |
| **Version** | 1.2 |
| **Status** | Draft |
| **Date** | 2026-07-17 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-0 |
| **Deploy Mode** | Microservice (`pc-audit-svc`) |
| **Target Repo** | `Platform Core/dev/services/pc-audit-svc` |
| **Phase** | Phase 1 (Core MVP) |
| **PRD Anchor** | [Platform Core PRD](../../../../../../docs/prd/Medhen-Platform-PRD.md) (`REQ-AUD-*`) |
| **Capability Anchor** | [Capability Doc BC-MDH-17](../../../../../../docs/prd/Medhen-Platform-Capability-Document.md#bc-mdh-17--audit--compliance-pc-audit-svc) |
| **Capabilities** | `CAP-AUD-001` to `CAP-AUD-A2` |
| **Methodologies** | Merkle Trees · Zero-Trust Ingestion · Open Table Formats (Iceberg) · OpenLineage |
| **Companion Specs** | `svc-16` IAM · `svc-01` Party Management |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-16 | Initial Tier-0 specification covering Sections 1-13. Drafted against PRD capabilities. |
| 1.1 | 2026-07-17 | Replaced MinIO with Apache Iceberg & Apache Ozone. Expanded DDD and Event Schemas. |
| 1.2 | 2026-07-17 | Integrated 6 Global Industry Standards: Merkle Trees, Payload Signing, Notary Anchoring, Examiner CLI Packs, AI Anomaly Detection, OpenLineage. |

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

`svc-17` Audit & Compliance (`BC-MDH-17`) is the single **immutable audit authority** for the entire Medhen Platform. It provides an append-only, mathematically verifiable ledger backed by **Merkle Trees** that records every data mutation, privileged action, and line of data provenance.

It acts as the globally standardized system of record for regulatory examinations, utilizing Zero-Trust upstream payload signing to prevent transport manipulation, external notary anchoring to guarantee non-repudiation, and AI-native real-time event streaming to detect insider threats.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem Space** | Regulators (like the NBE) require absolute proof that financial records have not been tampered with. Standard RDBMS and basic hash chains are insufficient; hash chains are O(N) to verify, transport layers can be intercepted, and exports can be repudiated. |
| **Value Delivered** | Provides a mathematically provable, immutable ledger mapped onto an enterprise data lake. Integrates with ML anomaly detection pipelines to actively prevent data exfiltration. Empowers external auditors with self-verifying CLI extraction packs and provides full OpenLineage data provenance. |
| **Stakeholders** | Compliance Officers, NBE Examiners, Risk Management, Security Operations (SOC), Platform Engineering, Data Platform Team. |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Primary REQ | Phase |
|:---|:---|:---|:---|
| `CAP-AUD-001` | **Data-change & OpenLineage audit:** Captures Who/What/When/Old/New along with tracing provenance across all bounded contexts. | `REQ-AUD-001` | 1 |
| `CAP-AUD-002` | **Zero-Trust Action audit:** Logs privileged actions with mandatory upstream payload signatures (ECDSA/RSA). | `REQ-AUD-002` | 1 |
| `CAP-AUD-003` | **Verifiable Merkle Tree log:** Stores records using a tamper-evident Merkle Tree with O(log N) inclusion proofs, anchored daily to an external Notary. | `REQ-AUD-003` | 1 |
| `CAP-AUD-004` | **Audit search & Time-Travel:** Provides filterable, high-performance search using Iceberg's native time-travel capabilities. | `REQ-AUD-004` | 1 |
| `CAP-AUD-005` | **Compliance export:** Generates digitally signed, immutable CSV/Parquet exports for external examination. | `REQ-AUD-004` | 1 |
| `CAP-AUD-A1` | **Legal hold:** Suspends deletion/crypto-shredding of specific records/entities under active investigation. | `REQ-AUD-010` | 2 |
| `CAP-AUD-A2` | **Examination egress pack:** Generates a self-contained data package and Offline Examiner CLI tool for mathematical verification by NBE. | `REQ-AUD-011` | 2 |

### 1.4 Bounded Context Responsibilities (`BC-MDH-17`)

| Owns | Exposes | Produces (via Outbox) | Invariants |
|:---|:---|:---|:---|
| `AuditLedgerEntry` aggregate | gRPC Ingestion API | `pc.audit.archived.v1` | All ingested payloads must possess a valid digital signature from the upstream producer. |
| `MerkleTreeRoot` | REST Search API | `pc.security.anomalies.v1` | Erasing PII must use crypto-shredding (destroying envelope key), preserving the Merkle Hash structure. |
| `LegalHold` entity | Async Export Jobs | `pc.audit.notary_anchored.v1`| Merkle roots cannot be altered retroactively. |

---

## 2. Technology Stack & Architecture

### 2.0 Operations-Plane Architecture Narrative

The architecture is built upon the principles of Google's Certificate Transparency (Trillian), integrating deeply with Apache Iceberg for cold analytical storage.

1. **Zero-Trust Ingestion:**
   - Upstream services (`pc-policy-svc`) construct the audit payload, digitally sign it with their workload mTLS private key, and publish it to Kafka. The Audit service verifies this signature via KMS, proving non-repudiation.
2. **Merkle Tree Structuring:**
   - Instead of a linear hash chain, events are inserted into a **Merkle Tree**. This generates a cryptographic Root Hash. If the NBE wants to verify a single transaction out of 10 million, the Audit service provides a small "Inclusion Proof" (a path of sibling hashes) that can be verified in O(log N) time.
3. **AI-Native Threat Detection:**
   - Upon successful ingestion, records are immediately streamed to a dedicated `platform.security.anomalies.v1` Kafka topic. The enterprise ML pipeline consumes this to detect Insider Threats (e.g., an adjuster viewing 500 PII records).
4. **Storage Tiering (Postgres -> Iceberg -> Ozone):** 
   - The Hot Ledger (Postgres) tracks the active Merkle Tree leaves. Micro-batches flush into **Apache Iceberg** (Cold/Analytical Ledger) running on **Apache Ozone** (WORM S3-compatible storage).
5. **Out-of-Band Anchoring:**
   - A daily cron job extracts the Merkle Root Hash and publishes it to an external API (The Notary), ensuring that even if the internal database is wiped by a rogue sysadmin, the historical root is unalterable.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Ingestion Buffer | **Kafka + Avro** | Deduplication, replayability, and durable buffering. |
| Hot Ledger / Merkle Tree| **PostgreSQL 18.x** | Fast state tracking for the active Merkle Tree leaves and roots. |
| Cold/Analytical Ledger| **Apache Iceberg** | Open table format; provides time-travel and schema evolution. |
| Deep Storage | **Apache Ozone** | Exabyte-scale object store. Provides native Object Lock (WORM). |
| Real-time Security | **Kafka Stream** | Emits normalized records to the ML anomaly pipeline. |
| Provenance | **OpenLineage** | Standardized metadata format for tracking data transformation flows. |

---

## 3. Functional Requirements & State Machines

### 3.1 Deeply Expanded Requirement Catalog (RFC 2119)

#### 3.1.1 Cryptographic Integrity & Zero-Trust Ingestion (`FR-AUD-ING-*`)
- **FR-AUD-ING-1 — Upstream Payload Signing:** The service SHALL reject any ingested payload (`*.cdc.v1` or gRPC) that does not contain a valid ECDSA/RSA digital signature originating from a recognized platform workload identity.
- **FR-AUD-ING-2 — Merkle Tree Construction:** The service SHALL construct a cryptographic Merkle Tree. Each `AuditLedgerEntry` constitutes a leaf node. The Root Hash SHALL be updated iteratively.
- **FR-AUD-ING-3 — OOB Notary Anchoring:** A daily scheduler SHALL extract the current Merkle Root Hash and publish it to the designated external Notary endpoint (e.g., NBE Secure Vault API).
- **FR-AUD-ING-4 — OpenLineage Provenance:** The service SHALL extract `TraceID`, `SpanID`, and OpenLineage metadata tags from incoming events and persist them natively in the Iceberg ledger to enable exact sequence reconstruction of complex financial transactions.

#### 3.1.2 Advanced Search & Export Orchestration (`FR-AUD-SCH-*`)
- **FR-AUD-SCH-1 — Time-Travel Queries:** The service SHALL leverage Apache Iceberg to support point-in-time snapshot queries.
- **FR-AUD-SCH-2 — Examination CLI Pack:** Upon exporting data, the system SHALL generate an "Examiner Pack". This archive MUST contain the queried Parquet data, a JSON manifest of Merkle Inclusion Proofs, and a standalone CLI binary (`medhen-auditor-cli`) allowing external regulators to verify the data integrity offline.

#### 3.1.3 AI Security & Retention (`FR-AUD-SEC-*`)
- **FR-AUD-SEC-1 — Anomaly Event Streaming:** Immediately upon validating an event, the service SHALL emit an `AnomalyDetectionEvent` to `platform.security.anomalies.v1` to power the real-time AI security pipeline.
- **FR-AUD-SEC-2 — Automated Crypto-Shredding:** Upon retention expiry for PII-linked records, the service SHALL command the KMS to permanently destroy the DEK. The Merkle Tree integrity MUST remain unbroken, as the tree hashes the ciphertext.

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Aggregate Roots

| Aggregate Root | Definition & Invariants |
|:---|:---|
| **`AuditLedgerEntry`** | Represents a single action or mutation in time. (Leaf Node) <br><br>**Invariants:**<br>• Must possess a verified `DigitalSignature`.<br>• Must be linked to a `MerkleTreeRoot`. |
| **`MerkleTreeRoot`** | Represents the immutable top hash of a given epoch. Published to the Notary. |
| **`ExportJob`** | Tracks the asynchronous saga of querying and building the Examiner CLI Pack. |

### 4.2 Entities vs Value Objects

| Concept | Type | Justification |
|:---|:---|:---|
| `ProvenanceMetadata` | Value Object | Encapsulates OpenLineage standard fields (`TraceID`, `RunID`). |
| `DigitalSignature` | Value Object | Contains the `KeyID` and `SignatureBytes` generated by the upstream producer. |

### 4.3 Command Catalog

| Command | Aggregate | Pre-conditions | Post-conditions (Success) |
|:---|:---|:---|:---|
| `AppendRecord` | `AuditLedgerEntry` | Valid digital signature. | Row inserted, Merkle Tree recalculated. |
| `PublishNotaryAnchor`| `MerkleTreeRoot` | Scheduled time reached. | Root Hash transmitted to external API. |
| `ExecuteCryptoShredding`| `AuditLedgerEntry`| Retention expired. | DEK destroyed in KMS. |

---

## 5. API Specifications (REST & gRPC)

### 5.1 REST API (Search & Export)
**Base path:** `/api/pc-audit/v1`

| Method | Endpoint | Purpose | Security |
|:---|:---|:---|:---|
| `GET`  | `/ledger/search` | Query events (supports pagination via cursor). | `audit.read` |
| `GET`  | `/ledger/{id}/proof` | Returns the Merkle Inclusion Proof for a specific record. | `audit.read` |
| `POST` | `/exports/examiner-pack` | Triggers the creation of an offline regulatory verification package. | `audit.export` |

### 5.2 gRPC API (Synchronous Ingestion)
| RPC | Request | Response |
|:---|:---|:---|
| `LogAction` | `ActionLogRequest` (Requires `digital_signature` field) | `ActionLogResponse` |

---

## 6. Event Schemas & Contracts (Avro)

### 6.1 Avro Schema Deep-Dive: `AuditLedgerEntry` (Internal / Iceberg)

```json
{
  "namespace": "medhen.platform.audit.v1",
  "type": "record",
  "name": "AuditLedgerEntry",
  "fields": [
    {"name": "seq_id", "type": "long"},
    {"name": "event_id", "type": "string", "logicalType": "uuid"},
    {"name": "producer_key_id", "type": "string", "doc": "KMS identifier of the workload that signed this."},
    {"name": "digital_signature", "type": "bytes", "doc": "ECDSA/RSA signature from the upstream producer."},
    {"name": "trace_id", "type": ["null", "string"], "doc": "OpenLineage / OpenTelemetry Trace ID."},
    {"name": "is_pii_encrypted", "type": "boolean"},
    {"name": "dek_reference_id", "type": ["null", "string"]},
    {"name": "delta_ciphertext", "type": ["null", "bytes"]},
    {"name": "merkle_leaf_hash", "type": "string", "doc": "The hash of this specific record used in the tree."}
  ]
}
```

---

## 7. Behaviour-Driven Scenarios (BDD)

### 7.1 Zero-Trust Ingestion Failure
**Scenario: AUD-BDD-01 | Rejected Malicious Payload**
* **Given** a malicious actor compromises the Kafka broker
* **When** they inject a forged `platform.policy.cdc.v1` event
* **Then** the Audit service extracts the `digital_signature` and `producer_key_id`
* **And** asks the KMS to verify the signature against the plaintext payload
* **And** the verification fails (as the attacker lacks the private mTLS key)
* **Then** the payload is discarded, an alert is fired to the SOC, and the Merkle tree remains unpolluted.

### 7.2 External Notary Verification
**Scenario: AUD-BDD-02 | Verifying Historical Immutability via Notary**
* **Given** a massive database corruption incident where internal Postgres hashes were altered
* **When** a security engineer checks the internal Merkle Root Hash for epoch 100
* **And** compares it against the external NBE Vault Notary API for epoch 100
* **Then** a mismatch is detected, proving the internal system was tampered with, triggering the Disaster Recovery protocol.

### 7.3 Offline Examination CLI
**Scenario: AUD-BDD-03 | NBE Regulator Offline Audit**
* **Given** an NBE examiner receives the exported `Examiner Pack` ZIP file via secure courier
* **When** they run `./medhen-auditor-cli verify --archive=data.parquet --proofs=proofs.json --root-hash=XYZ` on an air-gapped machine
* **Then** the CLI mathematically proves that every record in the Parquet file inherently rolls up to the publicly known root hash XYZ, certifying the data is 100% genuine.

---

## 8. Data Ownership & Persistence Schema

### 8.1 PostgreSQL Hot Ledger (Merkle Tree Enabled)
```sql
CREATE TABLE audit_merkle_leaves (
    seq_id BIGINT PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    leaf_hash VARCHAR(64) NOT NULL,
    digital_signature BYTEA NOT NULL,
    trace_id VARCHAR(100),
    delta_payload BYTEA
);

CREATE TABLE audit_merkle_roots (
    epoch_id BIGINT PRIMARY KEY,
    root_hash VARCHAR(64) NOT NULL,
    published_to_notary BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

---

## 9. Integration & Dependency Contracts

| External Service | Contract | Resilience / Fallback |
|:---|:---|:---|
| **KMS (Key Management)**| gRPC `VerifySignature` | Critical for Zero-Trust. Ingestion buffers if KMS is offline. |
| **Notary API (NBE)** | REST / Blockchain API | Async cron job. Retries until successful. |
| **ML Security Pipeline** | `platform.security.anomalies.v1` | Fire-and-forget streaming. Audit does not block if ML is slow. |

---

## 13. Engineering Definition of Done (DoD)

1. **Merkle Math Validated:** Unit tests must prove that generating an inclusion proof and verifying it off-ledger using `medhen-auditor-cli` succeeds mathematically.
2. **Signature Rejection:** Integration tests must simulate Kafka MITM attacks, ensuring payloads with invalid signatures are dropped.
3. **ML Streaming Validated:** The system must demonstrate <50ms latency between an event arriving and an anomaly event being published to the security topic.
