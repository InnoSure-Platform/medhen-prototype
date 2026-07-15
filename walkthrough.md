# Phase 1: Production Motor — Keycloak & Docker Mesh Walkthrough

I have implemented the three immediate Phase 1 priorities out of PRD §34. The prototype now uses actual Next.js Keycloak integration and runs entirely inside Docker Compose instead of local bash processes.

## 1. Keycloak on the Web Portal
- Replaced the custom `/api/auth/login` endpoint with `next-auth` (v4).
- Added `web/app/api/auth/[...nextauth]/route.ts` which uses the standard Keycloak OAuth2/OIDC provider.
- Rewrote the `AuthProvider.tsx` to wrap the app in a `<SessionProvider>` and securely manage the `accessToken` for downstream API calls.
- Updated the `app/login/page.tsx` to automatically redirect users to Keycloak's login page rather than using an embedded username/password form.

## 2. Mesh E2E on Docker Compose
- Created a multi-stage `Dockerfile` in the root that builds all `pc-*-svc` and `pc-gateway` Go binaries in an alpine builder, and outputs a lightweight Alpine runtime image.
- Updated `infra/docker-compose.yml` to run the entire backend stack. This means `pc-gateway`, `pc-policy-svc`, `pc-party-mgmt-svc`, and all others are now first-class containers.
- They correctly resolve `postgres`, `redpanda`, `valkey`, and `keycloak` over the Docker internal network.

## 3. Demo Runbook
- Updated `docs/demo/DEMO-RUNBOOK.md` to guide facilitators towards using `cd infra && docker compose up -d --build` rather than the old `mesh-up.sh` script.

> [!TIP]
> **Try it out:**
> 1. Start the stack: `cd infra && docker compose up -d --build`
> 2. Open the frontend: `cd web && npm run dev`
> 3. Go to http://localhost:3000 and try signing in. You should be redirected to the Keycloak login screen.
> 
> *Note: For the best experience, wait ~15 seconds for Keycloak to finish starting up inside Docker before clicking login.*

---

# Phase 1: Policy Domain (Endorsements, Renewals, Cancellations)

The Policy bounded context has been upgraded from a static "issued once" model to a living contract model. Policies can now undergo mid-term adjustments, renewals, and cancellations using **Immutable Versioning**.

## 1. Schema & Store (Immutable Versions)
- Added `parent_policy_id` and `version` to the `policies` table in PostgreSQL.
- When a policy is endorsed or renewed, the old row is marked `SUPERSEDED` and a new `ISSUED` row is created, preserving the entire history of the policy state over time.

## 2. Core Business Logic (`usecase/motor.go`)
- **Endorsements:** Implemented `EndorsePolicy` which recalculates the premium based on new risk details, calculates the difference against the old premium, and automatically generates an invoice for the difference.
- **Renewals:** Implemented `RenewPolicy` which clones the latest policy state, advances the effective dates by 1 year, re-rates the premium, and leaves the new policy in `PENDING_PAYMENT` status.
- **Cancellations:** Implemented `CancelPolicy` which immediately terminates the policy and sets the status to `CANCELLED`.

## 3. Web Portal Integration
- Added an **Active Policy Operations** component to the Staff Dashboard (`web/app/staff`).
- Added a dedicated **Policy Details View** (`web/app/staff/policies/[id]`) with action buttons for Endorse, Renew, and Cancel.
- Endorsing a policy now opens an **Endorsement Wizard Modal** where staff can modify the Sum Insured.

> [!TIP]
> **Try out Endorsements:**
> 1. Complete a standard motor quote and payment flow on the web portal to generate a policy.
> 2. Open the **Staff Dashboard**. You will see a new link: "Latest Quote/Policy".
> 3. Click **View Policy Details**.
> 4. Click **Endorse (Mid-Term)** and change the Sum Insured to see a new version of the policy instantly generated with a pro-rata invoice!

---

# Phase 1: Billing Domain (Installments & ERP)

The billing infrastructure has been upgraded to support flexible installment plans and backend reconciliation for General Ledger integration.

