# Kernel Library Specification: pc-shared-go (v1)

| Field | Detail |
|:------|:-------|
| **Document ID** | MDH-LIB-SPEC-PC-20-v1 |
| **Library ID** | `lib-pc-kernel` |
| **Library Name** | Platform Core Kernel (Shared Go) |
| **Bounded Context** | `BC-MDH-20` — Platform Kernel |
| **Version** | 1.0 |
| **Status** | Approved |
| **Date** | 2026-07-16 |
| **Classification** | Internal |
| **Tier** | Tier-0 |
| **Deploy Mode** | Go Module (`github.com/InnoSure-Platform/pc-shared-go`) |
| **Target Repo** | `shared/go` |
| **Phase** | Phase 0 (Pilot MVP seed) |
| **PRD Anchor** | [Service Registry](../../docs/prd/service-registry.md) |
| **Capabilities** | Idempotency, Transactional Outbox, Events, Telemetry, Middlewares |

**Revision history**

| Version | Date | Summary |
|:---|:---|:---|
| 1.0 | 2026-07-16 | Initial Tier-0 specification for the Go platform kernel primitives. |

---

## Document Structure Overview

1. **Kernel Overview**
2. **Technology Stack & Core Dependencies**
3. **Core Primitives & Capabilities**
4. **Package Architecture (Tactical Design)**
5. **Integration Contracts & Middleware**
6. **Observability & Telemetry Standards**
7. **Behaviour-Driven Usage Scenarios (BDD)**
8. **Engineering Definition of Done (DoD)**

---

## 1. Kernel Overview

### 1.1 Mission Statement

`pc-shared-go` (`BC-MDH-20`) is the **foundational, product-agnostic substrate** of the Medhen Platform. It functions as the "Platform Kernel," providing standardized, highly-optimized primitives that all tier-1 and tier-2 microservices consume. It ensures that cross-cutting concerns—such as distributed tracing, robust event publishing, strict idempotency, and multi-tenant security—are implemented consistently and correctly across the entire engineering organisation.

The kernel owns the following strict responsibilities:
1. **Transactional Outbox & Eventing** — Guarantees atomic domain state mutation and domain event persistence, along with standardized Kafka publishing and Avro serialization.
2. **Idempotency Engine** — Ensures API mutations are replay-safe across distributed services.
3. **Platform Middleware** — Standardized gRPC and HTTP interceptors for logging, metrics, panic recovery, and context propagation.
4. **Domain Primitives** — Standardized structures for ETB money, Ge'ez calendar dating, and robust error handling.
5. **Multi-Tenancy** — Centralized tenant context extraction and propagation.

### 1.2 Business Context

| Aspect | Description |
|:-------|:------------|
| **Problem** | Allowing each microservice to independently implement outbox patterns, Kafka integration, and telemetry leads to architectural drift, fragile integrations, and inconsistent observability, drastically reducing operational safety. |
| **Value** | By centralizing the hardest distributed systems problems (idempotency, distributed transactions, tracing) into a rigorously tested, high-performance Tier-0 library, feature teams can focus purely on business logic (e.g., rating, quoting) while inheriting enterprise-grade reliability. |
| **Stakeholders** | Platform Engineering, SREs, Product Engineering Teams. |

---

## 2. Technology Stack & Core Dependencies

As a Tier-0 library, `pc-shared-go` strictly minimizes external dependencies to reduce transitive vulnerability risk and dependency bloat for consumer services.

| Layer | Technology | Rationale |
|:---|:---|:---|
| Language | **Go 1.24+** | Optimized for standard library features and concurrency. |
| Persistence | **pgx/v5** (`github.com/jackc/pgx/v5`) | High-performance PostgreSQL driver used by the outbox and idempotency layers. |
| Eventing | **kafka-go** (`github.com/segmentio/kafka-go`) | Kafka driver with strong context/tracing support. |
| Caching | **go-redis/v9** (`github.com/redis/go-redis/v9`) | For distributed locks and caching primitives. |
| Observability | **OpenTelemetry** | Vendor-neutral tracing and metrics (via standard OTel Go SDKs). |
| JWT | **golang-jwt/v5** | Secure, standard parsing for Tenant and Actor context extraction. |
| UUID | **google/uuid** | Standardized v4 / v7 identity generation. |

