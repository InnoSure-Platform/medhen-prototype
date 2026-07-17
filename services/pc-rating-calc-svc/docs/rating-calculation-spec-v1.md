# svc-11: Rating & Premium Calculation Engine Specification (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-SVC-SPEC-PC-11-v1 |
| **Service ID** | `svc-11` |
| **Service Name** | Rating & Premium Calculation Engine |
| **Bounded Context** | `BC-MDH-04` — Rating & Premium Calculation |
| **Version** | 1.0 |
| **Status** | Draft |
| **Date** | 2026-07-17 |
| **Classification** | Internal — Confidential |
| **Tier** | Tier-0 |
| **Deploy Mode** | Microservice (`pc-rating-calc-svc`) |
| **Target Repo** | `Platform Core/dev/services/pc-rating-calc-svc` |
| **Phase** | Phase 1 (Core) |
| **PRD Anchor** | [Platform Core PRD](../../../../../../docs/prd/Medhen-Platform-PRD.md) (`REQ-RAT-*`) |
| **Capability Anchor** | [Capability Doc BC-MDH-04](../../../../../../docs/prd/Medhen-Platform-Capability-Document.md#bc-mdh-04--rating--premium-calculation-pc-rating-calc-svc) |
| **Capabilities** | `CAP-RATE-001` to `CAP-RATE-A3` |
| **Methodologies** | DDD · Hexagonal · Stateless Compute · CQRS-lite |
| **Companion Specs** | `svc-10` Product Definition · `svc-13` Policy Management |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-17 | Initial Tier-0 specification covering Sections 1-13. Drafted against PRD capabilities (`REQ-RAT-001` through `050`). Radically expanded for comprehensive implementation clarity, aligning with Tier-0 architectural standards. |

---

## Document Structure Overview

1. **Service Overview**
2. **Technology Stack & Architecture**
3. **Functional Requirements & Mathematical Pipelines**
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

`svc-11` Rating & Premium Calculation Engine (`BC-MDH-04`) is the **absolute pricing authority** for the Medhen Platform. It is a highly optimized, state-free computational engine that receives a risk profile and an effective product definition, processes it through dynamic factor matrices, and outputs a mathematically precise, fully itemized premium breakdown.

The engine strictly separates pricing *logic execution* from pricing *configuration* (owned by `svc-10` Product Definition). This decoupling enables massive horizontal scalability, ultra-low latency (< 10ms) quoting, and perfectly reproducible calculation audit trails without tethering the hot-path to relational database bottlenecks.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem Space** | Legacy systems tightly couple rating rules with policy administration. This leads to slow quoting speeds, floating-point rounding errors on complex pro-rata cancellations, and a total inability to audit exactly *why* a specific premium was charged two years ago. |
| **Value Delivered** | Provides an auditable, horizontally scalable pricing core. Ensures 100% financial correctness using arbitrary-precision decimals, enforces Ethiopian regulatory tax laws strictly, and enables actuaries to test hypothetical rate changes via a simulated "What-If" engine. |
| **Stakeholders** | Actuaries, Underwriters, Finance (for VAT/Stamp Duty reconciliation), Auditors & NBE (for regulatory defensibility), Brokers (fast quoting). |

### 1.3 Business Capabilities Delivered

| Capability (CAP) | Description | Primary REQ | Phase |
|:---|:---|:---|:---|
| `CAP-RATE-001` | **Premium Calculation Pipeline:** Execution of base-rate lookup, additive/multiplicative factor bindings, NCD (No Claim Discount) application, and bounds enforcement. | `REQ-RAT-001` | 1 |
| `CAP-RATE-002` | **Rating Audit & Transparency:** Emission of an immutable audit trail (`RatingCalculation`) detailing exactly which table versions and factors were applied, satisfying NBE regulatory defensibility. | `REQ-RAT-002` | 1 |
| `CAP-RATE-003` | **Pro-Rata Computations:** Mathematical sub-routines for policy lifecycle events (Additional/Return Premium on endorsements, Short-rate and Pro-rata mid-term cancellations). | `REQ-RAT-003` | 1 |
| `CAP-RATE-004` | **Tax & Regulatory Computations:** Hard-coded financial routing for Ethiopian VAT (15%), Stamp Duty, and specific local levies. | `REQ-RAT-010` | 1 |
| `CAP-RATE-A1` | **Telematics / Usage-Based:** API extensions to consume and aggregate telematics driving scores as dynamic risk modifiers. | `REQ-RAT-020` | 4 |
| `CAP-RATE-A2` | **ML-Assisted Pricing:** Integration with ML inference endpoints to append non-linear risk factors alongside traditional actuarial tables. | `REQ-RAT-021` | 4 |
| `CAP-RATE-A3` | **Multi-LOB Bundle Pricing:** Cross-referencing `BC-MDH-01` (Party) to apply dynamic discounts when a customer holds multiple active policies across Motor and Life. | `REQ-RAT-022` | 3 |

### 1.4 Bounded Context Responsibilities (`BC-MDH-04`)

| Owns | Exposes | Produces (via Async) | Invariants |
|:---|:---|:---|:---|
| `RatingEngine` execution logic | High-throughput gRPC `CalculatePremium` | `pc.rating.calculated.v1` | Calculations must be 100% deterministic (`INV-RAT-1`) |
| Decimal math and Rounding rules | REST `simulate` API (What-If) | Audit trace logs | Rounding to minor currency units ONLY occurs at the final step (`INV-RAT-2`) |
| Pro-Rata temporal logic | — | — | Cannot mutate risk state; it is a pure function (`INV-RAT-3`) |

### 1.5 Context Map

```mermaid
flowchart TB
    subgraph Core["svc-11 Rating Calc (BC-MDH-04)"]
        ENG["Calculation Engine Pipeline"]
        CACHE[("Redis Rate Cache")]
        MATH["Precision Math Subsystem"]
    end

    subgraph Upstream["Providers (Data Sources)"]
        PROD["pc-product-defn-svc (BC-MDH-02)"]
    end

    subgraph Downstream["Consumers (Execution)"]
        POL["pc-policy-svc (BC-MDH-03)"]
        UI["Actuarial Portal (What-If)"]
    end
    
    subgraph Telemetry["Observability & Audit"]
        AUD["pc-audit-svc (BC-MDH-17)"]
    end

    POL -->|CalculatePremium (gRPC)| ENG
    UI -->|SimulatePremium (REST)| ENG
    ENG -->|gRPC GetRateTable (Sync Fallback)| PROD
    PROD -->|pc.product.published.v1 (Kafka)| CACHE
    ENG -->|pc.rating.calculated.v1 (Kafka)| AUD
    ENG <--> MATH
```

---

## 2. Technology Stack & Architecture

### 2.0 Operations-Plane Architecture Narrative

Because `svc-11` sits directly in the hot-path of every quote, endorsement, and renewal generated by the platform, its architecture prioritizes **raw CPU throughput, zero I/O blocking, and extreme memory safety**.

Unlike other Tier-0 services, `svc-11` **does not have a primary PostgreSQL database**. It is a stateless functional engine.
1. **The Cache Dependency:** The engine relies on a Redis cache populated with active product rate tables. To ensure sub-10ms latency, it maintains an in-memory `LRU` cache of these tables, falling back to Redis, and only falling back to `svc-10` over gRPC on a total cache miss.
2. **The Math Engine:** Financial correctness is paramount. Standard `float64` inevitably causes rounding drifts over chained multiplicative factors. The core utilizes `shopspring/decimal` (arbitrary-precision fixed-point math). All intermediate calculations retain high precision (e.g., 6 decimal places), with bank-standard `HalfEven` rounding applied strictly at the boundary before constructing the final Gross Premium payload.
3. **The Audit Trail:** Because there is no local database, the engine cannot use the Transactional Outbox pattern. Instead, upon completing a calculation, it asynchronously fires a `RatingCalculatedEvent` to Kafka. This event contains the fully materialized calculation graph (inputs, factors applied, outputs) for long-term storage in `pc-audit-svc`.

### 2.1 Technology Selection

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language / runtime | **Go 1.26.x** | CPU efficiency, fast startup, zero-cost abstractions for heavy math pipelines. |
| Core API (Internal) | **gRPC / Protobuf** | Strict type safety and low-overhead binary serialization for inter-service calls. |
| Core API (External) | **REST / JSON** | Exposing `simulate` capabilities for actuarial portals. |
| Mathematics Library | **`shopspring/decimal`** | Eliminates IEEE 754 floating-point inaccuracies natively found in Go `float64`. |
| Local State / Cache | **Redis 7.x + BigCache** | Two-tier caching (In-process BigCache -> Network Redis) for rate tables. |
| Event Publishing | **Kafka + Avro** | Async telemetry and audit trails via `confluent-kafka-go`. |

### 2.2 Core Configuration References

| Config Key | Default | Purpose |
|:---|:---|:---|
| `math.precision.internal` | `6` | Decimal places retained during intermediate factor multiplications. |
| `math.precision.final` | `2` | Currency minor unit (e.g., Cents) rounding boundary for Ethiopian Birr (ETB). |
| `cache.lru.ttl_seconds` | `300` | In-process cache TTL. Flushed earlier if Kafka invalidation event arrives. |
| `audit.kafka_producer_acks` | `1` | Leader-ack mode. Ensures high throughput over strict durability for audit traces. |

---

## 3. Functional Requirements & Mathematical Pipelines

### 3.1 Detailed Requirement Catalog (RFC 2119)

#### 3.1.1 Premium Calculation Pipeline (`FR-RAT-CALC-*`)
- **FR-RAT-CALC-1 — Orchestration:** The service SHALL execute the following sequential calculation graph for every requested coverage:
  1. **Base Rate Extraction:** Lookup base premium from `RateTable` using primary risk dimensions.
  2. **Multiplicative Modifiers:** Apply sequential factor multiplications (e.g., Age Factor, Vehicle Type Factor).
  3. **Additive Modifiers:** Apply flat loading fees (e.g., High-Risk Geography + 500 ETB).
  4. **Discount Binding:** Apply NCD (No Claims Discount) percentages and Loyalty discounts.
  5. **Limits Check:** Enforce Minimum / Maximum premium bounds for the coverage.
  6. **Tax Aggregation:** Calculate VAT (15%) and Stamp Duty on the derived Net Premium.
  7. **Gross Assembly:** Sum all coverages into a final Gross Premium payload.
- **FR-RAT-CALC-2 — Multi-Dimensional Lookup:** The service SHALL support dynamic n-dimensional matrix lookups. If a requested dimension (e.g., `driver_age`) is missing from the `RatingRequest` but required by the cached `RateTable`, the service MUST abort and return `400 Bad Request / RAT-1001`.
- **FR-RAT-CALC-3 — Bounds Handling:** If a calculated sub-premium falls below the configured `min_premium` for that product coverage, the system SHALL artificially bind the premium to the minimum and log a `BOUNDS_APPLIED` audit trace.

#### 3.1.2 Pro-Rata & Endorsements (`FR-RAT-PRO-*`)
- **FR-RAT-PRO-1 — Pro-Rata Math:** The service SHALL calculate fractional time consumed. It MUST support `ACTUAL/365` (Exact days) and `30/360` (Financial standard) calendar methods based on product configuration.
- **FR-RAT-PRO-2 — Delta Computation:** For mid-term endorsements, the service SHALL accept a `prior_premium_breakdown` and a `new_risk_profile`. It will calculate the new annualized premium, prorate the difference for the remaining term days, and return the `DeltaPremium` (Additional or Return).
- **FR-RAT-PRO-3 — Short-Rate Penalties:** Upon cancellation, the service SHALL apply configured short-rate penalty scales (e.g., retaining 20% of the unearned premium as administrative overhead) before yielding the final `ReturnPremium`.

#### 3.1.3 Simulation & What-If Analysis (`FR-RAT-SIM-*`)
- **FR-RAT-SIM-1 — Overridden Matrices:** The `/api/pc-rating/v1/simulate` endpoint SHALL accept an embedded, temporary `RateTable` matrix within the request body. The engine MUST use this injected matrix instead of fetching from the cache, enabling actuaries to test price elasticity without mutating live configurations.

### 3.2 State Machine
Because `BC-MDH-04` is purely stateless, it does not manage entity lifecycles. Its "state machine" is purely the sequential transition of the `RatingCalculation` pipeline.

---

## 4. Domain Model & Events (Tactical DDD)

### 4.1 Bounded Context Boundary (`BC-MDH-04`)
The Rating context is entirely functional. It receives inputs (Risk + Policy State), fetches context (Product Rates), applies the domain logic (Mathematical Pipeline), and yields outputs (Premium + Audit Trace).

### 4.2 Core Domain Concepts (Value Objects)

| Concept | Type | Justification |
|:---|:---|:---|
| `RatingRequest` | Value Object | Ephemeral command. Contains `product_id`, `as_of_date`, and dynamic `risk_dimensions` map. |
| `PremiumBreakdown`| Value Object | The finalized, immutable output structure. |
| `CalculationTrace`| Value Object | A directed acyclic graph (DAG) representing the exact sequence of mathematical operations performed. |
| `ProRataFraction` | Value Object | E.g., `145/365`. Encapsulates time-delta logic. |

### 4.3 Domain Services

| Service | Responsibility |
|:---|:---|
| **`CalculationPipeline`** | Orchestrates the extraction -> modification -> bounds -> tax phases. |
| **`MatrixEvaluator`** | Navigates the cached JSONB `RateTable` dimensions to find the precise numerical factor for a given risk key. |
| **`TaxEngine`** | Applies jurisdiction-specific tax rules (e.g., Ethiopian VAT is calculated on (Net Premium + Administrative Loadings) but NOT on Stamp Duty). |

### 4.4 Exception Handling (Domain Exceptions)

| Domain Exception | Trigger Condition | Consequence |
|:---|:---|:---|
| `MissingRiskDimension` | Required key missing from input map. | Halts pipeline, HTTP 400. |
| `RateMatrixViolation` | Input values fall outside all defined bounds in the rate table. | Halts pipeline, HTTP 422. |
| `MathOverflowException` | Decimal calculation exceeds logical bounds (e.g., negative premium without discount). | Triggers PagerDuty, HTTP 500. |

---

## 5. API Specifications (REST & gRPC)

### 5.1 gRPC API (Core Quote Path)

The gRPC API is heavily consumed by `pc-policy-svc` during Quoting, Endorsements, and Renewals.

**Service Definition:** `medhen.platform.rating.v1.RatingService`

| RPC | Request | Response | SLA (P95) |
|:---|:---|:---|:---|
| `CalculatePremium` | `CalculateRequest` (Product, Risks, Dates) | `CalculateResponse` | < 10ms |
| `CalculateProRata` | `ProRataRequest` (Prior breakdown, New risks, Term Dates) | `ProRataResponse` | < 10ms |
| `CalculateCancellation`| `CancelRequest` (Reason, Target Date, Prior breakdown) | `CancelResponse` (Return Premium) | < 10ms |

#### 5.1.1 Payload Example: `CalculateRequest` & `CalculateResponse`

```protobuf
message CalculateRequest {
  string request_id = 1;         // Used for idempotency & tracing
  string tenant_id = 2;
  string product_code = 3;
  int64 as_of_date = 4;          // Unix milliseconds for effective versioning
  
  // Dynamic risk dimensions passed from Policy svc
  map<string, string> risk_dimensions = 5; 
  repeated string selected_coverages = 6;
}

message CalculateResponse {
  string calculation_id = 1;     // UUID generated by engine
  string net_premium = 2;        // String to avoid float64 precision loss
  string total_taxes = 3;
  string gross_premium = 4;
  
  repeated CoveragePremium coverage_breakdowns = 5;
  repeated AuditStep trace_log = 6;
}

message AuditStep {
  int32 step_order = 1;
  string operation = 2;          // e.g. "MULTIPLY_AGE_FACTOR"
  string value_applied = 3;      // e.g. "1.15"
  string table_version = 4;      // e.g. "v4-2026-01-01"
}
```

### 5.2 REST API (Simulation & Actuarial)

**Base path:** `/api/pc-rating/v1`

| Method | Endpoint | Purpose | Security / RBAC |
|:---|:---|:---|:---|
| `POST` | `/simulate` | Run calculation against hypothetical rate matrices | `actuary.simulate` |
| `POST` | `/recalculate` | Re-run a historical calculation explicitly for audit purposes | `compliance.officer` |

### 5.3 Error Taxonomy (RFC 7807 Problem Details)

| Domain Exception | HTTP Code | gRPC Status | Error Code | Client Action |
|:---|:---|:---|:---|:---|
| `MissingRiskDimension` | `400 Bad Request` | `INVALID_ARGUMENT` | `RAT-1001` | Append missing data to UI quoting form. |
| `EffectiveTableNotFound` | `404 Not Found` | `NOT_FOUND` | `RAT-1002` | Validate `product_code` or `as_of_date`. |
| `MatrixViolation` | `422 Unprocessable` | `FAILED_PRECONDITION`| `RAT-1003` | Risk profile is fundamentally unratable (e.g. Age > 120). |
| `CalculationTimeout` | `504 Gateway Timeout`| `DEADLINE_EXCEEDED` | `RAT-2001` | Upstream cache/DB degraded. Retry. |

---

## 6. Event Schemas & Contracts (Avro)

`svc-11` emits an audit stream of calculations. Because it is stateless, it publishes directly to Kafka (Producer) without a database outbox. If Kafka is temporarily unreachable, events are buffered in memory and logged to disk to prevent audit loss.

### 6.1 Topic Mapping

| Event | Topic | Partition Key | Schema ID |
|:---|:---|:---|:---|
| `RatingCalculated` | `pc.rating.calculated.v1` | `tenant_id:calculation_id` | `RatingCalculatedEvent` |

### 6.2 Avro Schema: `RatingCalculatedEvent`

```json
{
  "namespace": "medhen.platform.rating.v1",
  "type": "record",
  "name": "RatingCalculatedEvent",
  "fields": [
    {"name": "event_id", "type": "string", "logicalType": "uuid"},
    {"name": "tenant_id", "type": "string"},
    {"name": "calculation_id", "type": "string", "logicalType": "uuid"},
    {"name": "product_code", "type": "string"},
    {"name": "as_of_date", "type": {"type": "long", "logicalType": "timestamp-millis"}},
    {"name": "risk_hash", "type": "string", "doc": "SHA-256 hash of the input risk dimensions for audit integrity"},
    {"name": "net_premium", "type": "string"},
    {"name": "gross_premium", "type": "string"},
    {"name": "trace_breakdown", "type": "string", "doc": "JSON serialized string of the calculation DAG"},
    {"name": "occurred_at", "type": {"type": "long", "logicalType": "timestamp-millis"}}
  ]
}
```

---

## 7. Behaviour-Driven Scenarios (BDD)

### 7.1 Standard Premium Calculation
**Scenario: RAT-BDD-01 | Execute Full Pricing Pipeline**
* **Given** an active rate table for `MOT-COMP-01`
* **And** a base premium of 1000 ETB
* **And** an Age Factor (18-25) of 1.5, and NCD of 10%
* **When** a `CalculatePremium` request arrives for a 20-year-old with NCD entitlement
* **Then** the Subtotal is 1500 ETB (1000 * 1.5)
* **And** the NCD discount reduces the Net Premium to 1350 ETB
* **And** VAT (15%) is calculated precisely as 202.50 ETB
* **And** Gross Premium is returned as 1552.50 ETB
* **And** a `RatingCalculatedEvent` is dispatched asynchronously.

### 7.2 Pro-Rata Cancellation (Short-Rate)
**Scenario: RAT-BDD-02 | Mid-Term Cancellation Penalty**
* **Given** a Policy with a Gross Premium of 3650 ETB bound for 365 days
* **And** the product defines a 10% short-rate penalty upon cancellation
* **When** a `CalculateCancellation` request is made on day 100
* **Then** the theoretical unearned premium is 2650 ETB (265 days)
* **And** the short-rate penalty deducts 265 ETB (10% of unearned)
* **And** the final `ReturnPremium` is calculated as 2385 ETB.

### 7.3 Cache Desync Resilience
**Scenario: RAT-BDD-03 | Survive Product Service Outage**
* **Given** `svc-10` Product Definition goes offline
* **When** `svc-11` receives a rating request for `MOT-COMP-01`
* **Then** it serves the calculation directly from the local BigCache/Redis tier
* **And** yields a correct result in < 5ms without returning a 5xx error.

---

## 8. Data Ownership & Persistence Schema

`svc-11` does not own a relational database schema. It utilizes a highly optimized **Redis Cache Schema** designed purely for read performance.

### 8.1 Redis Rate Cache Schema

* **Key:** `mdh:rating:cache:{tenant_id}:{product_code}:{effective_date_epoch}`
* **Type:** Hash
* **Encoding:** Protobuf Binary (for ultra-fast unmarshalling in Go)
* **Payload Structure:** Multi-dimensional maps (e.g., `map[string]map[string]decimal.Decimal`).
* **TTL:** 24 Hours.
* **Invalidation:** Triggers proactively when a `pc.product.published.v1` event is consumed.

---

## 9. Integration & Dependency Contracts

| External Service | Contract / Protocol | Coupling | Resilience / Fallback |
|:---|:---|:---|:---|
| **`pc-product-defn-svc`** | gRPC `GetEffectiveProduct` | Sync (Inbound) | If `svc-10` is unreachable, rely entirely on Redis TTL cache. |
| **`pc-policy-svc`** | gRPC `CalculatePremium` | Sync (Outbound)| N/A. `svc-11` acts as the server. |
| **`pc-audit-svc`** | Kafka `pc.rating.calculated.v1`| Async (Produce)| If Kafka is down, events buffer in local memory via `confluent-kafka-go` batching. |

---

## 10. Non-Functional Requirements & SLOs

| Metric | SLO | Consequence of Breach | Measurement / Triggers |
|:---|:---|:---|:---|
| **Availability (gRPC)** | 99.999% | Tier-0. A failure completely halts quoting, binding, and endorsement journeys across the platform. | Prometheus `grpc_server_handled_total`. Alert if < 99.99% over 5m. |
| **Latency (gRPC Reads)** | P95 < 10ms | Quoting UX degrades; downstream sagas timeout. | OpenTelemetry Span Duration. |
| **Precision Drift** | 0.00 | Financial miscalculations lead to regulatory action, un-reconcilable billing ledgers, and catastrophic P&L drift. | Continuous automated property-based testing against mathematical boundaries. |

---

## 11. Observability Specification

The service utilizes the standard `pc-telemetry-sdk` for OpenTelemetry.

### 11.1 Golden Signals (Prometheus)

- **Traffic:** `grpc_server_handled_total{method="CalculatePremium|CalculateProRata"}`
- **Latency:** `grpc_server_handling_seconds_bucket`
- **Errors:** `grpc_server_handled_total{grpc_code!="OK"}`
- **Cache Hit Ratio:** `rate_cache_hits_total` / `rate_cache_lookups_total` (Critically important to monitor; drops indicate a cache stampede).

### 11.2 Custom Domain Metrics

- `premium_calculated_gross_total` (Histogram for anomaly detection: e.g., suddenly calculating billions in premium indicates a severe factor configuration error).
- `math_precision_warnings_total` (Logs when a decimal truncation was forced).

### 11.3 Logging (Structured `slog`)

```json
{"level":"INFO","time":"2026-07-16T12:00:00Z","msg":"Calculation pipeline complete","calculation_id":"c-8822","gross_premium":"1552.50","execution_time_ms":1.2,"trace_id":"...","span_id":"..."}
```

---

## 12. Operational Runbooks

### 12.1 Cache Stampede / P95 Latency Spike
**Symptom:** Horizontal Pod Autoscaler maxes out; P95 latency jumps to 200ms; `rate_cache_hits_total` ratio drops below 50%.
**Action:**
1. This occurs if Redis fails and all `svc-11` pods simultaneously attempt to fetch rates from `svc-10` synchronously over gRPC.
2. Ensure Singleflight mechanism is active in Go (prevents multiple identical requests from hitting `svc-10` concurrently).
3. Check Redis connectivity: `kubectl exec -it pod/pc-rating-calc-svc -- redis-cli PING`.

### 12.2 Purging Corrupted Rate Caches
**Symptom:** `svc-10` pushed an incorrect rate table, and `svc-11` is quoting wrong prices despite a fix being applied upstream (Kafka invalidation event missed).
**Action:**
1. Manually flush the rating namespace in Redis:
   ```bash
   redis-cli --scan --pattern "mdh:rating:cache:*" | xargs redis-cli DEL
   ```
2. The next Quote will experience a slight delay (~50ms) as it synchronously rehydrates from `svc-10`.

---

## 13. Engineering Definition of Done (DoD)

Before `svc-11` can be deployed to the `staging` environment for Phase 1, the following quality gates MUST be passed:

1. **Precision Guarantee:** No use of `float64` anywhere in the `domain/math` package. All math MUST utilize `shopspring/decimal`. Verified via CI static analysis.
2. **Test Coverage:** Core domain rules (Math Pipeline, Tax Engine, Pro-Rating) MUST have > 98% unit test coverage.
3. **Property-Based Testing:** Execution of 10,000+ randomized fuzz tests generating edge-case risk inputs to guarantee the engine never crashes with a `panic()`.
4. **Performance Profiling:** `pprof` traces must prove that the calculation pipeline allocates < 500KB heap per request to prevent Go Garbage Collection pauses under high load.
5. **Resiliency:** Load tests via `k6` must demonstrate P95 < 10ms latency at 2,000 TPS using cached data.
