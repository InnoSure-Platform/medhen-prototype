---
description: G2 reuse/simplification pass (IS-DQP-001 §3) — InnoSure guardrails, then the built-in /simplify engine
---

Run the G2 simplification pass on the current branch diff. This is a
**quality-only** pass (reuse, simplification, efficiency, altitude); it does
NOT hunt for bugs — that is `/is-code-review` — and it must never trade
correctness, a test, or a gate for tidiness.

Arguments (optional): $ARGUMENTS — paths/packages to focus on; otherwise the
whole branch diff (task branches diff against `layer/<layer>`, a layer branch
against `main`).

### Guardrails — a "simplification" that does any of these is FORBIDDEN

- **Touch a test to make coverage or mutation pass** — deleting an assertion,
  loosening a matcher, dropping a case. Tests are reviewed as hard as code
  (IS-AQG-001 §4.3); simplifying them away is gaming, not cleanup.
- **Swap or remove a locked tool** (gremlins, `rapid`, testcontainers,
  hand-written fakes — never mockery/gomock). Those are ADR decisions
  (IS-AQG-001 §4), not an agent's call.
- **Cross a hexagonal boundary "to clean up"** — pulling SQL/pgx/redis/kafka
  out of `infrastructure`, letting `domain` import application/infra, or letting
  a handler call a repo directly.
- **Drop a tenancy scope, an audit/outbox emission, or an idempotency guard**
  because it "looks redundant".
- **Lower a coverage floor, add `//nolint`, or weaken any gate.**

### Do

1. Load the diff cold (`git diff <base>...HEAD`).
2. Look for: duplication that should reuse an existing helper/port; needless
   indirection or dead code the diff introduced; a stdlib / existing-dependency
   call that replaces hand-rolled logic; wrong-altitude code (business logic in
   a handler, wiring in the domain).
3. Apply behaviour-preserving cleanups and keep tests green — re-run
   `make test-unit` after.
4. Then run the built-in engine for anything mechanical you missed: `/simplify`.

### After

Re-run `/precheck`. Summarise what you changed **and**, as evidence, what you
deliberately did NOT simplify because a guardrail applied — the next step is a
session-separated `/is-code-review <tier>` (IS-DQP-001 §4).
