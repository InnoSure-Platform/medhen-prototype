# Medhen Demo Storyboard — Steps 1–7

EIC stakeholder narrative aligned to PRD §34.1.

---

## Step 1 — Register party

**Screen:** `/quote` → party step  
**API:** `POST /api/v1/parties`

**Show:** Bilingual name (Abebe Kebede / አበበ ከበደ), Ethiopian address hierarchy (Region→Kebele), Telebirr phone.

**Talking point:** Single golden customer record — reused across quote, policy, claims.

---

## Step 2 — Product & shared core

**Screen:** `/staff` → risk schema (teaser)  
**API:** `GET /api/v1/products/MOTOR-PRIVATE-COMP/risk-schema`

**Show:** JSON fields for Motor only — Life/Property = same core, different schema config.

**Talking point:** Configuration over code — tier-1 PAS positioning.

---

## Step 3 — Quote + STP underwriting

**Screen:** `/quote` → risk → premium table  
**API:** `POST /api/v1/quotes`

**Show:** Itemized premium — base, VAT 15%, stamp duty. **UW decision: ACCEPT** (instant).

**Talking point:** < 15 min quote-to-bind; 80%+ STP target for standard motor.

---

## Step 4 — Bind

**API:** `POST /api/v1/quotes/{id}/bind`

**Show:** Policy moves to `PENDING_PAYMENT`, invoice created.

**Talking point:** Saga-ready bind ↔ billing (outbox → Kafka in mesh mode).

---

## Step 5 — Telebirr pay & issue

**Screen:** `/quote` → pay → **document cards**  
**API:** `POST /api/v1/billing/invoices/{id}/pay`

**Show:**
- Policy number: `EIC/MOT/2026/000001`
- Download **PDF** schedule (EN + AM), COI, QR windshield sticker
- Telebirr receipt ID

**Talking point:** Ethiopian localization — ETB, Telebirr, bilingual NBE-style documents.

---

## Step 6 — Mobile FNOL

**Screen:** `/claim`  
**API:** `POST /api/v1/claims`

**Show:** GPS coordinates, photo object keys, **FAST_TRACK** routing (≤ 50k ETB).

**Talking point:** Multi-channel FNOL; UK claims practice triage.

---

## Step 7 — Fast-track settlement + governance

**Screen:** `/claim` settle → `/staff`  
**API:** `POST /api/v1/claims/{id}/settle`, `GET /api/v1/audit`, `GET /api/v1/demo/kpis`

**Show:**
- Claim **SETTLED**, SMS notification (mock)
- Audit trail: party → quote → bind → pay → issue → FNOL → settle
- KPI: policies in-force, GWP

**Talking point:** End-to-end automation + immutable audit — regulatory defensibility.

---

## Screenshot checklist (facilitator)

Capture these for the pitch deck:

1. Home hero (EN + AM toggle)
2. Login screen (Keycloak)
3. Premium breakdown with STP ACCEPT
4. Policy issued + **document download grid**
5. Claim fast-track settled
6. Staff KPI + audit list
7. Risk schema (shared-core teaser)
