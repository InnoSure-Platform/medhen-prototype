---
description: G3 adversarial self-review of the current branch diff against the spec (IS-DQP-001 §4)
---

You are reviewing the current branch diff as a hostile outside reviewer. Do
NOT assume the implementation rationale was correct, and do not trust any
explanation from earlier in this session.

Arguments (optional): $ARGUMENTS — treat as the spec path/section and REQ IDs
under review. If not given, infer them from the branch name and commit
messages, and say what you inferred.

Steps:

1. Determine the diff base: task branches (`feat/<layer>-task-<n>-…`) are
   reviewed against their layer integration branch (`layer/<layer>`); a layer
   branch itself is reviewed against `main`. Then load the full diff cold:
   `git diff <base>...HEAD` and `git log <base>..HEAD --oneline`.
2. **Assume this diff contains at least one bug the tests miss. Find it.**
3. Check each InnoSure invariant (IS-AQG-001 §3.3) explicitly:
   - spec conformance: list any behaviour in the diff that appears in NO spec
     clause (unspecified behaviour), and any spec clause it contradicts
   - hexagonal boundaries (domain imports nothing; SQL/pgx/redis/kafka only in
     infrastructure; handlers → application services, never repos)
   - tenancy: any query, cache key, or event not tenant-scoped
   - audit/outbox: any state change that emits no audit action or event
   - idempotency: any replay path that begins a new transaction
   - migrations: schema-contract test updated; chain applies to a clean DB
4. Audit the NEW TESTS as hard as the code: assertions that cannot fail,
   fakes diverging from real adapter semantics, tests encoding implementation
   bugs as expected behaviour, `time.Sleep` instead of injected clock.
5. Output a findings table: # / severity (blocker, major, minor) / file:line /
   finding / suggested fix. If you find nothing in a category, say what you
   checked so the absence is evidence, not silence.