## 1. Schema & Store Updates
- Introduced `due_date` and `installment_number` columns to the `invoices` table.

## 2. Installment Logic (`usecase/motor.go`)
- **Flexible Plans:** `BindQuote` now parses the chosen installment plan. If `40_30_30` is selected, the system automatically creates 3 distinct invoices with staggered due dates (today, +30 days, +60 days) and handles the remainder math to ensure the exact total premium is collected.
- **Issuance Logic:** `PayInvoice` ensures that the policy transitions to `ISSUED` as soon as the **downpayment (Installment 1)** is successfully collected. Subsequent installment payments simply mark their respective invoices as paid and log an audit trail event.

## 3. End-Of-Day (EOD) ERP Reconciliation
- Implemented `RunEndOfDayReconciliation()` which queries all successful receipts across all channels (Telebirr) for a given date.
- It calculates the total batch aggregate and triggers an `ERP_RECONCILIATION_COMPLETED` event via the outbox pattern, simulating a push to the EIC General Ledger.

## 4. Web Portal Integration
- The Quote/Checkout page (`web/app/quote`) now features a dropdown allowing the customer to select their preferred Payment Plan prior to binding.
- The Staff Dashboard (`web/app/staff`) features a new **End of Day (EOD) Operations** section with a button to manually trigger and inspect the nightly batch reconciliation JSON payload.

> [!TIP]
> **Try out Installments & EOD:**
> 1. Complete a motor quote, and on the payment step, change the dropdown from **100% Upfront** to **Installments (40/30/30)**. Notice how the downpayment calculation changes dynamically!
> 2. Pay the downpayment. The policy is issued.
> 3. Navigate to the **Staff Dashboard** and click **Run EOD Reconciliation** to see the simulated General Ledger batch payload encompassing your payment!

---

# Phase 1: Underwriting Domain (Referral Workbench)

The automated underwriting engine has been enhanced with a "Referral Workbench", allowing risky policies to be manually reviewed by staff instead of being instantly declined.

## 1. Underwriting Engine Updates
- Modified `CreateQuote` to intercept `REFER` decisions from the Straight-Through Processing (STP) engine (e.g. when Sum Insured > 5M ETB).
- These quotes are now safely saved to the database in a `REFERRED` state, instead of throwing a hard error.

## 2. API Edge & Core Logic
- Created dedicated backend endpoints and usecase methods for underwriters to interact with referrals:
  - `GET /quotes?status=REFERRED` to fetch the queue.
  - `POST /quotes/{id}/approve` to transition the quote to `QUOTED` (approved for payment).
  - `POST /quotes/{id}/decline` to transition the quote to `DECLINED` (rejected).
- These actions are thoroughly tracked via `m.audit`, leaving an immutable audit trail of the underwriter's decision.

## 3. Web Portal Integration
- **Customer Quote Page (`web/app/quote`):** If a quote returns in the `REFERRED` state, the customer's "Pay" button is locked out, and a banner is displayed indicating that the quote is pending manual review.
- **Staff Dashboard (`web/app/staff`):** Introduced a new **Underwriting Referral Workbench** section that lists all pending quotes in a data table, providing one-click "Approve" and "Decline" actions for staff.

> [!TIP]
> **Try out the Referral Workbench:**
> 1. Complete a motor quote, but set the **Sum Insured** to `6,000,000 ETB` (or anything >5M).
> 2. You will see a "Pending Underwriting Review" banner preventing you from paying.
> 3. Log into the **Staff Dashboard**, and locate the new quote in the **Underwriting Referral Workbench**.
> 4. Click **Approve**. 
> 5. Return to the customer quote page (or just check the status), and observe that payment is now unlocked!

---

# Phase 1: Claims Domain (Reserves, Recovery, Total Loss)

The Claims system has been enhanced with robust financial operations, including reserving, subrogation/recovery tracking, and Total Loss scenarios.

## 1. Schema Updates
- Added `ReserveMinor` and `RecoveryMinor` fields to the `Claim` schema in both memory and PostgreSQL repositories, ensuring accurate tracking of financials beyond just the initial estimated amount.