---

## 3. Core Primitives & Capabilities

### 3.1 Idempotency Engine
To prevent duplicate state mutations during network retries, all mutating operations (`POST`, `PUT`, `PATCH`) must be idempotent. The Kernel provides an engine backed by PostgreSQL or Redis.

**Code Example: Idempotency Wrapper**
```go
import "github.com/InnoSure-Platform/pc-shared-go/idempotency"

// Inside an Application Command Handler
func (h *CreatePolicyHandler) Handle(ctx context.Context, cmd CreatePolicyCmd) (*Response, error) {
    idempKey := idempotency.KeyFromContext(ctx)
    
    // The engine automatically skips execution if the key was already successfully processed,
    // returning the cached Response payload.
    return idempotency.Execute(ctx, h.dbPool, idempKey, func(tx pgx.Tx) (*Response, error) {
        // 1. Execute Domain Logic
        policy := domain.NewPolicy(cmd)
        // 2. Persist State
        err := h.repo.Save(ctx, tx, policy)
        return &Response{ID: policy.ID}, err
    })
}
```

### 3.2 Transactional Outbox
To solve the dual-write problem, domain state and domain events are persisted in the exact same database transaction. A background relay (or CDC) reads the outbox table and guarantees at-least-once delivery to Kafka.

**Code Example: Publishing via Outbox**
```go
import "github.com/InnoSure-Platform/pc-shared-go/outbox"

func (h *ActivateProductHandler) Handle(ctx context.Context, tx pgx.Tx, product *domain.Product) error {
    // ... product state mutation ...

    event := &events.ProductActivated{
        ProductID: product.ID,
        Timestamp: time.Now(),
    }
    
    // The outbox package serializes the event (Avro) and writes it to the outbox table 
    // within the provided database transaction.
    return outbox.Publish(ctx, tx, outbox.Message{
        Topic:        "pc.product.lifecycle.v1",
        PartitionKey: product.TenantID + ":" + product.ID,
        Payload:      event,
    })
}
```

### 3.3 Domain Events & Kafka
Standardized publishing and consumption with automatic OpenTelemetry trace propagation and Avro schema registration.

### 3.4 Localisation & i18n
Standardized translation dictionaries for user-facing errors in English (`en`) and Amharic (`am`).

### 3.5 ETB Money & Ge'ez Calendar
Value objects representing the Ethiopian Birr to guarantee exact precision (avoiding floating-point errors) and datetime utilities for transitioning between the Gregorian and Ge'ez calendars.

---

## 4. Package Architecture (Tactical Design)

| Package | Responsibility | Key Interfaces / Structs |
|:---|:---|:---|
| `/auth` | JWT validation, RBAC primitive evaluation. | `Authorizer`, `Claims` |
| `/calendar` | Ge'ez and Gregorian calendar conversions. | `GeezDate`, `Converter` |
| `/errors` | Standardized RFC 7807 Problem Details envelope. | `AppError`, `ErrorCode` |
| `/events` | CloudEvents v1.0 standard envelopes. | `Envelope` |
| `/httpx` | REST helpers, standardized JSON unmarshaling/marshaling. | `RespondJSON`, `ParseBody` |
| `/i18n` | Internationalization catalogs. | `Translator` |
| `/idempotency` | Duplicate request detection and caching. | `Store`, `Execute()` |
| `/kafka` | Consumer groups, producer wrappers with tracing. | `Publisher`, `Subscriber` |
| `/middleware` | gRPC and HTTP interceptors. | `LoggingInterceptor`, `Recovery` |
| `/money` | Fixed-point integer representations of ETB. | `ETB`, `Add()`, `Multiply()` |
| `/outbox` | Transactional outbox pattern persistence. | `Publish()`, `Relay` |
| `/telemetry` | OpenTelemetry setup, custom metrics. | `InitTracer()`, `Counter` |
| `/tenant` | Multi-tenancy context extraction. | `TenantFromContext()` |

