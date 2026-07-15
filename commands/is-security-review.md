---
description: Security audit of the branch diff (IS-AQG-001 §6) — InnoSure trigger matrix + regulated-platform checks, then the built-in /security-review engine
---

Security-audit the current branch diff. Run this whenever the diff touches any
§6 trigger; for R4 diffs it is **mandatory** (IS-AQG-001 §3.2/§6). Read-only —
propose fixes, never weaken a control to clear a finding.

Arguments (optional): $ARGUMENTS — a specific area to focus on; otherwise the
whole diff.

### 1. Trigger check (IS-AQG-001 §6)

Confirm which apply and audit each: authn/authz/IAM · tenant isolation ·
crypto/signing (HMAC pack signing!) · secrets/config · input parsing / DSL
evaluation · SQL · file/artifact handling · any externally reachable endpoint.
If none apply, say so and stop — do not manufacture findings.

### 2. Regulated-platform expectations (standing)

- Secrets never in code or test fixtures — OpenBao/ESO paths only.
- All external input validated at the API boundary; the domain assumes nothing.
- Every privileged action emits an audit event, **including denials** — the
  `DENY_SOD` rule fires ONLY for true SoD violations, not for all approval errors.
- No PII / financial data in logs; structured `slog` only.
- Dependency hygiene: `govulncheck` clean; `go mod tidy` diff clean.

### 3. Active checks for R4 (IS-ARP-001 §4)

- **authz matrix** — each role × each protected operation.
- **tenancy probes** — tenant A cannot read or affect tenant B's data.
- **input abuse** — malformed / oversized / injection payloads rejected cleanly.

### 4. Run the built-in engine

Fold its findings into yours: `/security-review`.

### 5. Output

Findings table: `# / severity / file:line / issue / fix`, in the PR
evidence-package disposition format (IS-DQP-001 §6). A security finding is a
blocker until fixed or explicitly waived by the architect with a reason.