## 2. Core Business Logic
- **Adjust Reserve:** Handlers can dynamically change the reserve of an open claim. Every change is tracked securely through the audit trail.
- **Record Recovery:** Handlers can record subrogation or salvage recoveries on a claim, acting as an offset against total incurred losses.
- **Total Loss Settlement:** Settling a claim on a `TOTAL_LOSS` track enforces that the settlement cannot exceed the policy's `SumInsured`. Upon successful settlement, the underlying policy is automatically `CANCELLED` (since the risk asset is exhausted).

## 3. Staff Workbench
- Expanded the **Staff Dashboard** to include a new **Claims Workbench**.
- Underwriters and claims handlers can view a global list of claims, adjust reserves, input recovery amounts, and issue final settlements via intuitive dashboard actions.

> [!TIP]
> **Try out the new Claims Workbench:**
> 1. Create a policy (bind and pay for it).
> 2. File a claim on the policy using the self-service web portal (or via API). Wait for the claim to be registered.
> 3. Go to the **Staff Dashboard** and locate your claim in the **Claims Workbench**.
> 4. Use the **Reserve** button to adjust the expected cost of the claim.
> 5. Use the **Recovery** button to record a salvage recovery.
> 6. Use the **Settle** button to close the claim. If the claim track is `TOTAL_LOSS`, notice that the policy will instantly transition to a `CANCELLED` state in the system!

---

# Phase 1: Identity Domain (Full Fayda + KYC Workflow)

We integrated the **Fayda National ID** system into our Quote flow to satisfy strict Know Your Customer (KYC) requirements before any policy is bound.

## 1. Schema Updates
- The `Party` (customer) model now includes a `FaydaID` and a `KYCStatus` (defaults to `PENDING`). These are fully persisted to PostgreSQL.

## 2. Core Business Logic
- Extended the `FaydaClient` port to fetch full demographic info (simulated names and active status).
- Implemented `VerifyKYC` use-case in `Motor` which validates the Fayda ID, extracts the citizen's true identity, upgrades their KYC status to `VERIFIED`, and writes a secure audit event.
- Enhanced `BindQuote` to enforce that a quote cannot be bound unless the party is fully KYC verified.

## 3. Web Portal
- The self-service quote flow now has an explicit **Identity Verification** step inserted immediately after personal details but before the vehicle/risk assessment. 
- Customers enter their Fayda ID (e.g., `1234567890`) and the system performs a real-time (mocked) biometric/demographic check against the national registry.

> [!TIP]
> **Try out the Fayda workflow:**
> 1. Start the stack and go to the Customer quote flow.
> 2. Fill out your name and phone number.
> 5. Only after successful KYC verification will the system unlock the underwriting and quote issuance steps.

---

# Phase 1: Operations (GitOps, SLOs & BC Boundaries)

We have formalized the operational boundaries of the platform in alignment with the PRD Phase 1 requirements.

## 1. Per-BC Logical Schema Isolation
Instead of running a monolithic database setup, we modified the `EnsureSchema` system so that each Bounded Context (e.g. `pc_party`, `pc_policy`, `pc_billing`) explicitly provisions *only* its own logical schema in PostgreSQL.
- A `pc_party` service now strictly manages the `parties` table.
- A `pc_policy` service manages `quotes` and `policies`.
- A monolith `pc_medhen` deployment will still hydrate all tables, ensuring backward compatibility for local integration testing.

## 2. GitOps & CI/CD
We added scaffolding for a robust, enterprise-grade deployment pipeline:
- **GitHub Actions Pipeline (`.github/workflows/pipeline.yml`)**: Automates testing for Go, building for Next.js, and pushing Docker images to a registry.
- **GitOps Definitions (`k8s/`)**: Added foundational Kubernetes manifests (`deployment.yaml`, `service.yaml`) ready to be ingested by ArgoCD for continuous delivery.

## 3. Observability & SLOs
- Exposed a standard **`/metrics`** Prometheus endpoint in the Go backend.
- Declared infrastructure **SLOs** as code in `infra/slos.yaml`, tracking system availability (99.9% target) and latency (P99 < 1s for quoting).
