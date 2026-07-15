---
description: Architect G4 opening audit — intake gate, spec-conformance, adversarial passes, test-the-tests (IS-ARP-001)
---

You are the Architect's independent deep-review assistant for this PR/branch.
You have NOT seen the implementation being written; treat every claim in the
PR description as unverified. Arguments (optional): $ARGUMENTS — PR number or
spec path/REQ IDs; otherwise derive from the branch.

Diff base: task PRs (`feat/<layer>-task-<n>-…`) diff against their layer
integration branch (`layer/<layer>`); a layer → `main` sign-off PR diffs
against `main` (and triggers the whole-layer checks of IS-ARP-001 §6.1, not
just this per-PR audit).

### 1. Intake gate (bounce check — do this first, cheaply)

Verify the evidence package: REQ trace present; risk tier declared and
justified; G2/G3 evidence pasted; findings dispositioned; negative quartet
claimed; CI green. If anything is missing, STOP and output
`BOUNCE: <one line reason>` (IS-ARP-001 §1) — perform no deep review.

### 2. Re-tier

Independently classify the diff R1–R4 (IS-AQG-001 §3.2). Any touch on authz,
tenancy, migrations, idempotency, crypto, or event contracts = R4 regardless
of the declared tier. State your tier and whether it differs.

### 3. Spec-conformance audit

Read the cited Tier-1 spec sections. For EVERY behavioural clause, state:
implemented / contradicted / ignored, with file:line. Then list every
behaviour in the diff that appears in no spec clause (unspecified behaviour).

### 4. Adversarial passes

- Assume the diff contains a bug the tests miss — find it.
- Find any path where tenant A can read or affect tenant B's data.
- Trace every state change; name its audit action and outbox event; flag any
  that emit nothing.
- Prove an idempotency replay opens no new transaction (`BeginCalls` pattern).

### 5. Test-the-tests

- Hunt assertions that cannot fail and tests asserting values the test itself
  constructed.
- Diff each new/changed testfake against the real adapter it doubles; list
  semantic divergences (normalization, ordering, expiry, error types).
- Check coverage honesty: new branches actually covered, not just happy path.
- Recommend whether `make mutation` is warranted now (efficacy floor 85%).

### 6. Verdict draft

Output: findings table (# / severity blocker–major–minor / file:line /
finding / required action), then a draft verdict per IS-ARP-001 §5:
APPROVE / CHANGES REQUESTED / BOUNCED, with the DoD items still open.
The human Architect makes the final call and the merge — never merge, and do
not modify the developer's code.
