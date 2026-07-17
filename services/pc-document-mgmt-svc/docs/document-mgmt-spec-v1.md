# svc-08: Document Management Engine Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-08-v1 |
| **Service ID** | `svc-08` |
| **Service Name** | Document Management Engine |
| **Bounded Context** | `BC-MDH-08` — Document Management |
| **Version** | 1.3 |
| **Status** | Draft — Under Stakeholder Review |
| **Date** | 2026-07-17 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-0 |
| **Deploy Mode** | Microservice (`pc-document-mgmt-svc`) |
| **Target Repo** | `Platform Core/dev/pc-document-mgmt-svc` |
| **Phase** | Phase 1 & 2 |
| **PRD Anchor** | [Platform Core PRD](../prd/Medhen-Platform-PRD.md) (`REQ-DOC-*`) |
| **Capability Anchor** | [Capability Doc BC-MDH-08](../prd/Medhen-Platform-Capability-Document.md#bc-mdh-08--document-management-pc-document-svc) |
| **Capabilities** | `CAP-DOC-001` to `CAP-DOC-A3` |
| **Methodologies** | DDD · Hexagonal · EDA · CQRS-lite · Transactional Outbox |
| **Companion Specs** | `svc-03` Policy Administration · `svc-06` Claims Management · `svc-07` Billing & Payments |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-16 | Initial baseline specification. |
| 1.1 | 2026-07-17 | Escalated to Tier-0 standard. Transitioned object storage engine from MinIO to Apache Ozone for billion-object scale. Greatly expanded architectural depth. |
| 1.2 | 2026-07-17 | Elevated advanced capabilities (QR verification, E-signature, IDP) into Phase 2 scope and expanded detailed functional requirements. |
| 1.3 | 2026-07-17 | Upgraded to international standards: Responsive HTML5 output, Legal Hold workflows, and Automated PII Redaction. |

---

## Document Structure Overview

1. **Service Overview**
2. **Enterprise Architecture & Technology Stack**
3. **Functional Requirements (Tier-0 Density)**
4. **Domain Model & Events (Tactical DDD)**
5. **API & Interface Specifications**
6. **Event Schemas & Contracts (Avro)**
7. **Behaviour-Driven Scenarios (BDD)**
8. **Data Ownership, Persistence & Apache Ozone Storage**
9. **Integration & Dependency Contracts**
10. **Non-Functional Requirements & SLOs**
11. **Observability, Tracing & Auditing**
12. **Operational Runbooks & Resilience Scenarios**
13. **Engineering Definition of Done (DoD)**

---

## 1. Service Overview

### 1.1 Mission Statement

`svc-08` Document Management Engine (`BC-MDH-08`) is the **mission-critical, centralised, secure, and bilingual document processing and storage facility** powering the Medhen Platform. It operates as the sole authority for generating, storing, cryptographically signing, and retrieving every document artifact the platform produces. This includes policy schedules, certificates of insurance, statutory motor stickers, endorsements, renewal/cancellation notices, quotes, debit/credit notes, invoices/receipts, and claims correspondence. 

Operating at Tier-0, `svc-08` guarantees absolute reliability. Failure of this service blocks policy issuance, claims settlement, and regulatory reporting.

The engine owns the following strict responsibilities:
1. **High-Fidelity Document Generation** — Consuming standard JSON envelopes from upstream engines (Policy, Billing, Claims) and merging them into strictly versioned, CSS-Paged-Media compliant HTML templates to produce immutable, print-ready PDFs natively.
2. **First-Class Bilingual Rendering** — Enforcing English and Amharic (Noto Sans Ethiopic) bilingual typography natively. The service guarantees exact spatial layout parity across languages to satisfy Ethiopian localization demands without compromising structural integrity.
3. **Billion-Object Storage via Apache Ozone** — Secure, metadata-rich persistence of WORM (Write Once, Read Many) files. Moving beyond traditional S3 limits, the service utilizes Apache Ozone to handle massive namespace volumes with automated 7-year statutory retention policies.
4. **Cryptographic Verifiability** — Generating and embedding mathematically signed QR codes into statutory documents for offline verification.
5. **Malware Quarantine Network** — Integrating with an asynchronous ICAP/ClamAV pipeline to isolate and scan all customer-uploaded binaries.
6. **E-Signature Workflow** — Orchestrating secure, legally binding electronic signatures on proposals, schedules, and agreements natively within the platform.
7. **Intelligent Document Processing (IDP)** — Extracting structured data from unstructured physical uploads (e.g. driving licenses, claim evidence) via OCR and ML pipelines.

### 1.2 Product Context (Must-Include)

`svc-08` is inherently **product-agnostic**. The core transactional boundary strictly isolates document generation from insurance business logic. New lines of business (LOBs) interact with the service purely by registering new HTML templates and emitting JSON payloads map. 

### 1.3 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem** | Traditional core PAS tightly couple document generation with business logic. Furthermore, managing external e-signature tools and manual data extraction processes introduces heavy operational friction and security risks. |
| **Value** | `svc-08` decouples presentation from logic, integrates secure internal e-signature flows, and utilizes IDP for automated claims/KYC processing. This drives down operational overhead significantly. |
| **Stakeholders** | Product Managers (wordings), Underwriters (schedules), Claims Adjusters (IDP evidence processing), Legal & Compliance (e-signatures, 7-year retention auditing), End Customers. |

### 1.4 Business Capabilities Delivered

| Capability (CAP) | Description | Phase / Priority |
|:---|:---|:---|
| `CAP-DOC-001..011` | Document generation catalog (Schedules, COI, Cover Note, Motor Sticker, Endorsements, Quotes, Invoices, Claims Letters) | **Phase 1 (P0)** |
| `CAP-DOC-012` | Bilingual rendering engine (English + Amharic) | **Phase 1 (P0)** |
| `CAP-DOC-013` | Template lifecycle & CRUD via admin portal (HTML/CSS + merge fields) | **Phase 1 (P0)** |
| `CAP-DOC-014` | Tier-0 Secure Storage (Generation & Uploads) in Apache Ozone with metadata | **Phase 1 (P0)** |
| `CAP-DOC-A3` | QR-verifiable certificates for third-party verification | **Phase 2 (P1)** |
| `CAP-DOC-A1` | E-signature workflow integration & evidence tracking | **Phase 2 (P1)** |
| `CAP-DOC-A2` | Intelligent Document Processing (IDP) for KYC/Claims data extraction | **Phase 2 (P1)** |

### 1.5 In-Scope / Out-of-Scope Responsibilities

**In-Scope:**
* Orchestrating headless Chromium pools for high-throughput PDF rendering.
* Handling bilingual typography, dynamic font embedding, and pagination constraints.
* Centralised secure storage (via Apache Ozone S3-gateway) of generated and uploaded blobs.
* Serving document blobs to authorized clients/APIs via RBAC/ABAC token validation.
* Triggering ICAP anti-virus scans for user uploads.
* Managing e-signature state machines (Pending, Signed, Declined).
* Orchestrating IDP pipelines for data extraction and OCR text indexing.
* Emitting immutable `platform.document.*` events to the Kafka backbone.

**Out-of-Scope:**
* Delivering documents to customers via email/SMS (strictly owned by `BC-MDH-10` Notifications).
* Determining *if* or *when* a document should be generated.
* The ML/AI model training for IDP (the engine just calls the external/internal ML API).

---

## 2. Enterprise Architecture & Technology Stack

### 2.0 Operations-Plane Architecture Narrative

`svc-08` operates as a high-throughput, burst-tolerant data plane service. It handles asynchronous burst workloads alongside synchronous on-demand requests. 

To guarantee Tier-0 stability, the architecture enforces strict resource isolation:
1. **API Tier:** Lightweight Go routines handle ingress, metadata routing, and token validation.
2. **Rendering Pool:** CPU and memory-intensive PDF rendering is offloaded to a bounded pool of headless Chromium sidecars. 
3. **Storage Tier:** Object storage is externalised entirely to **Apache Ozone**.
4. **Processing Pipeline (Phase 2):** E-signature coordination and IDP extraction are handled via async worker pools interacting with cryptographic signers and OCR sidecars.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / runtime | **Go 1.26.x** | Fast startup, GC optimizations, concurrent multiplexing. |
| PDF Render Engine | **Chromium (Headless) + Puppeteer-Go** | Absolute pixel-perfect rendering of HTML5, CSS Grid/Flexbox, and native TTF embedding (Ethiopic fonts). |
| Primary Meta Store | **PostgreSQL 18.x** | Transactional metadata, state machines, e-signature tracking. |
| Object Store | **Apache Ozone** | Highly scalable, distributed object store accessed via Ozone S3 Gateway. |
| Event backbone | **Kafka** + **Avro** | Durable async commands and state events. |
| Malware Scanning | **ICAP / ClamAV Server** | Asynchronous streaming malware analysis. |
| IDP Engine (Ph 2) | **Tesseract OCR / Google DocAI Bridge**| Handles structured extraction from uploaded JPEGs/PDFs. |
| E-Sign PKI (Ph 2) | **Vault PKI / Custom KMS Integrations**| Signing hash values securely for platform-native signatures. |

---

## 3. Functional Requirements (Tier-0 Density)

### 3.1 Functional Requirement Catalog

#### 3.1.1 Document Generation (`FR-DOC-GEN-*`)

- **FR-DOC-GEN-1 — Event-Driven Orchestration.** The service SHALL consume domain events and accept gRPC commands (`GenerateDocument`) to asynchronously orchestrate PDF generation.
- **FR-DOC-GEN-2 — Typography & Pagination.** The service SHALL render documents utilizing CSS Paged Media. It MUST correctly embed required fonts (Noto Sans Ethiopic) to guarantee cross-device PDF portability.
- **FR-DOC-GEN-3 — Template Data Binding.** The service SHALL execute a template merge using JSON-schema payloads against the active `DocumentTemplate` version before rendering.
- **FR-DOC-GEN-4 — WORM Immutability.** A generated PDF SHALL be mathematically hashed (SHA-256) upon creation.

#### 3.1.2 Storage & Retrieval (`FR-DOC-STO-*`)

- **FR-DOC-STO-1 — Streaming Uploads.** The service SHALL expose `POST /v1/documents` for multipart streaming uploads up to 100MB, streaming directly to Apache Ozone.
- **FR-DOC-STO-2 — Relational Metadata.** Every file SHALL be indexed in PostgreSQL with `tenant_id`, `entity_type`, `entity_id`, and `document_type`.
- **FR-DOC-STO-3 — ICAP Malware Quarantine.** The service SHALL intercept user-uploaded documents and stream them through an ICAP server (ClamAV).

#### 3.1.3 Phase 2 Capabilities: E-Signature (`FR-DOC-ESIG-*`)

- **FR-DOC-ESIG-1 — Signature Orchestration.** The service SHALL expose a workflow to request electronic signatures on `ACTIVE` generated documents. It must track the signature state (`PENDING`, `SIGNED`, `DECLINED`) and record the IP address, timestamp, and user agent of the signatory.
- **FR-DOC-ESIG-2 — Cryptographic Sealing.** Upon completion of a signature workflow, the service SHALL apply a cryptographic digital signature (using the platform's root CA/KMS) to the PDF itself, rendering any future tampering immediately detectable by standard PDF readers.
- **FR-DOC-ESIG-3 — Audit Trail.** The service SHALL append an unalterable audit page to the end of the signed PDF, containing the complete timeline of the signature event.

#### 3.1.4 Phase 2 Capabilities: Intelligent Document Processing (`FR-DOC-IDP-*`)

- **FR-DOC-IDP-1 — OCR Triggering.** The service SHALL allow specific `document_type` uploads (e.g., `DRIVING_LICENSE`, `CLAIM_INVOICE`) to automatically trigger the IDP asynchronous pipeline after AV verification.
- **FR-DOC-IDP-2 — Structured Data Extraction.** The IDP pipeline SHALL extract structured key-value pairs (e.g., Name, ID Number, Expiry Date, Invoice Amount) and store this JSON structure alongside the document metadata.
- **FR-DOC-IDP-3 — Confidence Scoring.** The service SHALL record a confidence score for extracted data. Fields below the `idp.confidence_threshold` MUST be flagged for manual review via a domain event.

#### 3.1.5 Phase 2 Capabilities: QR Verification (`FR-DOC-QR-*`)

- **FR-DOC-QR-1 — Statutory QR Watermarking.** For generated Certificates of Insurance and Motor Stickers, the service SHALL dynamically generate a QR code containing an Ed25519-signed JWT payload with core policy details and a verification portal URL, embedding it visually into the PDF.

#### 3.1.6 Advanced Compliance & Output (`FR-DOC-COMP-*`)

- **FR-DOC-COMP-1 — Legal Hold Workflow.** The service SHALL allow an authorized user to apply a `LEGAL_HOLD` status to a DocumentRecord. A document on Legal Hold MUST override all automated retention policies and cannot be deleted or archived.
- **FR-DOC-COMP-2 — Omnichannel Output.** The service SHALL generate and store both a print-ready PDF and a responsive HTML5 payload (WCAG 2.1 AA compliant) for every generated document, allowing native rendering across all devices.
- **FR-DOC-COMP-3 — Automated PII Redaction.** The service SHALL expose a retrieval endpoint that streams the requested document through an automated PII redactor (masking IDs, emails, phone numbers) before transmitting it to untrusted third-party clients.

### 3.2 State Machine Definitions

**Document Uploads:**
| From State | Trigger Action | To State | Guards & Preconditions |
|:---|:---|:---|:---|
| `—` | `UploadInitiated` | `PENDING_SCAN` | File < 100MB, valid MIME type header. |
| `PENDING_SCAN` | `IcapScanResult (CLEAN)` | `VERIFIED` | ICAP server returns HTTP 200/204. |
| `PENDING_SCAN` | `IcapScanResult (THREAT)`| `QUARANTINED` | ICAP server detects signature match. |
| `VERIFIED` | `IDPTriggered` | `PROCESSING_IDP` | Only if `document_type` is IDP-eligible. |
| `PROCESSING_IDP` | `IDPCompleted` | `ACTIVE` | Extraction results stored. |
| `VERIFIED` | `EntityLink` | `ACTIVE` | Tied to an active Claim or Party record. |

**E-Signature Workflow:**
| From State | Trigger Action | To State | Guards & Preconditions |
|:---|:---|:---|:---|
| `—` | `RequestSignature` | `SIGNATURE_PENDING` | Document is generated and `ACTIVE`. |
| `SIGNATURE_PENDING` | `Sign` | `SIGNED` | Signatory identity validated via token. |
| `SIGNATURE_PENDING` | `Decline` | `SIGNATURE_DECLINED`| — |

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Aggregate Roots

| Aggregate Root | Definition & Invariants | Emitted Events |
|:---|:---|:---|
| **`DocumentTemplate`** | A versioned HTML/CSS layout mapped to a specific `document_type` and `locale`. | `DocumentTemplateCreated`<br>`DocumentTemplateUpdated` |
| **`DocumentRecord`** | Represents the metadata, lifecycle state, signature status, and physical storage pointer for a blob in Apache Ozone. | `DocumentGenerated`<br>`DocumentUploaded`<br>`DocumentQuarantined` |
| **`SignatureRequest`** | Tracks the lifecycle of an e-signature request bound to a `DocumentRecord`. | `SignatureRequested`<br>`DocumentSigned`<br>`SignatureDeclined` |

### 4.2 Entities vs Value Objects

| Concept | Type | Justification |
|:---|:---|:---|
| `DocumentRecord` | Entity (AR) | Has identity, tracks state transitions. |
| `SignatureRequest` | Entity (AR) | Manages independent state for signing processes. |
| `StorageRef` | Value Object | Encapsulates the Ozone volume, bucket, key, and MIME type. |
| `ExtractionResult` | Value Object | Encapsulates the parsed JSON from the IDP pipeline and confidence scores. |

---

## 5. API & Interface Specifications

Base path: `/api/pc-document-mgmt/v1`

### 5.1 REST API (Sync Operations & Uploads)

| Method | Endpoint | Purpose | AuthZ Policy |
|:---|:---|:---|:---|
| `POST` | `/documents/upload` | Multipart file upload (streaming). | `document.upload` |
| `GET` | `/documents/{id}` | Download file (returns binary stream). | `document.read` + ABAC owner |
| `GET` | `/documents/{id}/extracted-data` | Retrieve IDP JSON results. | `document.read` + ABAC owner |
| `POST` | `/documents/{id}/signature/request`| Initiate an e-signature workflow. | `document.sign_admin` |
| `POST` | `/documents/{id}/signature/sign` | Sign the document (applies seal). | `document.sign` + Signatory Auth |

### 5.2 gRPC API (Internal Commands & Fast Reads)

**Service Definition:** `medhen.platform.document.v1.DocumentService`

| RPC | Request | Response | SLA (P95) |
|:---|:---|:---|:---|
| `GenerateDocument` | `template_code`, `locale`, `entity_type`, `entity_id`, `payload_json` | `GenerateResponse(document_id)` | Acceptance < 50ms (Async) |
| `GetDocumentMetadata`| `document_id` | `DocumentRecordSnapshot` | < 10ms |

---

## 6. Event Schemas & Contracts (Avro)

All domain events are published to Kafka via the transactional outbox pattern.

### 6.1 Topic Mapping

| Event | Topic | Partition Key |
|:---|:---|:---|
| `DocumentGenerated` | `platform.document.generated.v1` | `tenant_id:entity_id` |
| `DocumentUploaded` | `platform.document.uploaded.v1` | `tenant_id:entity_id` |
| `DocumentSigned` | `platform.document.signature.signed.v1`| `tenant_id:document_id` |
| `IDPExtractionComplete`| `platform.document.idp.completed.v1`| `tenant_id:document_id` |

---

## 7. Behaviour-Driven Scenarios (BDD)

**Scenario: DOC-BDD-01 | Bilingual Policy Schedule Generation**
* **Given** a valid `MotorSchedule` template exists for locales `en` and `am`
* **When** `pc-policy-svc` emits a `PolicyIssued` event with `preferred_locale="am"`
* **Then** `pc-document-mgmt-svc` consumes the event and triggers generation using the Amharic template
* **And** the Chromium engine renders a PDF correctly embedding Noto Sans Ethiopic ligatures.

**Scenario: DOC-BDD-03 | E-Signature Completion & Sealing**
* **Given** a `SignatureRequest` is `SIGNATURE_PENDING` on a Quote PDF
* **When** the customer triggers `POST /signature/sign` with valid OAuth scopes
* **Then** the service transitions the request to `SIGNED`
* **And** cryptographically seals the PDF using the platform's internal PKI
* **And** emits a `DocumentSigned` event.

**Scenario: DOC-BDD-04 | IDP Data Extraction on Driving License**
* **Given** a verified uploaded document of type `DRIVING_LICENSE`
* **When** the IDP async pipeline completes processing
* **Then** the service attaches the `ExtractionResult` JSON to the document metadata
* **And** emits an `IDPExtractionComplete` event containing the parsed Name, License Number, and Date of Birth
* **And** flags fields with confidence < 80% for manual review.

---

## 8. Data Ownership, Persistence & Apache Ozone Storage

### 8.1 PostgreSQL DDL (Metadata)

```sql
CREATE TABLE document_records (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    locale VARCHAR(10),
    status VARCHAR(30) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    sha256_hash VARCHAR(64) NOT NULL,
    storage_volume VARCHAR(100) NOT NULL,
    storage_bucket VARCHAR(100) NOT NULL,
    storage_path TEXT NOT NULL,
    html_storage_path TEXT,           -- Responsive HTML5 output
    is_legal_hold BOOLEAN DEFAULT FALSE,
    idp_extracted_data JSONB,         -- Phase 2: IDP Output
    signature_status VARCHAR(30),     -- Phase 2: PENDING, SIGNED
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

### 8.2 Apache Ozone Configuration

`svc-08` interacts via the S3 Gateway but organizes data hierarchically:
* **Ozone Volume:** `medhen-tenant-{tenant_id}`
* **Ozone Bucket:** `active-documents` vs `quarantine-documents`

---

## 9. Integration & Dependency Contracts

| Service | Contract | Coupling | Resilience |
|:---|:---|:---|:---|
| **Apache Ozone (S3g)** | S3 API for blob storage | Sync (I/O) | Uploads fail with HTTP 503; async generation queues in Kafka. |
| **Chromium Render Pool** | Local IPC / Sidecar HTTP | Sync | Bounded pool prevents pod OOM. |
| **ClamAV ICAP Server** | TCP Streaming | Sync (Stream)| Uploads fail-closed (reject) if AV is unreachable. |
| **Vault PKI (Ph 2)** | Document signing keys | Sync | Retries via gRPC fallback. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | SLO | Consequence of Breach |
|:---|:---|:---|
| **Availability (Upload/Download)** | 99.99% | Operations blocked. |
| **Generation Latency (Async)** | P95 < 15s | Batch renewal notification delays. |
| **IDP Processing Latency (Ph 2)** | P95 < 30s | Slower claims ingestion processing. |

---

## 11. Observability, Tracing & Auditing

- **Traffic:** `grpc_server_handled_total{method="GenerateDocument"}`
- **Storage:** `ozone_upload_duration_seconds`
- **IDP/AI:** `idp_extraction_latency_seconds`, `idp_low_confidence_flags_total`
- **Signatures:** `signature_events_total{status="completed"}`

---

## 12. Operational Runbooks & Resilience Scenarios

### 12.1 IDP Pipeline Saturated

**Symptom:** High `idp_extraction_latency_seconds` during severe storm events (high claims uploads).
**Action:** IDP extraction queues are persistent. If latency is unacceptable, scale the IDP worker sidecars: `kubectl scale deploy pc-document-mgmt-idp-worker --replicas=30`.

---

## 13. Engineering Definition of Done (DoD)

1. **Rendering Fidelity:** Visual regression tests must prove identical layout rendering for English and Amharic text utilizing Noto Sans Ethiopic.
2. **Security & Malware:** Integration tests must upload the EICAR test file and successfully quarantine it.
3. **E-Signature Legal Validity:** Signed PDFs must pass Adobe Acrobat's native PAdES (PDF Advanced Electronic Signatures) verification showing the seal is valid and the document hasn't been altered post-signature.
4. **Resilience & Idempotency:** If Apache Ozone is abruptly disconnected mid-upload, the request must buffer safely in Kafka and retry without creating orphaned metadata.
