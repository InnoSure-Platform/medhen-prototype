# Medhen Web — Tier-0 Enterprise Frontend

Premium, accessible, bilingual (English / አማርኛ), themeable (light/dark) Next.js portal
for the Ethiopian Insurance Corporation's Medhen motor-insurance platform. Talks to the
Go modular monolith through a same-origin, token-less proxy.

## Stack

- **Next.js 15** (App Router) · **React 19** · **TypeScript** (strict)
- **Tailwind CSS v4** (CSS-first `@theme`) with a **Style-Dictionary-shaped design-token
  pipeline** (`tokens/` → `pnpm tokens:build` → `styles/primitives.css` + `styles/tokens.css`)
- **shadcn/Radix** primitives + **CVA** + **lucide-react** icons
- **next-intl** (locale routing `/en`, `/am`) · **next-themes** (class dark mode)
- **TanStack Query** + **openapi-fetch** (types generated from `../api/openapi/medhen.v1.yaml`)
- **React Hook Form + Zod** forms · **TanStack Table** · **Recharts** · **Framer Motion**
  · **sonner** toasts · **cmdk** command palette
- **Storybook** (react-vite) as the living design system · **Vitest** / **Playwright** (planned)

## Getting started

```bash
pnpm install
cp .env.local.example .env.local   # fill in Keycloak + NEXTAUTH secrets
pnpm gen:api                        # regenerate API types from the OpenAPI spec
pnpm tokens:build                   # regenerate design-token CSS
pnpm dev                            # http://localhost:3000  (redirects to /en)
pnpm storybook                      # http://localhost:6006  design system
```

The app proxies API traffic to the monolith at `MEDHEN_API` (default `http://localhost:8080`).
Start the backend with `make infra-up && make api` from the repo root.

## Scripts

| Script | Purpose |
|---|---|
| `pnpm dev` / `build` / `start` | Next.js dev / production build / serve |
| `pnpm typecheck` · `pnpm lint` | `tsc --noEmit` · ESLint |
| `pnpm gen:api` | Regenerate `lib/api/schema.d.ts` from the OpenAPI contract |
| `pnpm tokens:build` | Regenerate design-token CSS from `tokens/src/*.json` |
| `pnpm storybook` / `build-storybook` | Run / build the design system |

## Architecture

```
app/[locale]/
  (marketing)/        public landing + login          (top-nav shell)
  (customer)/customer dashboard, quote wizard, policies, claims, invoices, documents
  (broker)/broker     book-of-business, clients, new-business, commissions
  (staff)/staff       KPI dashboard, quotes, policies, claims settle, underwriting, audit, users
app/api/              medhen/[...path] proxy · auth/[...nextauth] · auth/federated-logout
components/ui/        design-system primitives (Storybook stories alongside)
components/shell/     app shell (sidebar, topbar, command palette, theme/locale toggles)
components/patterns/  composed patterns (PageHeader, …)
lib/                  api client + hooks, i18n, format, nav, recents, query, utils
tokens/               design-token source + build script
messages/             en.json · am.json  (next-intl ICU)
```

### Security (ported verbatim, security-reviewed)

- `app/api/medhen/[...path]/route.ts` attaches the access token + tenant **server-side**;
  the browser never holds a token and never learns the upstream origin.
- Auth fails **closed at runtime** — no hardcoded secret fallbacks. Verify:
  `env -u KEYCLOAK_SECRET -u NEXTAUTH_SECRET pnpm build` must still succeed.
- `middleware.ts` combines next-intl locale routing with role gates (staff/broker/admin).
- CSP / HSTS / frame / nosniff headers in `next.config.ts`.

### Theming & i18n

Design tokens are authored once in `tokens/src/*.json` (Tokens-Studio shape, Figma-exchangeable),
built into light (`:root`) and dark (`.dark`) CSS variables, and mapped into Tailwind's color
namespace via `@theme inline`. Amharic renders in **Noto Sans Ethiopic**; English/UI in Inter,
display in Plus Jakarta Sans, IDs/money in mono.
