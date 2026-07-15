# Medhen — 15-Minute EIC Demo Runbook

**Audience:** Facilitator (InnoSphere / EIC stakeholders)  
**Duration:** ~15 minutes  
**Outcome:** Live Motor buy→claim on Medhen with bilingual PDFs, audit trail, KPI tile

---

## Before the room (10 min prep)

```bash
cp .env.example .env
make build-all         # optional if just using docker compose
cd infra && docker compose up -d --build
cd ../web && npm install && npm run dev   # terminal 2
```

**Verify:**
```bash
curl -sf http://localhost:8080/health
./scripts/mesh-smoke.sh   # optional full automated check
```

| URL | Purpose |
|---|---|
| http://localhost:3000 | **pc-web** portal (facilitator UI) |
| http://localhost:8080 | API gateway |
| http://localhost:8081 | Keycloak admin (`admin` / `admin`) |
| http://localhost:16686 | Jaeger (optional) |

**Login credentials (demo):**

| User | Password | Role |
|---|---|---|
| `demo-agent` | `medhen-demo` | Agent / quote & bind |
| `demo-claims` | `medhen-demo` | Claims handler |

Seed persona: **Abebe Kebede** (አበበ ከበደ) — see `seeds/demo-personas.json`

---

## Minute-by-minute script

| Min | Step | Say (EN) | Do |
|:---:|---|---|---|
| 0–1 | **Open** | "Medhen is EIC's end-to-end insurance platform — one shared core for Motor, Life, Property." | Open http://localhost:3000 · switch **አማርኛ** |
| 1–2 | **Sign in** | "Tier-0 identity — every API call is JWT-secured via Keycloak." | Login `demo-agent` / `medhen-demo` |
| 2–5 | **Quote (1–3)** | "Register customer, rate motor risk, STP auto-underwriting — no re-keying." | **Quote** → Continue → Calculate → show premium lines (VAT 15%, stamp) |
| 5–7 | **Pay & issue (4–5)** | "Telebirr payment, instant policy issuance, bilingual PDF pack." | Pay with Telebirr → show policy number `EIC/MOT/…` → **Download PDFs** |
| 7–9 | **Claim (6–7)** | "Mobile FNOL with GPS — fast-track settlement under threshold." | **Claim** → submit → Settle fast-track |
| 9–11 | **Governance (8)** | "Immutable audit, GWP KPI — UK-conduct grade traceability." | **Staff** → audit trail + KPI tile |
| 11–13 | **Shared core** | "Same platform core — only product config changes for Life/Property." | Staff → risk schema JSON |
| 13–15 | **Close** | "Production mesh: Postgres, Kafka outbox, microservices — zero throwaway." | Mention architecture slide / Q&A |

Full narrative: [`STORYBOARD.md`](./STORYBOARD.md)

---

## Fallbacks

| Issue | Fix |
|---|---|
| Login fails | Keycloak still starting — wait 30s, retry. Check `curl localhost:8081/realms/medhen` |
| API 401 | Ensure logged in on web; or run `./scripts/keycloak-token.sh` for CLI |
| Mesh down | `cd infra && docker compose restart` |
| Monolith fallback | `MEDHEN_DOCS_DIR=./data/docs ./bin/medhen-api` (no JWT if `KEYCLOAK_URL` unset) |

---

## CLI rehearsal (no UI)

```bash
export KEYCLOAK_URL=http://localhost:8081
./scripts/demo-e2e.sh
```

---

## Telebirr sandbox (when credentials arrive)

1. Add to `.env`: `TELEBIRR_APP_ID`, `TELEBIRR_APP_SECRET`, `TELEBIRR_SHORT_CODE`, `TELEBIRR_BASE_URL`
2. Restart mesh: `cd infra && docker compose up -d --build`
3. Prove: `./scripts/telebirr-prove.sh`

---

## After demo

```bash
cd infra && docker compose down
```

**Artifacts to share:** PDFs in `data/docs/`, KPI JSON from `/api/v1/demo/kpis`, audit from `/api/v1/audit`
