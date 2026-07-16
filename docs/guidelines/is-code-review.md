---
description: G3 risk-tiered adversarial review (IS-AQG-001 §3.2/§3.3) — InnoSure invariants, then the built-in /code-review engine. Arg: tier low|high|max|ultra
---

Review the current branch diff as a **hostile outside reviewer** at the given
risk tier. Do NOT trust any explanation from earlier in this session — the
writing session is biased toward believing its own code works (IS-AQG-001 §3.4).

Arguments: $ARGUMENTS — the risk tier (`low` | `high` | `max` | `ultra`) and,
optionally, the spec path / REQ IDs. If no tier is given, infer it from the diff
per the ladder below and **state the tier you picked** (the declared tier is a
claim to be justified, not decoration).

### 1. Tier the diff (IS-AQG-001 §3.2)

- **R1 cosmetic** (docs, comments, renames, test-only) → `low`
- **R2 routine** (new handler/query on existing patterns) → `high`
- **R3 structural** (new aggregate, new port/adapter, cross-layer change) →
  `max` + adversarial pass
- **R4 critical** (authz/IAM, tenancy, migrations, idempotency, crypto/HMAC,
  outbox/event contracts, regulator-facing) → run `max` here AND flag the
  architect's `ultra` + `/is-security-review`. Never self-certify an R4 alone.

### 2. InnoSure invariants — check EACH explicitly (IS-AQG-001 §3.3)

State pass/finding with `file:line` for every one. "Nothing found" must say what
you checked, so the absence is evidence, not silence.

1. **Spec conformance** — behaviour matches the cited REQ IDs and spec'd error
   codes; list any behaviour in NO spec clause (unspecified behaviour is the
   most common agentic failure mode).
2. **Hexagonal boundaries** — `domain` imports nothing from
   application/api/infrastructure; no SQL/pgx/redis/kafka outside
   `infrastructure`; handlers → application services, never repos.
3. **Tenancy** — every query, cache key, and event is tenant-scoped; no
   cross-tenant read path, ever.
4. **Audit & events** — every state change emits its spec'd audit action + outbox
   event; assert observable effects, not "no error returned".
5. **Idempotency** — replays honour the key and open no new transaction (the
   `BeginCalls` assertion pattern).
6. **Migration ↔ repo contract** — schema-contract test updated; the migration
   chain applies to a clean DB.
7. **Concurrency** — suites run under `-race`; async workers are testable
   synchronously.

### 3. Tests reviewed as hard as code (IS-AQG-001 §4.3)

Hunt: assertions that cannot fail; fakes diverging from real-adapter semantics
(normalization, ordering, expiry, error types); tests encoding a bug as expected
behaviour; `time.Sleep` instead of an injected clock; coverage-gaming (lines
executed, outcomes unasserted).

### 4. Run the built-in engine at the tier

After your invariant pass, run the multi-agent engine and fold its findings in:

- `low` / `medium` / `high` / `max`: `/code-review <tier>` (add `--fix` to stage
  patches, `--comment` to post inline PR comments).
- `ultra` (architect tier — cloud multi-agent): `/code-review ultra`.

### 5. Output

A findings table: `# / severity (blocker · major · minor) / file:line / finding /
required action`, in the disposition format the PR evidence package needs
(IS-DQP-001 §6). Disposition every finding: fixed (commit hash) or waived
(one-line reason). Never lower a floor, weaken a test, or invent spec behaviour
to clear a finding.
