---
description: Print the InnoSure quality-gate contract — the full G0→G5 chain, the command per gate, the CI gate order, and the risk-tier ladder (IS-AQG-001)
---

Print the InnoSure gate contract for this repo as a reference. **Do not run any
gate** — this is the map, not the territory. Source of record: IS-AQG-001,
IS-DQP-001 (developer side), IS-ARP-001 (architect side), all in `docs/quality/`.

## Gate chain — who runs what

| Gate | Owner | Command(s) | Question answered |
|------|-------|------------|-------------------|
| **G0** intake | dev | — (read the layer `task.md` trace) | Is the task ↔ SSD § ↔ REQ trace present? |
| **G1** implement | dev | — | Code + tests together, spec-first, in-layer |
| **G2** self-check | dev | `/precheck` → `/is-simplify` → re-`/precheck` | Builds, lints, boundaries, coverage floors? |
| **G3** self-review | dev | `/is-code-review <tier>`, `/self-review`, `/is-security-review` (if any §6 trigger) | Have I caught my own defects, from a fresh hostile context? |
| **G4** self-QA | dev | `/qa-verify` (+ the test tiers the diff requires) | Behaviour verified, incl. the negative quartet? |
| *spec/task edit* | dev | `/spec-trace` | Is each `task.md` a faithful, complete decomposition of the SSD? |
| *integration* | dev | `/integration-ready` | Is the service IG0-ready (IS-IQP-001 §2.1)? |
| **G4** deep-review | architect | `/deep-review` | Does it conform to the spec? (fresh context, sole-merge gate) |
| **G5** merge | architect | — | Sole merger — developers never merge their own PR |

The three `/is-*` commands wrap the built-in `/simplify`, `/code-review`, and
`/security-review` engines: each first applies the InnoSure invariants, risk
tier, and guardrails, then invokes the built-in as its engine. Run the `/is-*`
versions — not the raw built-ins — so the review is calibrated to this platform.

## Risk-tier ladder (IS-AQG-001 §3.2) — spend review effort by blast radius

| Tier | Examples | Dev self-review (G3) | Architect review (G4) |
|------|----------|----------------------|-----------------------|
| R1 cosmetic | docs, comments, renames, test-only | `low` | `medium` + spot-read |
| R2 routine | new handler/query on existing patterns | `high` | `max` |
| R3 structural | new aggregate, new port/adapter, cross-layer | `max` + adversarial | `ultra` |
| R4 critical | authz/IAM, tenancy, migrations, idempotency, crypto/HMAC, event contracts, regulator-facing | `max` + `/is-security-review` | `ultra` + `/is-security-review` + active security testing (IS-ARP-001 §4) |

The architect re-tiers independently — the declared tier is a claim, not a fact.

## CI gate order (IS-AQG-001 §7) — a red gate makes the PR ineligible for review

```
build → lint+format → boundary check → unit -race → coverage floors (95/90/85/80)
→ mutation efficacy (≥85, ratchet →95) → contract/schema → integration (containers)
→ security/dep scan (govulncheck, gitleaks) → drift gates (contract-hygiene,
  spec-trace, platform-kit-drift)
────────────────────────────────────────────────────────────────────────────────
ONLY IF ALL GREEN → eligible for /deep-review
```

## Non-negotiables — never do these in ANY command to pass a gate

Lower a coverage floor · add `//nolint` · skip or weaken a test · swap a locked
tool (gremlins/rapid/testcontainers/hand-written fakes; never mockery/gomock) ·
invent spec behaviour · self-merge · edit `CLAUDE.md`, `docs/quality/*`, the CI
workflow, or the PR template inside a feature PR (those are the architect's
distribution — IS-DQP-001 §7).