> **Note on Emerging Packages:** The kernel is evolutionary. As new cross-cutting requirements materialize (e.g., custom IAM sync engines, platform-wide feature flags), they will be added directly to the `pc-shared-go` service incrementally over time.

---

## 5. Integration Contracts & Middleware

The Kernel mandates specific interceptor chains for all entrypoints to ensure platform consistency.

### 5.1 Required gRPC Interceptor Chain
Every gRPC server deployed on the platform MUST utilize the shared interceptor chain:

```go
import "github.com/InnoSure-Platform/pc-shared-go/middleware/grpc"

server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        pcgrpc.TraceInterceptor(),      // 1. Extract/inject OTel Traces
        pcgrpc.TenantInterceptor(),     // 2. Extract x-tenant-id into Context
        pcgrpc.AuthInterceptor(keys),   // 3. Validate JWT and extract Actor
        pcgrpc.LoggingInterceptor(),    // 4. Structured slog with trace_id
        pcgrpc.RecoveryInterceptor(),   // 5. Catch panics and return internal error
    ),
)
```

### 5.2 Required HTTP Middleware Chain
Similarly, for REST endpoints:

```go
import "github.com/InnoSure-Platform/pc-shared-go/middleware/http"

r := chi.NewRouter()
r.Use(pchttp.TraceMiddleware)
r.Use(pchttp.TenantMiddleware)
r.Use(pchttp.IdempotencyExtractionMiddleware)
r.Use(pchttp.LoggingMiddleware)
r.Use(pchttp.RecoveryMiddleware)
```

---

## 6. Observability & Telemetry Standards

`pc-shared-go` abstracts away the complexities of OpenTelemetry.

- **Tracing**: Every HTTP and gRPC request automatically starts a span. Database queries (via `pgx`) and Kafka publishes injected with the context will become child spans automatically.
- **Logging**: The `/telemetry` package overrides the default `slog.Handler` to ensure that every log output automatically pulls `trace_id`, `span_id`, and `tenant_id` from the context.

**Code Example: Contextual Logging**
```go
import "github.com/InnoSure-Platform/pc-shared-go/telemetry/log"

// The trace_id and tenant_id are automatically appended to the JSON log output
log.Info(ctx, "Activating policy", slog.String("policy_id", policy.ID))
```

---

## 7. Behaviour-Driven Usage Scenarios (BDD)

### 7.1 Idempotency Conflict Resolution
* **Given** a client issues a `CreateClaim` command with Idempotency Key `UUID-1`
* **And** the request was successfully processed and the database committed
* **When** a network timeout occurs and the client retries the exact same request with `UUID-1`
* **Then** the Kernel's `idempotency.Execute` middleware detects the cached key
* **And** intercepts the request BEFORE it reaches the domain logic
* **And** returns the exact same cached HTTP 201 response.

### 7.2 Outbox Atomicity
* **Given** an Application Handler wrapped in a `pgx.Tx`
* **When** the domain state is successfully updated in PostgreSQL
* **And** the `outbox.Publish()` function is called
* **But** the handler panics before `tx.Commit()`
* **Then** the `RecoveryInterceptor` catches the panic
* **And** the transaction is rolled back
* **And** neither the domain state NOR the outbox event are persisted, ensuring strict consistency.

---

## 8. Engineering Definition of Done (DoD)

Before any new package or primitive in `pc-shared-go` is merged to `main`, it MUST pass the Tier-0 Quality Gates:

1. **Test Coverage:** Minimum 95% unit test coverage. As the foundation, bugs here cascade to the entire platform.
2. **Data Race Freedom:** `go test -race` must pass cleanly for all concurrent primitives (e.g., caching, pub/sub).
3. **Overhead SLA:** Middleware chains (Auth, Logging, Tracing) must add less than `< 1ms` of latency to the request lifecycle (P99).
4. **Zero Panics:** The `RecoveryInterceptor` must be proven to catch all application-level panics and safely convert them to gRPC `Internal` or HTTP `500` status codes without crashing the pod.
5. **Security:** Zero high/critical vulnerabilities in transitive dependencies (scanned via `govulncheck`).
