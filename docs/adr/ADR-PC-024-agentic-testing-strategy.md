# ADR-PC-024: Agentic testing strategy — agent roles per test type

- **Status:** Proposed — stub pending Architect ratification
- **Date:** 2026-07-13
- **Applies to:** Platform Core (`pc-*`) services and product (`{xx}-`) services
- **Source of truth:** [Service Registry](../prd/service-registry.md) · [Capability Document](../prd/Medhen-Platform-Capability-Document.md) · builds on [ADR-PC-002](./ADR-PC-002-service-module-and-hexagonal-layout.md) (hexagonal layers) and the per-layer task.md decomposition (IS-DQP-001)

## Context

In an agentic framework, multiple agents contribute to a service's implementation. A question arises: should the **implementation agent** write tests for the code it wrote (fast feedback, deep understanding of intent), or should a **separate verification agent** write tests independently (independent validation, catches blind spots)? The answer is **both — but strategically, per test type.**

- **Unit tests** (module-local, fast): implementation agent writes these; tests the module's contract in isolation. High confidence; agent understands the intent.
- **Integration tests** (cross-layer, cross-service): separate verification agent writes these; tests from the boundary inward (how modules interact). Independent perspective; catches coupling bugs the implementation agent missed.
- **E2E / acceptance tests:** verification agent or product owner; tests the whole service flow end-to-end.

## Decision

**Testing is layered, and agent roles are divided by test scope:**

| Layer | Implementation | Tests | Verification | Min coverage |
|---|---|---|---|---|
| **Domain** (entity logic, rules) | Implementation agent | Unit tests + domain fixtures | Verification agent reviews | 80% |
| **Application** (use cases, transactions) | Implementation agent | Unit tests + fixture mocks | Verification agent reviews | 85% |
| **API** (gRPC/REST contracts) | Implementation agent | Contract tests (proto/OpenAPI compliance) | Verification agent (integration to boundary) | 75% |
| **Infrastructure** (DB, cache, messaging) | Implementation agent | Testcontainers + local mocks | Verification agent (end-to-end infra flow) | 50% |
| **Contracts / e2e** (cross-service, published contracts) | Verification agent | End-to-end integration tests; pc-contracts compliance gates | Verification agent | 100% (contract coverage) |

**Implementation agent workflow (per layer):**
1. Author code (domain entity, application use case, API handler, infra repo, etc.).
2. Write **unit tests** — tests the module's behavior in isolation (mocked dependencies).
3. Mark the layer-task DoD `testing: unit tests written + coverage ≥ threshold`.
4. Verification agent picks up from there.

**Verification agent workflow:**
1. Review the implementation + unit tests for gaps.
2. Write **integration tests** — test the layer's interactions with adjacent layers and the boundary (e.g. how the API calls the application, how the application persists to the DB).
3. Run the full per-layer test suite + verify coverage gates pass.
4. Mark the layer-task DoD `testing: integration tests written + spec-trace gate PASS`.

**Coverage enforcement:**
The `spec-trace.sh` gate (IS-DQP-001 §1) now verifies coverage thresholds per layer, reading them from the task.md and checking the repo's actual coverage (via `go test -covermode=atomic` or equivalent). Stale or missing coverage is a gate failure.

## Consequences

Enforced by the per-layer task.md (IS-DQP-001) and the `spec-trace.sh` mechanical gate: every layer task declares its coverage floor (domain 80%, app 85%, api 75%, infra 50%, contracts 100%); implementation + verification agents both own the tests in their respective scope; spec-trace fails if coverage is missing or below floor. This prevents weak tests and ensures independent verification at the integration boundary.

Requires: (1) per-layer task.md to declare coverage requirements in the DoD; (2) `spec-trace.sh` to check coverage metrics against those declarations; (3) CI to run `go test -covermode=atomic` and feed results to spec-trace; (4) agents to understand the division of labor (implementation writes unit, verification writes integration/e2e).

This is a stub capturing the platform's testing philosophy — expand with alternatives considered (100% implementation ownership vs 100% verification ownership) and full rationale before ratifying.
