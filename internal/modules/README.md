# internal/modules

Each subdirectory is a **sealed bounded context** in the modular monolith.
Modules migrate here from `services/pc-*-svc` during Phase 3 of
[`docs/refactor/modular-monolith-plan.md`](../../docs/refactor/modular-monolith-plan.md).

## Rules (enforced by arch-lint — see `.go-arch-lint.yml`)

- A module may depend on another module **only** through its published `ports`
  package (Go interfaces), injected via the composition root in
  `cmd/medhen-api`. Never import another module's `domain`, `app`, or `adapters`.
- The `domain` layer imports only `internal/platform/*` — no framework, DB, or
  HTTP.
- Everything may depend on `internal/platform/*` (the shared kernel).

## Module layout

```
<module>/
├── domain/     # entities, value objects, domain services (pure)
├── app/        # command/query handlers (use cases), UoW orchestration
├── ports/      # PUBLIC interfaces this module exposes + consumes
├── adapters/   # driven adapters (postgres repos, external clients)
├── rest/       # driving adapter (HTTP handlers)
├── events/     # domain events published/consumed
└── module.go   # implements app.Module: Name/Init/Mount + facade wiring
```

## Target modules (13)

`iam` · `party` · `product` · `rating` · `underwriting` · `policy` · `billing` ·
`claims` · `document` · `notification` · `integration` · `audit` · `reporting`
