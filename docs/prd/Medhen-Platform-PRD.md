# Medhen Platform — Product Requirements Document (PRD)

## Ethiopian Insurance Corporation · End-to-End Insurance Automation

**Document ID:** PRD-MEDHEN-2026-001
**Classification:** Confidential
**Version:** 1.0
**Date:** July 2026
**Author:** InnoSphere Technologies — Platform Engineering
**Status:** Draft — Pending Stakeholder Review
**Authoritative capability source:** [`Medhen-Platform-Capability-Document.md`](./Medhen-Platform-Capability-Document.md) (MDH-CAP-001 v1.0) — the BC-decomposed strategic/architectural source of truth. This PRD is the **engineering contract** that makes every capability testable; it traces bidirectionally to the Cap doc.
**Derives from:** [`product_definition_document.md`](./product_definition_document.md) (PDD v1.0 — capability catalog); superseded, preserved for audit
**Traceability:** Bidirectional `REQ ↔ CAP ↔ BC ↔ Regulatory` (see [§6](#6-bidirectional-traceability-framework)); BC-IDs (`BC-MDH-01…20`) are defined in the Cap doc Part III

---

> [!IMPORTANT]
> This PRD is the **single source of truth** for functional and non-functional requirements of the Medhen platform. It supersedes the capability-catalog PDD as the requirements authority. Every requirement carries a stable `REQ-ID`, one or more `CAP-ID` capability links, a priority, a phase, and Given-When-Then acceptance criteria. Any capability not traceable to a requirement here is out of scope unless added through the change-control process in §37.

---

## Document Control

| Role | Name | Responsibility |
|:---|:---|:---|
| **Product Owner (EIC)** | TBD | Requirements ownership, priority arbitration, business sign-off |
| **Principal Solutions Architect** | TBD | Cross-product architectural alignment, shared-core integrity |
| **Platform Engineering Lead** | TBD | Implementation estimation, technical constraints |
| **Underwriting SME (EIC)** | TBD | Underwriting & product-rule validation |
| **Claims SME (EIC)** | TBD | Claims-process validation |
| **Finance / Actuarial Lead (EIC)** | TBD | Billing, reinsurance, IFRS-17 validation |
| **Compliance Officer (EIC)** | TBD | NBE / data-protection / UK-conduct alignment |
| **Security Architect** | TBD | IAM, audit, encryption, residency validation |
| **QA Lead** | TBD | Acceptance-criteria & contract-test validation |

### Revision History

| Version | Date | Author | Changes |
|:---|:---|:---|:---|
| 0.1 | Jul 2026 | InnoSphere Platform Engineering | Initial PRD derived from PDD v1.0; foundation + Module 1 (Party) template |
| 1.0 | Jul 2026 | InnoSphere Platform Engineering | Full module coverage, cross-cutting requirements, traceability appendices |

### Requirement ID Convention

All requirements follow the format **`REQ-{DOMAIN}-{NUMBER}`**. Numbers are allocated in tens per functional group (001, 010, 020 …) so additive requirements can be inserted without renumbering.

| Domain | Module / Concern | Capability Prefix | Bounded Context (Cap doc) |
|:---|:---|:---|:---|
| `PTY` | Party & Customer Management | `CAP-PARTY` | `BC-MDH-01` |
| `PRD` | Product Definition Engine | `CAP-PROD` | `BC-MDH-02` |
| `POL` | Policy Administration | `CAP-POL` | `BC-MDH-03` |
| `RAT` | Rating & Premium Calculation | `CAP-RATE` | `BC-MDH-04` |
| `UW`  | Underwriting | `CAP-UW` | `BC-MDH-05` |
| `CLM` | Claims Management | `CAP-CLM` | `BC-MDH-06` |
| `BIL` | Billing & Payments | `CAP-BIL` | `BC-MDH-07` |
| `DOC` | Document Management | `CAP-DOC` | `BC-MDH-08` |
| `WFA` | Workflow & Approvals | `CAP-WF` | `BC-MDH-09` |
| `NOT` | Notifications | `CAP-NOT` | `BC-MDH-10` |
| `RPT` | Reporting & Analytics | `CAP-RPT` | `BC-MDH-11` |
| `RIN` | Reinsurance & Coinsurance | `CAP-RI` | `BC-MDH-12` |
| `CPL` | Complaints & Disputes *(new)* | `CAP-CPL` | `BC-MDH-13` |
| `FIN` | Financial Crime — Fraud / AML / Sanctions *(new)* | `CAP-FIN` | `BC-MDH-14` |
| `COM` | Commission Management *(new)* | `CAP-COM` | `BC-MDH-15` |
| `IAM` | Identity & Access Management | `CAP-IAM` | `BC-MDH-16` |
| `AUD` | Audit & Compliance | `CAP-AUD` | `BC-MDH-17` |
| `INT` | Integration & Anti-Corruption Layer | `CAP-INT` | `BC-MDH-18` |
| `OBS`/`NFR` | Observability & Non-Functional | `CAP-OBS` | `BC-MDH-19` |
| `CORE`| Shared-Core Kernel & Product Extensibility *(new)* | `CAP-CORE` | `BC-MDH-20` |
| `I18N`/`BRN` | Internationalization & Multi-Branch | `CAP-I18N`/`CAP-BR` | cross-cutting (Kernel BC-MDH-20 / §II.8) |
| `SEC` | Security Requirements | — | cross-cutting (§10 Cap doc) |
| `DAT` | Data Requirements | — | cross-cutting |
| `CMP` | Compliance & Regulatory Requirements | — | cross-cutting (Cap doc Part VII) |
| `MTR`/`LIF`/`PRP`/`MAR`/`LIA`/`ENG` | Product-line extensions | per LOB | `BC-MDH-20` extension contract |

All `REQ-ID`s and `CAP-ID`s are **stable across revisions**. New items append at the next free number per domain; existing IDs are never renumbered or reused.

### Priority Definitions

| Priority | Label | Definition |
|:---|:---|:---|
| **P0** | Must Have | Platform cannot go live without this. Blocking for the phase's GA. (Maps to PDD "P1 — Critical".) |
| **P1** | Should Have | Critical for tier-1 positioning; required within the phase but not GA-blocking. (Maps to PDD "P1".) |
| **P2** | Nice to Have | Productivity / depth improvement. Planned for next iteration. (Maps to PDD "P2".) |
| **P3** | Future | Strategic capability. Planned for a later phase. (Maps to PDD "P3".) |

### Phase Definitions

Phase 0 is the demo-winning pilot; the production shared core is delivered in Phase 1 and extended per LOB thereafter.

| Phase | Label | Scope |
|:---|:---|:---|
| **Phase 0** | **Pilot / Design-Partner MVP** | A thin, end-to-end **Motor** vertical slice on synthetic + sandbox data — quote → STP underwriting → Telebirr payment → bilingual issuance (schedule + COI + QR sticker) → mobile FNOL → fast-track settlement — built to **win the EIC contract**. Full multi-repo microservices + DDD/hexagonal/EDA/CQRS/outbox/saga, scoped to the demo's service set (not all 20). ~10–12 weeks. See [§34.1](#341-phase-0--pilot-mvp-detail). |
| **Phase 1** | Foundational Core + Motor (Production) | Shared core (Party, Product, Policy, Rating, UW, Claims, Billing, Docs, Workflow, Notifications, IAM, Audit, i18n) **hardened** end-to-end on **Motor** — full lifecycle, real rails, migration |
| **Phase 2** | Life | Life & group products; installment/whole-of-life billing; medical UW |
| **Phase 3** | Commercial Lines | Property, Marine, Liability, Engineering, Workmen's Compensation |
| **Phase 4** | Specialty + Intelligence | Bonds, Accident & Health, Travel; AI/ML (IDP, fraud scoring, damage estimation), agentic assist |

> **Phase 0 vs Phase 1.** Phase 0 is the *demo-winning* pilot: happy-path only, synthetic data, mocked/sandbox integrations — but built on the **production architecture** (multi-repo microservices, DDD, hexagonal/clean, EDA, CQRS, outbox, saga), scoped to the subset of services the Motor demo exercises. Phase 1 is the *production* build: full Motor lifecycle (endorsements/renewals/cancellations), installments, reconciliation, ERP sync, real Fayda/payment rails, data migration, edge cases. No Phase 0 work is throwaway — the pilot services are hardened and their thin slices thicken into the full REQ set.

### Capability ID Convention

Every capability carries a stable `CAP-{PREFIX}-{NUMBER}` ID (reused from the PDD where it exists; new modules introduce new prefixes). Advanced / forward-looking capabilities use an `A1`, `A2`, … suffix. Each functional `REQ` row lists the `CAP-ID`(s) it implements; Appendix B is the forward matrix (REQ → CAP), Appendix C the reverse matrix (CAP → REQ) with a `Full` / `Partial` / `Orphan` coverage flag.

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Product Vision & Objectives](#2-product-vision--objectives)
3. [Stakeholders & Personas](#3-stakeholders--personas)
4. [Product Scope & Boundaries](#4-product-scope--boundaries)
5. [Architectural Foundations & Shared-Core Extensibility](#5-architectural-foundations--shared-core-extensibility)
6. [Bidirectional Traceability Framework](#6-bidirectional-traceability-framework)
7. [Module: Party & Customer Management (PTY)](#7-module-party--customer-management-pty) — `BC-MDH-01`
8. [Module: Product Definition Engine (PRD)](#8-module-product-definition-engine-prd) — `BC-MDH-02`
9. [Module: Policy Administration (POL)](#9-module-policy-administration-pol) — `BC-MDH-03`
10. [Module: Rating & Premium Calculation (RAT)](#10-module-rating--premium-calculation-rat) — `BC-MDH-04`
11. [Module: Underwriting (UW)](#11-module-underwriting-uw) — `BC-MDH-05`
12. [Module: Claims Management (CLM)](#12-module-claims-management-clm) — `BC-MDH-06`
13. [Module: Billing & Payments (BIL)](#13-module-billing--payments-bil) — `BC-MDH-07`
14. [Module: Document Management (DOC)](#14-module-document-management-doc) — `BC-MDH-08`
15. [Module: Workflow & Approvals (WFA)](#15-module-workflow--approvals-wfa) — `BC-MDH-09`
16. [Module: Notifications (NOT)](#16-module-notifications-not) — `BC-MDH-10`
17. [Module: Reporting & Analytics (RPT)](#17-module-reporting--analytics-rpt) — `BC-MDH-11`
18. [Module: Reinsurance & Coinsurance (RIN)](#18-module-reinsurance--coinsurance-rin) — `BC-MDH-12`
19. [Module: Complaints & Disputes (CPL)](#19-module-complaints--disputes-cpl--new) — `BC-MDH-13` *(new)*
20. [Module: Financial Crime — Fraud / AML / Sanctions (FIN)](#20-module-financial-crime--fraud--aml--sanctions-fin--new) — `BC-MDH-14` *(new)*
21. [Module: Commission Management (COM)](#21-module-commission-management-com--new) — `BC-MDH-15` *(new)*
22. [Module: Identity & Access Management (IAM)](#22-module-identity--access-management-iam) — `BC-MDH-16`
23. [Module: Audit & Compliance (AUD)](#23-module-audit--compliance-aud) — `BC-MDH-17`
24. [Module: Integration & Anti-Corruption Layer (INT)](#24-module-integration--anti-corruption-layer-int--new) — `BC-MDH-18` *(new)*
25. [Module: Observability & Telemetry (OBS)](#25-module-observability--telemetry-obs--new) — `BC-MDH-19` *(new)*
26. [Module: Shared-Core Kernel & Product Extensibility (CORE)](#26-module-shared-core-kernel--product-extensibility-core--new) — `BC-MDH-20` *(new)*
27. [Cross-Cutting Non-Functional Requirements (NFR)](#27-cross-cutting-non-functional-requirements-nfr)
28. [Security Requirements (SEC)](#28-security-requirements-sec)
29. [Data Requirements (DAT)](#29-data-requirements-dat)
30. [Compliance & Regulatory Requirements (CMP)](#30-compliance--regulatory-requirements-cmp)
31. [Acceptance-Criteria Framework](#31-acceptance-criteria-framework)
32. [Risks & Mitigations](#32-risks--mitigations)
33. [Dependencies & Constraints](#33-dependencies--constraints)
34. [Phased-Delivery View](#34-phased-delivery-view)
35. [Appendices](#35-appendices) — REQ→CAP, CAP→REQ coverage, Regulatory→REQ, NFR/SLO, glossary

> Product-line extensions (Motor, Life, Property, Marine, Liability, Engineering) are delivered via the `CORE` extension contract (§26) and product configuration (§8); LOB-specific risk schemas & content are captured there rather than as separate PRD sections.

---

## 1. Executive Summary

### 1.1 Purpose

This PRD defines the complete functional and non-functional requirements for **Medhen** — an enterprise-grade, multi-product insurance automation platform for the **Ethiopian Insurance Corporation (EIC)**. Medhen digitizes the entire insurance value chain (onboarding → quote → underwriting → binding → payment → issuance → servicing → claims → settlement → renewal) across all lines of business, adopting **UK insurance procedures and conduct standards** as the operational blueprint while integrating natively with Ethiopian payment rails, regulation (NBE), language (Amharic), and administrative structures.

### 1.2 Product Statement

> Medhen is the reusable, regulator-defensible, future-proof insurance platform whose **shared core** (party, product, policy, rating, underwriting, claims, billing) lets EIC launch a new insurance product line — motor, life, property, marine — through **configuration, not code**, while every cross-cutting concern (identity, audit, documents, workflow, notifications, compliance) is owned once and consumed by all product lines.

### 1.3 Business Objectives

| # | Objective | Success Metric | Target |
|:---|:---|:---|:---|
| BO-1 | Automate the end-to-end insurance lifecycle | % of policy transactions executed straight-through (no manual re-keying) | ≥ 80% for standard Motor risks by Phase 1 GA |
| BO-2 | Launch new products via configuration, not code | Engineering effort to launch a new rateable product variant | 0 code changes for a variant of an existing LOB |
| BO-3 | Reduce quote-to-bind cycle time | Median time from quote creation to policy issuance (standard risk) | < 15 minutes online / same-day at branch |
| BO-4 | Accelerate claims settlement | Median FNOL-to-settlement for fast-track motor claims | < 5 business days |
| BO-5 | Improve premium collection | Collected / billed premium ratio | ≥ 95% within 90 days |
| BO-6 | UK-conduct & NBE regulatory defensibility | Regulatory findings attributable to system gaps | 0 |
| BO-7 | Operate at tier-1 reliability | Aggregate platform availability (business hours) | ≥ 99.9% (target; see NFR) |
| BO-8 | Full auditability | % of state changes with immutable, replayable audit trail | 100% |

### 1.4 Design Principles

| Principle | Description |
|:---|:---|
| **Configuration over Code** | New products launch through product configuration and versioned rules, not code changes |
| **Shared Core, Product Extensions** | One platform core powers all LOBs via a plugin/extension architecture (see §5) |
| **UK Procedure Alignment** | Workflows follow FCA Consumer Duty, ICOBS/IDD, Insurance Act 2015, Lloyd's/UK market practice |
| **Ethiopian Localization** | Native ETB, Ethiopian calendar, Amharic, local payment rails, NBE + Ethiopian data-protection compliance |
| **Event-Driven** | All state changes emit domain events enabling real-time downstream processing |
| **Bi-Temporal Auditability** | Complete historical record of every policy state at any point in time |
| **API-First** | Every capability exposed via documented REST/gRPC APIs |

### 1.5 Document Scope

| In Scope | Out of Scope |
|:---|:---|
| Functional + non-functional requirements for all modules & LOBs | Detailed low-level design (lives in per-service specs/ADRs — to be created) |
| Bidirectional traceability REQ ↔ CAP ↔ Module ↔ Regulatory | Implementation code |
| Given-When-Then acceptance criteria for every functional REQ | Contract test suites & fixtures |
| Compliance citations (NBE, FCA-conduct, Ethiopian DPP) | Regulator-facing filing artefacts |
| Shared-core extensibility contract | Product-line business content (rate values, rule thresholds — supplied by EIC) |

---

## 2. Product Vision & Objectives

### 2.1 Vision Statement

To be the **reference insurance platform for the Ethiopian market** — combining tier-1 engineering rigor (DDD, Hexagonal/Clean architecture, EDA, CQRS-where-justified, Outbox, Saga, bi-temporal modeling) with productized reusability (one shared core, many product lines, configuration-driven) and international competitiveness (UK conduct alignment, IFRS-17-ready finance, AI/ML-ready data platform).

### 2.2 Product Principles

| # | Principle | Implication |
|:---|:---|:---|
| PP-1 | **Core Independence** | The shared core evolves independently of any single product line; no LOB can force a core breaking change. |
| PP-2 | **Product Extensibility** | Every core context exposes stable extension points (risk schema, rating, rules, documents, claim types) consumed by product plugins. |
| PP-3 | **Backward-Compatible Evolution** | Changes are additive; existing product lines never break. Breaking changes require a major version + co-existence window. |
| PP-4 | **Single Source of Truth** | Every cross-cutting concern (identity, audit, documents, notifications) is owned by exactly one module. Product lines consume, never duplicate. |
| PP-5 | **Regulator-Defensible by Architecture** | Audit, retention, residency, and access control are architectural primitives, not afterthoughts. |
| PP-6 | **Observable, Idempotent, Resumable** | Every service emits standard telemetry; every external operation is idempotent; every workflow is resumable. |
| PP-7 | **Bilingual & Localized by Default** | English + Amharic and Ethiopian localization are first-class, not retrofitted. |

### 2.3 Supported Lines of Business

| # | Line of Business | Code | Phase |
|:---|:---|:---|:---|
| 1 | Motor Insurance | `MOTOR` | 1 |
| 2 | Life Insurance | `LIFE` | 2 |
| 3 | Property / Fire | `PROPERTY` | 3 |
| 4 | Marine | `MARINE` | 3 |
| 5 | Liability | `LIABILITY` | 3 |
| 6 | Engineering | `ENGINEERING` | 3 |
| 7 | Workmen's Compensation | `WORKMEN_COMP` | 3 |
| 8 | Bonds & Guarantees | `BONDS` | 4 |
| 9 | Accident & Health | `ACCIDENT_HEALTH` | 4 |
| 10 | Travel | `TRAVEL` | 4 |

---

## 3. Stakeholders & Personas

### 3.1 Primary Personas

#### Persona 1: Customer / Policyholder

| Attribute | Detail |
|:---|:---|
| **Name** | Abebe — private vehicle owner, Addis Ababa |
| **Goal** | Buy and renew motor cover online, pay via Telebirr, file a claim from his phone |
| **Pain Points** | Branch queues, opaque premiums, slow claim settlement, paper documents |
| **Success Criteria** | Self-service quote-buy-pay in minutes; digital policy & sticker; claim status transparency |
| **Primary REQ domains** | `PTY`, `POL`, `BIL`, `CLM`, `NOT` |

#### Persona 2: Insurance Agent / Broker

| Attribute | Detail |
|:---|:---|
| **Name** | Selamawit — tied agent managing a personal-lines book |
| **Goal** | Quote, bind, endorse, and renew policies for clients quickly; track commission |
| **Pain Points** | Re-keying customer data, manual rate lookups, unclear commission statements |
| **Success Criteria** | Guided quote wizard, instant rating, commission transparency, client 360 view |
| **Primary REQ domains** | `PTY`, `POL`, `RAT`, `COM` |

#### Persona 3: Underwriter

| Attribute | Detail |
|:---|:---|
| **Name** | Dawit — Senior Underwriter, Motor |
| **Goal** | Clear referrals within authority, apply loadings/conditions, decline bad risks |
| **Pain Points** | Manual risk assessment, unclear authority limits, no claims-history visibility |
| **Success Criteria** | Auto-STP for standard risks; referral workbench; authority matrix enforced; audit of every decision |
| **Primary REQ domains** | `UW`, `PRD`, `POL`, `WFA` |

#### Persona 4: Claims Adjuster / Handler

| Attribute | Detail |
|:---|:---|
| **Name** | Meron — Claims Adjuster |
| **Goal** | Register FNOL, investigate, set reserves, propose settlement within authority |
| **Pain Points** | Paper files, no reserve discipline, manual settlement math, fraud blind spots |
| **Success Criteria** | Digital claim file, guided investigation checklist, reserve controls, fraud indicators, fast-track routing |
| **Primary REQ domains** | `CLM`, `FIN`, `WFA`, `DOC` |

#### Persona 5: Finance Officer

| Attribute | Detail |
|:---|:---|
| **Name** | Kebede — Finance Officer |
| **Goal** | Reconcile payments, process refunds, sync journals to ERP, close the period |
| **Pain Points** | Unmatched payments, manual ERP entries, IFRS-17 data gaps |
| **Success Criteria** | Automated reconciliation, ERP event sync, refund controls, aged-premium reporting |
| **Primary REQ domains** | `BIL`, `RPT`, `INT`, `COM` |

#### Persona 6: Compliance Officer / MLRO

| Attribute | Detail |
|:---|:---|
| **Name** | Tigist — Chief Compliance Officer |
| **Goal** | Defend EIC under NBE examination; ensure fair-value & conduct compliance; manage complaints & AML |
| **Pain Points** | Audit gaps, manual regulatory returns, no sanctions screening, no complaints MI |
| **Success Criteria** | Immutable replayable audit; NBE returns auto-generated; sanctions/PEP screening; DISP-compliant complaints |
| **Primary REQ domains** | `AUD`, `CMP`, `FIN`, `CPL`, `RPT` |

#### Persona 7: Product Manager

| Attribute | Detail |
|:---|:---|
| **Name** | Nardos — Product Manager |
| **Goal** | Launch and version products, coverages, rate tables, and UW rules without engineering |
| **Pain Points** | Code-release dependency for pricing/rule changes; no product versioning |
| **Success Criteria** | Config-driven product lifecycle; versioned rates & rules with effective dating; clone-to-create |
| **Primary REQ domains** | `PRD`, `RAT`, `UW`, `DOC`, `CORE` |

#### Persona 8: Platform / Product Engineer

| Attribute | Detail |
|:---|:---|
| **Name** | Yonas — Backend Engineer |
| **Goal** | Add a new LOB by implementing the product-extension contract, not touching the core |
| **Pain Points** | Hidden coupling, breaking changes, undocumented extension points |
| **Success Criteria** | Stable core ports/events; documented product-plugin SDK; additive schema evolution enforced in CI |
| **Primary REQ domains** | `CORE`, all module extension points, `NFR`, `INT` |

### 3.2 Secondary Personas

| Persona | Touch Points |
|:---|:---|
| **Branch Manager** | Branch-scoped approvals, production/collection reporting, staff oversight |
| **Executive / Director** | Executive dashboard — GWP, loss ratio, retention, combined ratio |
| **Reinsurance Officer** | Treaty configuration, cession register, bordereaux, recoveries |
| **Regulator (NBE)** | Statutory returns, motor third-party reporting, examination data egress |
| **System Administrator** | User/role management, configuration, monitoring |

---

## 4. Product Scope & Boundaries

### 4.1 In Scope (Owned by the Medhen Platform)

The platform owns the modules enumerated in the domain table (§Requirement ID Convention). Each module is the sole owner of its capability surface; product lines consume module capabilities without duplicating them.

| Capability Family | Modules |
|:---|:---|
| Parties & relationships | `PTY` |
| Product configuration | `PRD`, `CORE` |
| Contract lifecycle | `POL`, `RAT`, `UW` |
| Claims & recovery | `CLM`, `FIN` (fraud) |
| Financial operations | `BIL`, `COM`, `RIN` |
| Servicing & conduct | `DOC`, `WFA`, `NOT`, `CPL` |
| Insight & governance | `RPT`, `AUD`, `CMP` |
| Platform primitives | `IAM`, `I18N`, `BRN` |

### 4.2 Out of Scope (Not Built by Medhen)

| Concern | Owner | Why |
|:---|:---|:---|
| Rate values, rule thresholds, product wordings | EIC (business content) | Supplied as configuration/data, not engineered |
| Core ERP / general ledger | Existing EIC ERP | Medhen integrates via events + API sync (`INT`) |
| National ID issuance / verification source | Fayda (national digital ID) | Medhen integrates as a consumer (`INT`) |
| Bank / mobile-money ledgers | Telebirr, CBE, Dashen, banks | Medhen integrates via payment gateways (`INT`) |
| Actuarial pricing model development | EIC actuarial function | Medhen consumes resulting rate tables & factors |

### 4.3 Assumptions

| # | Assumption |
|:---|:---|
| A-1 | EIC will supply product content (rates, rules, wordings, authority limits) as configuration inputs. |
| A-2 | Telebirr, CBE Birr, and Amole expose stable payment APIs with sandbox environments. |
| A-3 | The existing EIC ERP can consume events and/or expose a sync API for journal entries. |
| A-4 | Fayda national digital ID APIs are available for KYC identity verification. |
| A-5 | NBE will recognize Medhen's audit, retention, and reporting posture under examination. |
| A-6 | Data residency: production data is hosted within Ethiopia (on-prem or local cloud) per regulation. |

### 4.4 Constraints

| # | Constraint |
|:---|:---|
| C-1 | Amharic (Ge'ez script) and English are both first-class; all customer-facing documents render bilingually. |
| C-2 | Motor third-party liability is compulsory (NBE Regulation 554/2024); certificate & sticker are legally mandated. |
| C-3 | Financial records retained ≥ 7 years; audit trail immutable for regulatory examination. |
| C-4 | ETB currency; VAT 15% and stamp duty per Ethiopian tax law applied at rating. |
| C-5 | Ethiopian calendar must be supported for display and, where legally relevant, for policy-period computation. |

---

## 5. Architectural Foundations & Shared-Core Extensibility

> This section is the engineering heart of the platform — how a **single shared core** serves **many product lines** (motor, life, asset) through **configuration and plugins rather than code forks**. It is expanded further in the [`CORE` module requirements (§26)](#26-module-shared-core-kernel--product-extensibility-core--new) and per-module extension points.

### 5.1 Architecture Style

| Decision | Choice |
|:---|:---|
| Backend | Go microservices |
| Frontend | Next.js (App Router) |
| Database | PostgreSQL (per-service), Redis (cache/state) |
| Architecture | DDD, Hexagonal / Clean Architecture |
| Communication | EDA (Apache Kafka) for async; gRPC for sync hot-path; REST at the edge |
| Patterns | CQRS (read-side projections), Outbox, Saga (orchestration), Idempotency keys |
| Repository | Multi-repo (one per service) |
| Identity | Keycloak (self-hosted, OIDC) |
| Object storage | MinIO (documents) |

### 5.2 The Shared-Core / Product-Plugin Model

The platform is decomposed into a **stable core** and **product extensions**:

- **Stable core contexts** (`PTY`, `POL`, `RAT`, `UW`, `CLM`, `BIL`, `DOC`, `WFA`, `NOT`, `IAM`, `AUD`) own product-agnostic concepts: a *party* is a party whether motor or life; a *policy* has a lifecycle independent of what it insures; a *claim* has FNOL→settlement regardless of peril.
- **Product extensions** (motor, life, property …) supply the product-specific surface **through declared extension points**, never by forking core code:

| Extension Point | What a product plugin declares | Owned/validated by |
|:---|:---|:---|
| **Risk schema** | The insured-item data model (e.g., motor vehicle fields vs. life-assured fields) as a versioned JSON-Schema | `CORE` + `PRD` |
| **Rating logic** | Rate tables, factors, loadings, discounts as versioned configuration | `RAT` + `PRD` |
| **Underwriting rules** | Auto-accept / refer / decline rules as versioned decision logic | `UW` + `PRD` |
| **Coverages** | Coverage catalog, limits, deductibles, exclusions, sub-limits | `PRD` |
| **Claim types** | Loss-type taxonomy, investigation checklists, settlement rules | `CLM` + `PRD` |
| **Documents** | Templates & merge fields (schedule, certificate, endorsement) | `DOC` + `PRD` |

This is what makes **`Configuration over Code`** real: a new *variant* of an existing LOB is pure configuration; a genuinely new LOB implements the extension contract (schema + adapters) without modifying the core aggregates. Detailed requirements are in the `CORE`, `PRD`, `RAT`, and `UW` modules.

### 5.3 Cross-Cutting Guarantees

| Guarantee | Mechanism |
|:---|:---|
| Reliable eventing | Outbox pattern + Kafka; idempotent consumers; schema registry with backward-compatibility CI gate |
| Consistency across services | Saga orchestration for multi-service workflows (e.g., bind → invoice → cession) |
| Bi-temporal history | Policy versions tracked in business-time + system-time; no data loss on amendment |
| Auditability | Immutable, append-only audit of every data change and action |
| Tenancy/branch scoping | Branch-level access control and data partitioning |

---

## 6. Bidirectional Traceability Framework

### 6.1 The Linkage

Every functional requirement participates in a traceability chain:

```
Regulatory driver  →  REQ-{DOMAIN}-{NUM}  →  CAP-{PREFIX}-{NUM}  →  Module  →  (future) ADR / Service-Spec
```

- Each `REQ` row carries an inline **`Capabilities`** column listing the `CAP-ID`(s) it implements.
- Each `REQ` row carries **`Priority`**, **`Phase`**, **`Acceptance Criteria`** (Given-When-Then), **`Dependencies`**, and **`Notes`** columns.
- **Appendix B** — forward matrix (REQ → CAP).
- **Appendix C** — reverse matrix (CAP → REQ) with coverage flag (`Full` / `Partial` / `Orphan`); orphan capabilities surface for review.
- **Appendix G** — Regulatory → REQ reverse matrix (NBE, FCA-conduct, Ethiopian DPP, tax).

> [!NOTE]
> Medhen has no ADR or service-spec corpus yet. The `ADR` and `Service-Spec` columns used in the InnoGuard PRD are **deferred**; the equivalent columns here are `Dependencies` and `Notes`. When architecture decisions are recorded, an ADR column and Appendix D/F will be added without renumbering any REQ.

### 6.2 Coverage Commitments

| Commitment | Rule |
|:---|:---|
| No orphan capabilities | Every `CAP-ID` maps to ≥ 1 `REQ`; orphans flagged in Appendix C |
| No untraceable requirements | Every `REQ` maps to ≥ 1 `CAP-ID` |
| Regulatory coverage | Every compliance driver in §CMP maps to ≥ 1 `REQ` (Appendix G) |
| Stability | IDs never renumbered; additive-only evolution |

### 6.3 Acceptance-Criteria Standard

Every functional `REQ` states acceptance criteria in **Given-When-Then** form so QA can derive contract/integration tests directly. NFR requirements state measurable targets (latency, throughput, availability) instead.

---

## 7. Module: Party & Customer Management (PTY)

> **Bounded Context:** [`BC-MDH-01`](./Medhen-Platform-Capability-Document.md#bc-mdh-01--party--customer-management-pc-party-mgmt-svc) · **Service:** `pc-party-mgmt-svc` · **Database:** `pc_party_db`
> **Capability source:** Cap doc Part III, BC-MDH-01 · **PDD source:** [§3 Module 1](./product_definition_document.md) · **Phase:** 1 (core)

### 7.0 Mission Statement

`PTY` is the **central registry** of every individual and organization that interacts with the platform — policyholders, insureds, beneficiaries, agents, brokers, adjusters, service providers, reinsurers. It follows the ACORD party model (one party, many roles), owns identity/KYC verification, and is the authority for the Customer-360 view consumed by every other module. It is product-agnostic: the same party record underlies a motor, life, or property relationship.

### 7.1 Capability Scope

| CAP-ID | Capability | REQ-IDs (Forward Map) | Status |
|:---|:---|:---|:---|
| `CAP-PARTY-001` | Party registration (individual + organization) | `REQ-PTY-001`, `REQ-PTY-002`, `REQ-PTY-003`, `REQ-PTY-004` | Phase 1 |
| `CAP-PARTY-002` | Party profile management | `REQ-PTY-010`, `REQ-PTY-011`, `REQ-PTY-012`, `REQ-PTY-013`, `REQ-PTY-014`, `REQ-PTY-015` | Phase 1 |
| `CAP-PARTY-003` | KYC (Know Your Customer) | `REQ-PTY-020`, `REQ-PTY-021`, `REQ-PTY-022`, `REQ-PTY-023` | Phase 1 |
| `CAP-PARTY-004` | Party roles | `REQ-PTY-030`, `REQ-PTY-031`, `REQ-PTY-032` | Phase 1 |
| `CAP-PARTY-005` | Party search & lookup | `REQ-PTY-040`, `REQ-PTY-041`, `REQ-PTY-042` | Phase 1 |
| `CAP-PARTY-006` | Customer 360° view | `REQ-PTY-050`, `REQ-PTY-051`, `REQ-PTY-052` | Phase 1 |
| `CAP-PARTY-A1` | Consent & data-subject rights (Ethiopian DPP) | `REQ-PTY-060` | Phase 1 — *enhancement* |
| `CAP-PARTY-A2` | Vulnerable-customer flagging (UK Consumer Duty) | `REQ-PTY-061` | Phase 2 — *enhancement* |
| `CAP-PARTY-A3` | Sanctions/PEP screening hook at onboarding | `REQ-PTY-062` | Phase 2 — *enhancement (delegates to FIN)* |

**Inbound:** IAM (user↔party linkage), Fayda national ID (identity verification), branch onboarding.
**Outbound to modules:** `POL` (policyholder/insured), `CLM` (claimant/third-party), `BIL` (payer/payee), `COM` (agent/broker), `FIN` (screening subject), `RPT` (party dimensions).

### 7.2 Functional Requirements

#### 7.2.1 Party Registration

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-001` | The service SHALL register an **individual** party capturing name, date of birth, gender, national ID (Fayda/Kebele), and tax ID (TIN), persisting a unique `party_id`. | `CAP-PARTY-001` | P0 | 1 | **Given** valid individual details, **when** registration is submitted, **then** a party record is created with a unique `party_id`, type `INDIVIDUAL`, and a `PartyCreated` event is published within 1s. | Fayda connector (`INT`) | — |
| `REQ-PTY-002` | The service SHALL register an **organization** party capturing legal name, registration number, TIN, and industry classification. | `CAP-PARTY-001` | P0 | 1 | **Given** valid organization details, **when** submitted, **then** a party of type `ORGANIZATION` is created and `PartyCreated` published. | — | — |
| `REQ-PTY-003` | The service SHALL run automated **duplicate detection** on registration using national ID (exact), phone (exact), and name+DOB (fuzzy) matching, and warn the operator before creating a probable duplicate. | `CAP-PARTY-001` | P0 | 1 | **Given** a new registration whose national ID or fuzzy name+DOB matches an existing party above threshold, **when** submitted, **then** the candidate match(es) are returned with a confidence score and the operator must confirm or link before proceeding. | — | Fuzzy match reused by `PTY-041` |
| `REQ-PTY-004` | The service SHALL **merge** two party records identified as duplicates, re-pointing all relationships (policies, claims, payments) to the surviving party, with a full, reversible audit trail. | `CAP-PARTY-001` | P1 | 1 | **Given** two confirmed-duplicate parties, **when** a merge is executed, **then** all child relationships re-point to the survivor, a `PartyMerged` event is emitted, and the merge is recorded immutably in the audit log with before/after state. | `AUD` | — |

#### 7.2.2 Party Profile Management

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-010` | The service SHALL allow updates to party personal/organizational details with full change tracking (who, what, when, old→new). | `CAP-PARTY-002` | P0 | 1 | **Given** an edit to a party field, **when** saved, **then** the change is persisted, a `PartyUpdated` event is emitted, and the audit log records old and new values. | `AUD` | — |
| `REQ-PTY-011` | The service SHALL manage addresses (Home, Work, Mailing) using Ethiopian administrative units **Region → Zone → Woreda → Kebele**, with one primary address per type. | `CAP-PARTY-002` | P0 | 1 | **Given** an address with valid admin-unit hierarchy, **when** added, **then** it is stored and validated against the reference admin-unit dataset; setting a new primary demotes the previous primary. | Reference data (`DAT`) | — |
| `REQ-PTY-012` | The service SHALL manage contact methods (Mobile, Phone, Email, Fax) with primary designation and format validation (E.164 phone, RFC-5322 email). | `CAP-PARTY-002` | P0 | 1 | **Given** a contact value, **when** added, **then** it is format-validated; an invalid value is rejected with a field-level error. | — | — |
| `REQ-PTY-013` | The service SHALL record party bank accounts (CBE, Dashen, Awash, etc.) for premium collection and claim/refund settlement. | `CAP-PARTY-002` | P1 | 1 | **Given** bank account details, **when** added, **then** they are stored, masked in UI/logs, and available to `BIL`/`CLM` for settlement. | `SEC` (masking) | — |
| `REQ-PTY-014` | The service SHALL manage party status: `ACTIVE`, `SUSPENDED`, `DEACTIVATED`, `BLACKLISTED`, with a reason on every transition. | `CAP-PARTY-002` | P1 | 1 | **Given** a status change with reason, **when** applied, **then** the new status is enforced across modules (e.g., a `BLACKLISTED` party cannot bind a new policy) and audited. | `AUD` | Enforcement points in `POL`, `UW` |
| `REQ-PTY-015` | The service SHALL allow upload and management of a party profile photograph. | `CAP-PARTY-002` | P3 | 1 | **Given** an image within size/type limits, **when** uploaded, **then** it is stored in object storage and linked to the party. | `DOC`/MinIO | — |

#### 7.2.3 KYC (Know Your Customer)

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-020` | The service SHALL accept upload of KYC documents: National/Kebele ID, Passport, Driving License, Business Registration, Trade License, each with type, number, issue/expiry dates. | `CAP-PARTY-003` | P0 | 1 | **Given** a KYC document, **when** uploaded, **then** it is stored with metadata and the party's KYC status becomes `PENDING`. | `DOC` | — |
| `REQ-PTY-021` | The service SHALL provide a manual **verification workflow** for uploaded KYC documents with approve/reject and reason. | `CAP-PARTY-003` | P0 | 1 | **Given** a pending KYC document, **when** an authorized verifier approves it, **then** the document is marked `VERIFIED`; on reject, a reason is captured and the customer is notified. | `WFA`, `NOT` | — |
| `REQ-PTY-022` | The service SHALL track document **expiry** and trigger renewal reminders ahead of expiry. | `CAP-PARTY-003` | P1 | 1 | **Given** a KYC document with an expiry date, **when** the configured lead time is reached, **then** a reminder notification is scheduled and the document is flagged as expiring. | `NOT` | — |
| `REQ-PTY-023` | The service SHALL maintain an overall **KYC status** per party: `PENDING`, `VERIFIED`, `EXPIRED`, `REJECTED`, derived from constituent documents. | `CAP-PARTY-003` | P0 | 1 | **Given** a party's KYC documents, **when** their states change, **then** the aggregate KYC status is recomputed and exposed to `POL`/`UW` as an underwriting/binding gate. | — | Gate consumed by `POL-bind` |

#### 7.2.4 Party Roles

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-030` | The service SHALL assign one or more roles to a party: Policyholder, Insured, Beneficiary, Agent, Broker, Adjuster, Service Provider, Reinsurer. | `CAP-PARTY-004` | P0 | 1 | **Given** a party, **when** a role is assigned, **then** the role is recorded with an effective date and the party appears in role-scoped queries (e.g., agent lists). | — | One party, many roles (ACORD) |
| `REQ-PTY-031` | The service SHALL capture role-specific attributes (e.g., agent license number, broker accreditation, adjuster specialization). | `CAP-PARTY-004` | P1 | 1 | **Given** a role assignment, **when** role-specific attributes are entered, **then** they are validated per role schema and stored. | `CORE` (role schema) | — |
| `REQ-PTY-032` | The service SHALL manage agent/broker records including licensing, territory assignment, and commission-structure linkage. | `CAP-PARTY-004` | P1 | 1 | **Given** an agent/broker party, **when** licensing/territory/commission is set, **then** it is available to `COM` for commission calculation and to `POL` for producer assignment. | `COM` | — |

#### 7.2.5 Party Search & Lookup

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-040` | The service SHALL provide basic search by name, national ID, phone, or email. | `CAP-PARTY-005` | P0 | 1 | **Given** a search term, **when** submitted, **then** matching parties return within P95 ≤ 500ms, paginated and branch-scoped by the caller's authorization. | `IAM` | — |
| `REQ-PTY-041` | The service SHALL provide advanced multi-field search with filters (type, role, status, city, date range). | `CAP-PARTY-005` | P1 | 1 | **Given** multiple filter criteria, **when** submitted, **then** only parties matching all criteria return, respecting branch scope. | — | — |
| `REQ-PTY-042` | The service SHALL provide full-text search across party fields via a search index. | `CAP-PARTY-005` | P2 | 2 | **Given** a free-text query, **when** submitted, **then** ranked results return from the search index within P95 ≤ 800ms. | Search index (`DAT`) | Elasticsearch or equivalent |

#### 7.2.6 Customer 360° View

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-050` | The service SHALL present a consolidated **Customer 360** view: profile, policies, claims, payments, documents, and interactions. | `CAP-PARTY-006` | P0 | 1 | **Given** a party, **when** the 360 view is opened, **then** it aggregates data from `POL`, `CLM`, `BIL`, `DOC` via read-side projections within P95 ≤ 2s. | `RPT` (projections) | CQRS read-side |
| `REQ-PTY-051` | The service SHALL display a **relationship map** (family, business connections, agent-client). | `CAP-PARTY-006` | P3 | 3 | **Given** a party with relationships, **when** the map is opened, **then** connected parties and relationship types are visualized. | — | — |
| `REQ-PTY-052` | The service SHALL present a chronological **interaction history** (calls, emails, claims, payments, policy changes). | `CAP-PARTY-006` | P2 | 2 | **Given** a party, **when** the timeline is opened, **then** all recorded interactions display in reverse-chronological order with source module. | `RPT`, `NOT` | — |

#### 7.2.7 Enhancements (Conduct, Privacy & Screening)

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PTY-060` | The service SHALL capture and version **consent** for data processing and marketing, and support **data-subject requests** (access, rectification, erasure-where-lawful) per the Ethiopian Personal Data Protection Proclamation. | `CAP-PARTY-A1` | P0 | 1 | **Given** a party, **when** consent is granted/withdrawn, **then** the versioned consent state is recorded and enforced (e.g., marketing suppressed on withdrawal); **given** an erasure request, **when** no legal-retention obligation applies, **then** PII is redacted while retaining audit-required financial records. | `AUD`, `CMP`, `NOT` | Resolves PDD gap; retention conflict handled per `CMP` |
| `REQ-PTY-061` | The service SHALL support flagging a party as a **vulnerable customer** with the driver captured, so downstream journeys can apply enhanced-care handling (UK Consumer Duty). | `CAP-PARTY-A2` | P1 | 2 | **Given** an authorized user, **when** a vulnerability flag and reason are set, **then** the flag is surfaced (access-controlled) to servicing/claims journeys and audited; the flag is never exposed to unauthorized roles. | `IAM`, `AUD` | Resolves PDD gap |
| `REQ-PTY-062` | The service SHALL invoke **sanctions/PEP screening** at party onboarding and on material profile change, delegating to the `FIN` module, and gate binding on unresolved hits. | `CAP-PARTY-A3` | P1 | 2 | **Given** a new or updated party, **when** screening runs, **then** a screening result is recorded; a positive hit creates a `FIN` case and blocks policy binding until dispositioned. | `FIN` | Resolves PDD gap |

### 7.3 Module-Level NFRs

| Dimension | Target | Notes |
|:---|:---|:---|
| Party lookup P95 latency | ≤ 500ms | Basic search; 360 view ≤ 2s |
| Registration throughput | ≥ 50 registrations/sec sustained | Branch + online onboarding peaks |
| Duplicate-detection recall | ≥ 95% on labelled set | Fuzzy name+DOB + exact ID |
| Availability | ≥ 99.9% | Core dependency of every journey |
| PII protection | Encrypted at rest (AES-256), masked in logs | Bank accounts, national ID |

### 7.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` Policy | `REQ-PTY-001`, `REQ-PTY-023`, `REQ-PTY-030` | Policyholder/insured identity + KYC gate at bind |
| `CLM` Claims | `REQ-PTY-001`, `REQ-PTY-030` | Claimant & third-party identity |
| `BIL` Billing | `REQ-PTY-013` | Payer/payee bank accounts |
| `COM` Commission | `REQ-PTY-032` | Agent/broker & commission structure |
| `FIN` Financial Crime | `REQ-PTY-062` | Screening subject |
| `RPT` Reporting | `REQ-PTY-010`, `REQ-PTY-050` | Party dimensions & 360 projection |

### 7.5 Open Questions / Risks

- **OQ-PTY-1.** Is Fayda the authoritative identity source for individuals, or is Kebele ID accepted as fallback where Fayda enrolment is incomplete? → Confirm with EIC/NBE; affects `REQ-PTY-001` verification strength.
- **OQ-PTY-2.** Erasure vs. 7-year financial retention conflict — confirm the lawful-basis matrix with Compliance so `REQ-PTY-060` redaction scope is precise.
- **OQ-PTY-3.** Should organization parties support hierarchical group structures (parent/subsidiary) in Phase 1 or defer to Phase 3 (commercial lines)?

---

## 8. Module: Product Definition Engine (PRD)

> **Bounded Context:** [`BC-MDH-02`](./Medhen-Platform-Capability-Document.md#bc-mdh-02--product-definition-engine-pc-product-defn-svc) · **Service:** `pc-product-defn-svc` · **Database:** `pc_product_db`
> **Capability source:** Cap doc Part III, BC-MDH-02 · **PDD source:** [§4 Module 2](./product_definition_document.md) · **Phase:** 1 (core)

### 8.0 Mission Statement

`PRD` is the **factory** that configures all insurance products through versioned, effective-dated configuration rather than code. It is a control-plane context (§6 Cap doc): product changes are deliberate, approved, and immutable once published; the data plane loads the *effective* version at transaction time. It is the mechanism that makes *Configuration over Code* real.

### 8.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-PROD-001` | Product lifecycle management | `REQ-PRD-001`, `REQ-PRD-002`, `REQ-PRD-003`, `REQ-PRD-004`, `REQ-PRD-005` | Phase 1 |
| `CAP-PROD-002` | Coverage configuration | `REQ-PRD-010`, `REQ-PRD-011`, `REQ-PRD-012`, `REQ-PRD-013` | Phase 1 |
| `CAP-PROD-003` | Rate table management | `REQ-PRD-020`, `REQ-PRD-021`, `REQ-PRD-022` | Phase 1 |
| `CAP-PROD-004` | Underwriting rule configuration | `REQ-PRD-030`, `REQ-PRD-031` | Phase 1 |
| `CAP-PROD-005` | Document template association | `REQ-PRD-040` | Phase 1 |
| `CAP-PROD-A1` | Product-line extension registration | `REQ-PRD-050` | per-LOB |
| `CAP-PROD-A2` | Fair-value / product governance record | `REQ-PRD-051` | Phase 2 |

### 8.2 Functional Requirements

#### 8.2.1 Product Lifecycle

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PRD-001` | The service SHALL create an insurance product defining LOB, name, description, and effective dates. | `CAP-PROD-001` | P0 | 1 | **Given** valid product details, **when** created, **then** a product is persisted in `DRAFT` with a unique `product_id`. | `CORE` | — |
| `REQ-PRD-002` | The service SHALL version product definitions with full change history; each published version is immutable. | `CAP-PROD-001` | P0 | 1 | **Given** an active product, **when** a new version is published, **then** the prior version is retained immutably and quotes reference the version effective at quote time. | — | — |
| `REQ-PRD-003` | The service SHALL enforce the status lifecycle `DRAFT → REVIEW → APPROVED → ACTIVE → SUSPENDED → RETIRED`. | `CAP-PROD-001` | P0 | 1 | **Given** a product in a state, **when** a transition is requested, **then** only permitted transitions succeed; illegal transitions are rejected and audited. | `WFA`, `AUD` | — |
| `REQ-PRD-004` | The service SHALL clone an existing product as the basis for a new variant. | `CAP-PROD-001` | P2 | 1 | **Given** a source product, **when** cloned, **then** a new `DRAFT` product is created copying coverages, rates, rules, and templates. | — | — |
| `REQ-PRD-005` | The service SHALL apply effective dating; only active-and-effective products are quotable. | `CAP-PROD-001` | P1 | 1 | **Given** a product with `effective_from`/`effective_to`, **when** a quote requests it outside the window, **then** it is not offered. | `POL` | — |

#### 8.2.2 Coverage, Rate & Rule Configuration

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PRD-010` | The service SHALL configure coverages per product (code, name, mandatory/optional). | `CAP-PROD-002` | P0 | 1 | **Given** a product, **when** a coverage is defined, **then** it is available for selection in quoting per its mandatory/optional flag. | — | Compulsory motor TP marked mandatory |
| `REQ-PRD-011` | The service SHALL set coverage limits (default/min/max) and deductibles (flat / % of sum insured / % of loss). | `CAP-PROD-002` | P0 | 1 | **Given** a coverage, **when** limits/deductibles are set, **then** rating and quoting enforce the bounds. | `RAT` | — |
| `REQ-PRD-012` | The service SHALL define coverage dependencies, exclusions, and sub-limits. | `CAP-PROD-002` | P1 | 1 | **Given** dependent coverages, **when** a dependent is selected without its parent, **then** selection is rejected with an explanatory error. | — | — |
| `REQ-PRD-013` | The service SHALL define standard exclusions per coverage as structured text rendered into documents. | `CAP-PROD-002` | P1 | 1 | **Given** exclusions, **when** a policy schedule is generated, **then** applicable exclusions appear in the document. | `DOC` | — |
| `REQ-PRD-020` | The service SHALL create rate tables with multi-dimensional lookup keys and rating factors (multiplicative/additive/percentage). | `CAP-PROD-003` | P0 | 1 | **Given** a rate table with dimensions, **when** rating looks up a risk, **then** the correct cell/factor is returned deterministically. | `RAT` | — |
| `REQ-PRD-021` | The service SHALL version rate tables with effective dating; existing quotes retain their original rates. | `CAP-PROD-003` | P1 | 1 | **Given** a re-rated table, **when** an in-flight quote recalculates, **then** it uses the rate version effective at quote creation. | — | — |
| `REQ-PRD-022` | The service SHALL bulk-import rate tables from CSV/Excel with validation. | `CAP-PROD-003` | P2 | 1 | **Given** a rate file, **when** imported, **then** rows are validated and errors reported per-row without partial commit. | — | — |
| `REQ-PRD-030` | The service SHALL author auto-accept, referral-trigger, and decline rules per product. | `CAP-PROD-004` | P0 | 1 | **Given** configured rules, **when** underwriting evaluates a submission, **then** the rules produce accept/refer/decline consistently. | `UW`, `CORE` | — |
| `REQ-PRD-031` | The service SHALL define evaluation order/priority for conflicting rules. | `CAP-PROD-004` | P2 | 1 | **Given** conflicting rules, **when** evaluated, **then** the highest-priority rule wins deterministically. | — | — |
| `REQ-PRD-040` | The service SHALL associate document templates and product-specific merge fields to products. | `CAP-PROD-005` | P1 | 1 | **Given** a product, **when** a template is associated, **then** document generation resolves the correct template & merge fields. | `DOC` | — |

#### 8.2.3 Enhancements (Extensibility & Governance)

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-PRD-050` | The service SHALL register a new product-line extension (risk schema, rating adapter, UW-rule bindings, claim taxonomy, merge-fields) via the Kernel extension contract without core code change. | `CAP-PROD-A1` | P1 | per-LOB | **Given** a new LOB extension, **when** registered and activated, **then** quoting/binding/claims operate for that LOB with no modification to core aggregates. | `CORE` (`BC-MDH-20`) | Realizes shared-core goal |
| `REQ-PRD-051` | The service SHALL attach a fair-value / product-governance assessment to each product version (UK Consumer Duty / PROD). | `CAP-PROD-A2` | P1 | 2 | **Given** a product version, **when** submitted for approval, **then** a fair-value assessment record is required and stored; absence blocks activation. | `WFA`, `CMP` | Resolves PDD gap |

### 8.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Effective-product resolution (gRPC) | P95 ≤ 50ms (cached) |
| Config publish propagation | < 60s to data-plane caches |
| Availability | ≥ 99.9% |

### 8.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` Policy | `REQ-PRD-005`, `REQ-PRD-010` | Effective product & coverages for quoting |
| `RAT` Rating | `REQ-PRD-020`, `REQ-PRD-021` | Rate tables & factors |
| `UW` Underwriting | `REQ-PRD-030`, `REQ-PRD-031` | UW rules |
| `DOC` Documents | `REQ-PRD-040` | Template & merge-field bindings |

### 8.5 Open Questions / Risks

- **OQ-PRD-1.** Who owns product-approval authority (Product Manager vs. committee) and what is the maker-checker chain? → Confirm with EIC to configure `REQ-PRD-003` workflow.
- **OQ-PRD-2.** Rate-table authoring — spreadsheet import (`REQ-PRD-022`) vs. in-app editor as primary path for actuarial team?

---

## 9. Module: Policy Administration (POL)

> **Bounded Context:** [`BC-MDH-03`](./Medhen-Platform-Capability-Document.md#bc-mdh-03--policy-administration-pc-policy-svc) · **Service:** `pc-policy-svc` · **Database:** `pc_policy_db`
> **Capability source:** Cap doc Part III, BC-MDH-03 · **PDD source:** [§5 Module 3](./product_definition_document.md) · **Phase:** 1 (core)

### 9.0 Mission Statement

`POL` owns the complete policy lifecycle — quotation → binding → issuance → endorsement → renewal → cancellation → expiry/lapse — with **bi-temporal versioning** so any state is reconstructable at any business-time and system-time. It is LOB-agnostic; LOB specifics arrive via the risk schema from `CORE`.

### 9.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-POL-001` | Quotation management | `REQ-POL-001`…`REQ-POL-010` | Phase 1 |
| `CAP-POL-002` | Policy binding | `REQ-POL-020`…`REQ-POL-025` | Phase 1 |
| `CAP-POL-003` | Policy issuance | `REQ-POL-030`…`REQ-POL-034` | Phase 1 |
| `CAP-POL-004` | Policy servicing (endorsements) | `REQ-POL-040`…`REQ-POL-051` | Phase 1 |
| `CAP-POL-005` | Policy renewal | `REQ-POL-060`…`REQ-POL-069` | Phase 1 |
| `CAP-POL-006` | Policy cancellation | `REQ-POL-070`…`REQ-POL-076` | Phase 1 |
| `CAP-POL-007` | Policy search & retrieval | `REQ-POL-080`…`REQ-POL-085` | Phase 1 |
| `CAP-POL-A1` | Statutory cooling-off / cancellation rights | `REQ-POL-090` | Phase 2 |
| `CAP-POL-A3` | Coinsurance sharing | `REQ-POL-091` | Phase 3 |

### 9.2 Functional Requirements

#### 9.2.1 Quotation

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-POL-001` | The service SHALL create a quote by selecting product and entering customer + risk details via a guided multi-step wizard (Customer → Product → Risk → Coverage → Summary). | `CAP-POL-001` | P0 | 1 | **Given** a product and risk data valid against the LOB risk schema, **when** the wizard completes, **then** a `DRAFT` quote is created with a unique `quote_id`. | `PTY`, `PRD`, `CORE` | — |
| `REQ-POL-002` | The service SHALL calculate the premium by invoking the Rating engine and display an itemized breakdown. | `CAP-POL-001` | P0 | 1 | **Given** a quote, **when** premium is calculated, **then** the breakdown (base + factors + loadings − discounts + taxes) is returned and stored with the rate-table version. | `RAT` | — |
| `REQ-POL-003` | The service SHALL clear the quote through Underwriting (auto or referral) before it becomes `QUOTED`. | `CAP-POL-001` | P0 | 1 | **Given** a calculated quote, **when** UW auto-accepts, **then** the quote becomes `QUOTED`; **when** referred, **then** it becomes `REFERRED` pending decision. | `UW` | — |
| `REQ-POL-004` | The service SHALL enforce a configurable quote validity period (default 30 days) and auto-expire quotes after it. | `CAP-POL-001` | P1 | 1 | **Given** a `QUOTED` quote past validity, **when** the expiry job runs, **then** it transitions to `EXPIRED` and emits `QuoteExpired`. | — | — |
| `REQ-POL-005` | The service SHALL support re-quote (modify details and recalculate) and side-by-side quote comparison. | `CAP-POL-001` | P1 | 1 | **Given** a quote, **when** edited, **then** the premium recalculates using the originally-effective rate version; comparison shows options side by side. | `RAT` | — |
| `REQ-POL-006` | The service SHALL generate a printable quote summary document. | `CAP-POL-001` | P1 | 1 | **Given** a `QUOTED` quote, **when** the document is requested, **then** a quote summary PDF is generated bilingually. | `DOC` | — |
| `REQ-POL-007` | The service SHALL support quick quote, quote assignment to an agent/underwriter, and internal quote notes. | `CAP-POL-001` | P2 | 1 | **Given** minimal input for a standard risk, **when** quick-quote is used, **then** an indicative premium is produced; quotes can be assigned and annotated. | — | — |

#### 9.2.2 Binding & Issuance

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-POL-020` | The service SHALL accept a quote and verify payment (full premium or minimum down payment) before binding. | `CAP-POL-002` | P0 | 1 | **Given** an accepted quote, **when** payment is confirmed by Billing, **then** binding proceeds; **when** not, **then** the policy stays `ACCEPTED`. | `BIL` | Saga |
| `REQ-POL-021` | The service SHALL bind a policy from an accepted quote, generating a policy number `EIC/{LOB}/{BRANCH}/{YYYY}/{SEQ}`. | `CAP-POL-002` | P0 | 1 | **Given** payment confirmed and KYC verified, **when** bind executes, **then** an `ACTIVE`/`BOUND` policy with a unique number is created and `PolicyBound` emitted. | `PTY` (KYC gate) | — |
| `REQ-POL-022` | The service SHALL issue a temporary cover note immediately on binding and support a future effective date. | `CAP-POL-002` | P1 | 1 | **Given** a bound policy, **when** requested, **then** a cover note is issued; a future `effective_date` defers cover start. | `DOC` | — |
| `REQ-POL-030` | The service SHALL generate the policy schedule, Certificate of Insurance, and (motor) windshield sticker on issuance. | `CAP-POL-003` | P0 | 1 | **Given** a bound motor policy, **when** issued, **then** schedule + COI + sticker are generated and `PolicyIssued` emitted. | `DOC` | NBE Reg. 554/2024 |
| `REQ-POL-031` | The service SHALL deliver policy documents via email, SMS link, or portal download. | `CAP-POL-003` | P1 | 1 | **Given** issued documents, **when** delivery runs, **then** the customer receives links per preference and delivery is tracked. | `NOT` | — |
| `REQ-POL-032` | The service SHALL support batch policy issuance for fleet/group policies. | `CAP-POL-003` | P2 | 2 | **Given** a fleet upload, **when** batch-issued, **then** member policies/certificates are generated with per-item results. | — | — |

#### 9.2.3 Endorsements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-POL-040` | The service SHALL initiate mid-term endorsements: change insured item, add/remove coverage, change sum insured, change policyholder, add/remove insured party, change address. | `CAP-POL-004` | P0 | 1 | **Given** an active policy, **when** an endorsement is requested, **then** the change is captured and a premium impact computed. | `RAT` | — |
| `REQ-POL-041` | The service SHALL compute pro-rata-temporis additional premium or refund for mid-term changes. | `CAP-POL-004` | P0 | 1 | **Given** a mid-term change, **when** priced, **then** the pro-rata additional/refund amount is calculated for the unexpired period. | `RAT` | — |
| `REQ-POL-042` | The service SHALL route endorsements exceeding auto-approve thresholds through approval. | `CAP-POL-004` | P1 | 1 | **Given** an endorsement above threshold, **when** submitted, **then** it routes to workflow approval before taking effect. | `WFA` | — |
| `REQ-POL-043` | The service SHALL create a new bi-temporal policy version and generate an endorsement schedule showing changes. | `CAP-POL-004` | P0 | 1 | **Given** an approved endorsement, **when** applied, **then** a new version records business-time + system-time with no loss of prior state, and an endorsement document is generated. | `DOC`, `AUD` | — |
| `REQ-POL-044` | The service SHALL provide endorsement history with before/after comparison. | `CAP-POL-004` | P1 | 1 | **Given** a policy, **when** history is viewed, **then** all endorsements display with before/after values. | — | — |

#### 9.2.4 Renewal, Cancellation & Search

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-POL-060` | The service SHALL identify policies due for renewal at configurable lead times (60/45/30 days). | `CAP-POL-005` | P0 | 1 | **Given** approaching expiry, **when** the renewal job runs, **then** due policies are flagged `PENDING_RENEWAL`. | — | — |
| `REQ-POL-061` | The service SHALL auto-generate a renewal quote using current rate tables, applying NCD and loss-experience loading. | `CAP-POL-005` | P0 | 1 | **Given** a renewable policy, **when** re-rated, **then** NCD and loss loading are applied and a renewal quote produced. | `RAT` | — |
| `REQ-POL-062` | The service SHALL send renewal invitations and process renewal acceptance (with or without changes) and payment. | `CAP-POL-005` | P1 | 1 | **Given** a renewal quote, **when** the customer accepts and pays, **then** a new term is bound; changes are permitted before acceptance. | `NOT`, `BIL` | — |
| `REQ-POL-063` | The service SHALL handle non-renewal (insurer or policyholder) and lapse policies not renewed within grace. | `CAP-POL-005` | P1 | 1 | **Given** an unrenewed policy past grace, **when** the lapse job runs, **then** it transitions to `LAPSED` with reason. | — | — |
| `REQ-POL-064` | The service SHALL produce a renewal pipeline report (upcoming, status, retention rate). | `CAP-POL-005` | P1 | 1 | **Given** a period, **when** the report runs, **then** it shows renewals due, converted, and retention %. | `RPT` | — |
| `REQ-POL-070` | The service SHALL process cancellation with reason capture, pro-rata/short-rate refund, notice period, and open-claims check. | `CAP-POL-006` | P0 | 1 | **Given** a cancellation request, **when** processed, **then** the correct return premium is computed, notice enforced, and a warning raised if open claims exist. | `RAT`, `BIL` | — |
| `REQ-POL-071` | The service SHALL route cancellations through approval and generate a cancellation confirmation + debit/credit notes. | `CAP-POL-006` | P1 | 1 | **Given** an approved cancellation, **when** finalized, **then** documents are generated and `PolicyCancelled` emitted with refund amount. | `WFA`, `DOC`, `BIL` | — |
| `REQ-POL-080` | The service SHALL support search by policy number, customer, insured item (plate/chassis), and advanced multi-criteria filters, plus a comprehensive detail view and timeline. | `CAP-POL-007` | P0 | 1 | **Given** a query, **when** executed, **then** matching policies return branch-scoped within P95 ≤ 500ms; the detail view aggregates coverages, history, claims, payments. | `IAM` | — |

#### 9.2.5 Enhancements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-POL-090` | The service SHALL enforce a statutory cooling-off / cancellation-rights window with the correct refund basis (UK conduct). | `CAP-POL-A1` | P1 | 2 | **Given** a new policy within the cooling-off window, **when** the customer cancels, **then** the statutory refund basis applies and is documented. | `RAT`, `CMP` | Resolves PDD gap |
| `REQ-POL-091` | The service SHALL record coinsurance shares (lead/follow, share %) on large risks. | `CAP-POL-A3` | P2 | 3 | **Given** a coinsured risk, **when** bound, **then** co-insurer shares are recorded and reflected in settlement/reporting. | `RIN` | Resolves PDD gap |

### 9.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Policy lookup | P95 ≤ 500ms |
| Bind transaction (saga) | ≤ 5s end-to-end |
| Bi-temporal integrity | Zero data loss; any-point reconstruction |
| Availability | ≥ 99.9% (Tier-0) |

### 9.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `BIL` Billing | `REQ-POL-021`, `REQ-POL-041`, `REQ-POL-070` | Invoice/refund on bind/endorse/cancel |
| `DOC` Documents | `REQ-POL-030`, `REQ-POL-043` | Document generation triggers |
| `RIN` Reinsurance | `REQ-POL-021`, `REQ-POL-091` | Cession & coinsurance |
| `COM` Commission | `REQ-POL-021`, `REQ-POL-062` | Commission earning events |
| `RPT` Reporting | `REQ-POL-064`, `REQ-POL-080` | Production & renewal projections |

### 9.5 Open Questions / Risks

- **OQ-POL-1.** Grace-period and lapse rules per LOB — confirm EIC policy so `REQ-POL-063` is configured correctly.
- **OQ-POL-2.** Does the Ethiopian calendar affect statutory policy-period computation, or is Gregorian authoritative for contract dates? → Affects `CORE` calendar primitives.

---

## 10. Module: Rating & Premium Calculation (RAT)

> **Bounded Context:** [`BC-MDH-04`](./Medhen-Platform-Capability-Document.md#bc-mdh-04--rating--premium-calculation-pc-rating-calc-svc) · **Service:** `pc-rating-calc-svc` · **Database:** `pc_rating_db` (rate cache)
> **Capability source:** Cap doc Part III, BC-MDH-04 · **PDD source:** [§6 Module 4](./product_definition_document.md) · **Phase:** 1 (core)

### 10.0 Mission Statement

`RAT` is a **stateless** calculation service that, given risk data + effective product configuration, returns a fully itemized, reproducible premium breakdown. It is the pricing authority for quotes, endorsements, cancellations, and renewals, and records every calculation for audit.

### 10.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-RATE-001` | Premium calculation | `REQ-RAT-001`…`REQ-RAT-008` | Phase 1 |
| `CAP-RATE-002` | Rating audit & transparency | `REQ-RAT-010`, `REQ-RAT-011`, `REQ-RAT-012` | Phase 1 |
| `CAP-RATE-003` | Pro-rata calculations | `REQ-RAT-020`, `REQ-RAT-021`, `REQ-RAT-022` | Phase 1 |

### 10.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-RAT-001` | The service SHALL calculate a full premium breakdown given risk data + product code. | `CAP-RATE-001` | P0 | 1 | **Given** risk data + effective product, **when** invoked, **then** a breakdown (per-coverage net, loadings, discounts, taxes, gross) is returned in < 1s. | `PRD` | — |
| `REQ-RAT-002` | The service SHALL look up base rates from product rate tables using risk dimensions. | `CAP-RATE-001` | P0 | 1 | **Given** a rate table and risk dimensions, **when** looked up, **then** the correct base rate is returned deterministically. | `PRD`, `CORE` | — |
| `REQ-RAT-003` | The service SHALL apply multiplicative and additive rating factors. | `CAP-RATE-001` | P0 | 1 | **Given** configured factors, **when** applied, **then** the premium reflects each factor in the documented order. | — | — |
| `REQ-RAT-004` | The service SHALL apply loadings (adverse claims, high-value, hazards) and discounts (NCD, multi-policy, group, fleet, loyalty). | `CAP-RATE-001` | P1 | 1 | **Given** eligibility, **when** rated, **then** loadings and discounts appear as itemized lines. | — | — |
| `REQ-RAT-005` | The service SHALL enforce min/max premium bounds per product/coverage. | `CAP-RATE-001` | P1 | 1 | **Given** a computed premium outside bounds, **when** finalized, **then** it is clamped to the bound and flagged. | — | — |
| `REQ-RAT-006` | The service SHALL calculate applicable taxes (VAT 15%, stamp duty). | `CAP-RATE-001` | P0 | 1 | **Given** a net premium, **when** taxed, **then** VAT and stamp duty are computed per Ethiopian tax rules and itemized. | `CORE` | — |
| `REQ-RAT-010` | The service SHALL record every calculation (inputs, factors applied, result) linked to the rate-table version used. | `CAP-RATE-002` | P0 | 1 | **Given** any calculation, **when** completed, **then** an immutable `RatingCalculation` audit record with the rate-table version is stored. | `AUD` | — |
| `REQ-RAT-011` | The service SHALL support what-if calculation without creating a quote. | `CAP-RATE-002` | P2 | 1 | **Given** hypothetical risk data, **when** what-if is invoked, **then** a breakdown is returned without persisting a quote. | — | — |
| `REQ-RAT-020` | The service SHALL calculate pro-rata-temporis additional premium/refund for endorsements. | `CAP-RATE-003` | P0 | 1 | **Given** a mid-term change with unexpired days, **when** computed, **then** the pro-rata amount is returned. | — | — |
| `REQ-RAT-021` | The service SHALL calculate cancellation return premium (pro-rata or short-rate per policy terms). | `CAP-RATE-003` | P0 | 1 | **Given** a cancellation method, **when** computed, **then** the correct return premium is returned. | — | — |
| `REQ-RAT-022` | The service SHALL calculate renewal premium with NCD and loss-experience adjustments. | `CAP-RATE-003` | P1 | 1 | **Given** a renewing policy, **when** re-rated, **then** NCD and loss loading are applied to current rates. | — | — |

### 10.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Premium calculation | < 1s (P95 ≤ 300ms typical) |
| Statelessness | Horizontally scalable; no session state |
| Reproducibility | 100% for same inputs + rate version |
| Availability | ≥ 99.9% (gates quoting) |

### 10.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` Policy | `REQ-RAT-001`, `REQ-RAT-020`, `REQ-RAT-021`, `REQ-RAT-022` | Quote/endorse/cancel/renew pricing |
| `AUD` Audit | `REQ-RAT-010` | Calculation audit records |

### 10.5 Open Questions / Risks

- **OQ-RAT-1.** Short-rate cancellation scale — confirm EIC's short-rate table for `REQ-RAT-021`.
- **OQ-RAT-2.** Rounding conventions for ETB and tax — confirm to standardize `CORE` money primitives.

---

## 11. Module: Underwriting (UW)

> **Bounded Context:** [`BC-MDH-05`](./Medhen-Platform-Capability-Document.md#bc-mdh-05--underwriting-pc-underwriting-svc) · **Service:** `pc-underwriting-svc` · **Database:** `pc_underwriting_db`
> **Capability source:** Cap doc Part III, BC-MDH-05 · **PDD source:** [§7 Module 5](./product_definition_document.md) · **Phase:** 1 (core)

### 11.0 Mission Statement

`UW` automates risk assessment, enforces the **authority matrix**, and manages the **referral workflow** — auto-accept (STP), auto-decline, or refer, with every decision audited and conditions captured, aligned with UK fair-presentation and delegated-authority practice.

### 11.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-UW-001` | Automated risk assessment | `REQ-UW-001`…`REQ-UW-006` | Phase 1 |
| `CAP-UW-002` | Referral management | `REQ-UW-010`…`REQ-UW-015` | Phase 1 |
| `CAP-UW-003` | Authority matrix | `REQ-UW-020`…`REQ-UW-023` | Phase 1 |
| `CAP-UW-A1` | Fair-presentation / disclosure capture | `REQ-UW-030` | Phase 2 |

### 11.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-UW-001` | The service SHALL evaluate a submission against product-configured underwriting rules automatically. | `CAP-UW-001` | P0 | 1 | **Given** a quote, **when** assessed, **then** the rules yield accept/refer/decline with reasons, inline with quoting. | `PRD` | — |
| `REQ-UW-002` | The service SHALL auto-accept (STP) standard risks matching all auto-accept criteria. | `CAP-UW-001` | P0 | 1 | **Given** a risk meeting auto-accept, **when** assessed, **then** it is approved without referral and the quote proceeds. | — | STP target ≥ 80% motor |
| `REQ-UW-003` | The service SHALL auto-decline submissions matching hard decline rules. | `CAP-UW-001` | P1 | 1 | **Given** a risk matching a decline rule, **when** assessed, **then** it is declined with a reason and `QuoteDeclined` emitted. | — | — |
| `REQ-UW-004` | The service SHALL auto-generate a referral when referral conditions are triggered. | `CAP-UW-001` | P0 | 1 | **Given** a risk triggering referral (e.g., sum insured > threshold), **when** assessed, **then** a referral is created at the required authority level. | `WFA` | — |
| `REQ-UW-005` | The service SHALL check the proposer's claims history across policies. | `CAP-UW-001` | P1 | 1 | **Given** a proposer, **when** assessed, **then** prior claims are retrieved and factored into referral/loading. | `PTY`, `CLM` | — |
| `REQ-UW-006` | The service SHALL compute a numeric risk score for prioritization. | `CAP-UW-001` | P2 | 1 | **Given** weighted factors, **when** scored, **then** a risk score is attached to the assessment. | — | — |
| `REQ-UW-010` | The service SHALL create referrals with reason, risk details, and required authority level, presented in an underwriter workbench/queue. | `CAP-UW-002` | P0 | 1 | **Given** a triggered referral, **when** created, **then** it appears in the assigned underwriter's queue with full context. | `WFA` | — |
| `REQ-UW-011` | The service SHALL record referral decisions: approve / approve-with-conditions / decline / refer-higher, capturing conditions (exclusions, loadings, warranties). | `CAP-UW-002` | P0 | 1 | **Given** a referral, **when** an underwriter decides, **then** the decision + conditions are stored and `ReferralDecisionMade` emitted. | `AUD` | — |
| `REQ-UW-012` | The service SHALL escalate referrals to higher authority when the current level cannot decide or SLA is breached. | `CAP-UW-002` | P1 | 1 | **Given** an SLA breach or refer-higher, **when** triggered, **then** the referral routes to the next authority level. | `WFA` | — |
| `REQ-UW-013` | The service SHALL track referral response time against configurable SLAs per authority level. | `CAP-UW-002` | P1 | 1 | **Given** a referral, **when** open, **then** elapsed time is tracked and breaches alerted. | `NOT` | — |
| `REQ-UW-020` | The service SHALL support configurable authority levels with premium and sum-insured limits per product. | `CAP-UW-003` | P0 | 1 | **Given** authority config, **when** a decision is attempted beyond a user's limit, **then** it is blocked and must escalate. | `IAM` | — |
| `REQ-UW-021` | The service SHALL assign authority levels to individual underwriters and support committee referral with decision tracking. | `CAP-UW-003` | P1 | 1 | **Given** an exceptional risk, **when** referred to committee, **then** the committee decision is recorded with participants. | — | — |
| `REQ-UW-022` | The service SHALL maintain a full authority audit (who approved what, at what level, with what conditions). | `CAP-UW-003` | P0 | 1 | **Given** any UW decision, **when** made, **then** an immutable audit record captures authority level and conditions. | `AUD` | — |
| `REQ-UW-030` | The service SHALL capture the proposer's duty-of-fair-presentation / disclosure as a structured record. | `CAP-UW-A1` | P1 | 2 | **Given** a submission, **when** disclosures are captured, **then** they are stored and available for claims-time review (Insurance Act 2015). | `CMP` | Resolves PDD gap |

### 11.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Auto-assessment latency | < 1s (inline with quote) |
| STP rate (standard motor) | ≥ 80% |
| Referral SLA accuracy | 100%; breach alert < 1 min |
| Availability | ≥ 99.9% |

### 11.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` Policy | `REQ-UW-001`, `REQ-UW-011` | Quote advancement per decision |
| `WFA` Workflow | `REQ-UW-004`, `REQ-UW-012` | Referral routing/escalation |
| `RPT` Reporting | `REQ-UW-001`, `REQ-UW-022` | Conversion & referral-rate reporting |

### 11.5 Open Questions / Risks

- **OQ-UW-1.** Confirm the authority-matrix limits (PDD §7.2 provides a starting table) with EIC underwriting management.
- **OQ-UW-2.** External data sources (vehicle registration, prior-claims bureau) availability for `REQ-UW-005` enrichment — depends on `INT` connectors.

---

## 12. Module: Claims Management (CLM)

> **Bounded Context:** [`BC-MDH-06`](./Medhen-Platform-Capability-Document.md#bc-mdh-06--claims-management-pc-claims-svc) · **Service:** `pc-claims-svc` · **Database:** `pc_claims_db`
> **Capability source:** Cap doc Part III, BC-MDH-06 · **PDD source:** [§8 Module 6](./product_definition_document.md) · **Phase:** 1 (core)

### 12.0 Mission Statement

`CLM` owns the full claims lifecycle following UK/London-market practice: FNOL → coverage check → triage → investigation → reserve → assessment → settlement → recovery → closure (with reopening). It enforces reserve discipline, settlement authority, and fast-track routing, and delegates fraud scoring to `FIN`.

### 12.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-CLM-001` | First Notice of Loss (FNOL) | `REQ-CLM-001`…`REQ-CLM-009` | Phase 1 |
| `CAP-CLM-002` | Triage & assignment | `REQ-CLM-010`…`REQ-CLM-014` | Phase 1 |
| `CAP-CLM-003` | Investigation | `REQ-CLM-020`…`REQ-CLM-025` | Phase 1 |
| `CAP-CLM-004` | Reserve management | `REQ-CLM-030`…`REQ-CLM-034` | Phase 1 |
| `CAP-CLM-005` | Assessment & settlement | `REQ-CLM-040`…`REQ-CLM-046` | Phase 1 |
| `CAP-CLM-006` | Claim payment | `REQ-CLM-050`…`REQ-CLM-054` | Phase 1 |
| `CAP-CLM-007` | Recovery (subrogation & salvage) | `REQ-CLM-060`…`REQ-CLM-063` | Phase 1 |
| `CAP-CLM-008` | Closure & reopening | `REQ-CLM-070`…`REQ-CLM-073` | Phase 1 |
| `CAP-CLM-A1` | AI damage estimation | `REQ-CLM-080` | Phase 4 |

### 12.2 Functional Requirements

#### 12.2.1 FNOL & Triage

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-CLM-001` | The service SHALL register a claim with loss details (date, time, location, description, parties) via web, mobile, phone, or branch. | `CAP-CLM-001` | P0 | 1 | **Given** loss details, **when** submitted through any channel, **then** a claim is created with a unique number `CLM/{LOB}/{BRANCH}/{YYYY}/{SEQ}` and `ClaimSubmitted` emitted. | — | — |
| `REQ-CLM-002` | The service SHALL look up the relevant policy by number, plate, or customer, and validate coverage on the loss date. | `CAP-CLM-001` | P0 | 1 | **Given** a claim, **when** coverage is validated, **then** the policy's active status on the loss date and loss-type coverage are confirmed; if not covered, the claim moves to `COVERAGE_DENIED`. | `POL` | — |
| `REQ-CLM-003` | The service SHALL classify the loss type per product taxonomy and capture third-party details and documents at FNOL, with optional GPS location. | `CAP-CLM-001` | P1 | 1 | **Given** FNOL, **when** completed, **then** loss type, third parties, uploaded evidence, and GPS (if provided) are recorded. | `DOC`, `PRD` | — |
| `REQ-CLM-010` | The service SHALL score claim severity and auto-assign to adjusters by specialization, workload, and location; supervisors may reassign. | `CAP-CLM-002` | P1 | 1 | **Given** a registered claim, **when** triaged, **then** a severity score is set and the claim is assigned; reassignment is permitted and audited. | — | — |
| `REQ-CLM-011` | The service SHALL route low-value, simple claims to fast-track settlement per configurable thresholds. | `CAP-CLM-002` | P2 | 1 | **Given** a claim under fast-track thresholds, **when** triaged, **then** it is routed to fast-track and the assigned adjuster is notified. | `NOT` | — |

#### 12.2.2 Investigation & Reserves

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-CLM-020` | The service SHALL provide a product-specific investigation checklist and manage claim documents (photos, police/medical reports, estimates, invoices). | `CAP-CLM-003` | P1 | 1 | **Given** a claim, **when** investigation begins, **then** the LOB checklist is presented and documents are uploaded/managed. | `DOC`, `PRD` | — |
| `REQ-CLM-021` | The service SHALL record site/vehicle inspection findings and third-party expert reports (surveyor, engineer, medical). | `CAP-CLM-003` | P1 | 1 | **Given** an investigation, **when** findings/reports are added, **then** they are stored with the claim. | — | — |
| `REQ-CLM-022` | The service SHALL flag potential fraud indicators and request a fraud score from `FIN`. | `CAP-CLM-003` | P2 | 2 | **Given** a claim, **when** fraud indicators trigger, **then** a `FIN` score is requested; a high score routes the claim to SIU. | `FIN` | — |
| `REQ-CLM-023` | The service SHALL maintain a chronological claim diary with notes from all handlers. | `CAP-CLM-003` | P0 | 1 | **Given** a claim, **when** a note is added, **then** it appears in the diary with author and timestamp, immutably. | `AUD` | — |
| `REQ-CLM-030` | The service SHALL set and adjust separate reserves (indemnity, expense, recovery) with reason and approver, subject to reserve-authority limits. | `CAP-CLM-004` | P0 | 1 | **Given** a claim, **when** a reserve is set/adjusted, **then** it is bounded by the handler's authority, requires a reason, and is audited with full history; `ReserveSet`/`ReserveAdjusted` emitted. | `AUD` | — |

#### 12.2.3 Assessment, Settlement, Payment, Recovery, Closure

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-CLM-040` | The service SHALL calculate settlement based on assessment, deductibles, policy limits, and depreciation. | `CAP-CLM-005` | P0 | 1 | **Given** an assessed claim, **when** calculated, **then** the settlement applies deductibles, limits, and depreciation correctly. | `POL` | — |
| `REQ-CLM-041` | The service SHALL let an adjuster propose a settlement with rationale and route it through approval by amount vs. authority. | `CAP-CLM-005` | P0 | 1 | **Given** a proposed settlement above authority, **when** submitted, **then** it routes to the correct approver; within authority, it is approved directly; `SettlementProposed`/`ClaimApproved` emitted. | `WFA` | — |
| `REQ-CLM-042` | The service SHALL support partial/interim settlements and generate settlement offer or denial letters. | `CAP-CLM-005` | P1 | 1 | **Given** a claim, **when** a partial settlement or denial is issued, **then** the corresponding document is generated and the claimant notified. | `DOC`, `NOT` | — |
| `REQ-CLM-043` | The service SHALL handle motor total loss (market value, write-off, salvage). | `CAP-CLM-005` | P1 | 1 | **Given** a total-loss motor claim, **when** assessed, **then** market value is set, write-off recorded, and salvage tracked. | — | — |
| `REQ-CLM-050` | The service SHALL process settlement payments to claimant or service provider, supporting multiple payees and methods (bank, Telebirr, CBE Birr, Amole, check), with approval above threshold. | `CAP-CLM-006` | P0 | 1 | **Given** an approved settlement, **when** payment is processed, **then** Billing disburses to the correct payee(s) via the selected method idempotently; `ClaimSettled` emitted. | `BIL` | — |
| `REQ-CLM-060` | The service SHALL identify and track subrogation recovery and manage salvage disposal, netting recoveries against claim cost. | `CAP-CLM-007` | P1 | 1 | **Given** a settled claim with third-party fault, **when** subrogation is pursued, **then** recovery efforts and amounts are tracked and netted; `RecoveryReceived` emitted. | `BIL`, `RIN` | — |
| `REQ-CLM-070` | The service SHALL close a claim with a final financial summary after a closure checklist, releasing remaining reserves; and reopen on new information. | `CAP-CLM-008` | P0 | 1 | **Given** a settled claim, **when** the closure checklist passes, **then** it is closed with a financial summary and reserves released; **when** new info arrives, **then** it can reopen to `UNDER_INVESTIGATION`. | — | — |
| `REQ-CLM-080` | The service SHALL estimate motor damage & repair cost from uploaded photos (AI-assisted, explainable, human-confirmed). | `CAP-CLM-A1` | P3 | 4 | **Given** claim photos, **when** AI estimation runs, **then** an indicative damage/cost estimate is produced for adjuster confirmation, never auto-final without review. | `DOC` | Resolves PDD gap (AI) |

### 12.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| FNOL registration | P95 ≤ 2s |
| Fast-track settlement | < 5 business days median |
| Reserve-change audit | 100% with reason & approver |
| Availability | ≥ 99.9% (Tier-0) |

### 12.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `BIL` Billing | `REQ-CLM-050`, `REQ-CLM-060` | Settlement payment & recovery accounting |
| `FIN` Fin-Crime | `REQ-CLM-022` | Fraud scoring |
| `RIN` Reinsurance | `REQ-CLM-030`, `REQ-CLM-060` | Reserve & recovery cession |
| `RPT` Reporting | `REQ-CLM-070` | Loss ratio, frequency/severity |

### 12.5 Open Questions / Risks

- **OQ-CLM-1.** Reserve-authority limits per adjuster seniority — confirm with EIC claims management.
- **OQ-CLM-2.** Approved service-provider (garage/hospital) network scope and direct-billing — Phase 3 (`CAP-CLM-A3`)?

---

## 13. Module: Billing & Payments (BIL)

> **Bounded Context:** [`BC-MDH-07`](./Medhen-Platform-Capability-Document.md#bc-mdh-07--billing--payments-pc-billing-svc) · **Service:** `pc-billing-svc` · **Database:** `pc_billing_db`
> **Capability source:** Cap doc Part III, BC-MDH-07 · **PDD source:** [§9 Module 7](./product_definition_document.md) · **Phase:** 1 (core)

### 13.0 Mission Statement

`BIL` owns the money — billing accounts, invoices, credit/debit notes, payments across Ethiopian rails, installments, refunds, reconciliation, and ERP journal sync. It is the financial-correctness authority (idempotent, auditable, reconciled) and drives policy transitions on payment confirmation and overdue.

### 13.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-BIL-001` | Billing account management | `REQ-BIL-001`…`REQ-BIL-004` | Phase 1 |
| `CAP-BIL-002` | Invoice management | `REQ-BIL-010`…`REQ-BIL-015` | Phase 1 |
| `CAP-BIL-003` | Payment processing | `REQ-BIL-020`…`REQ-BIL-028` | Phase 1 |
| `CAP-BIL-004` | Installment billing | `REQ-BIL-030`…`REQ-BIL-036` | Phase 1 |
| `CAP-BIL-005` | Refunds | `REQ-BIL-040`…`REQ-BIL-042` | Phase 1 |
| `CAP-BIL-006` | Reconciliation | `REQ-BIL-050`…`REQ-BIL-053` | Phase 1 |
| `CAP-BIL-007` | ERP integration | `REQ-BIL-060`…`REQ-BIL-063` | Phase 1 |
| `CAP-BIL-A3` | IFRS 17 revenue events | `REQ-BIL-070` | Phase 2 |

### 13.2 Functional Requirements

#### 13.2.1 Accounts, Invoices & Payments

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-BIL-001` | The service SHALL auto-create a billing account per policy on binding with a real-time balance (total, paid, outstanding, overdue). | `CAP-BIL-001` | P0 | 1 | **Given** `PolicyBound`, **when** consumed, **then** a billing account is created and balance reflects the invoice. | `POL` | — |
| `REQ-BIL-002` | The service SHALL generate account statements and full transaction history. | `CAP-BIL-001` | P1 | 1 | **Given** a date range, **when** requested, **then** a statement of all transactions is produced. | — | — |
| `REQ-BIL-010` | The service SHALL auto-generate itemized invoices on bind, endorsement, and renewal (per-coverage premiums, taxes, fees), with status tracking (Draft/Issued/Partially Paid/Paid/Overdue/Void) and PDF. | `CAP-BIL-002` | P0 | 1 | **Given** a billing event, **when** invoiced, **then** an itemized invoice is created, a PDF generated, and status maintained. | `DOC` | — |
| `REQ-BIL-011` | The service SHALL generate credit notes (return premium) and debit notes (additional premium). | `CAP-BIL-002` | P1 | 1 | **Given** an endorsement/cancellation, **when** the premium delta is computed, **then** the correct credit/debit note is generated. | `POL` | — |
| `REQ-BIL-020` | The service SHALL process payments via Telebirr, CBE Birr, and Amole through the Integration ACL. | `CAP-BIL-003` | P0 | 1 | **Given** a payment initiation, **when** sent via `INT`, **then** the gateway flow completes and status is recorded; success emits `PaymentReceived`. | `INT` | — |
| `REQ-BIL-021` | The service SHALL handle gateway callbacks idempotently with signature verification. | `CAP-BIL-003` | P0 | 1 | **Given** a duplicate/late callback, **when** received, **then** it is deduplicated and never double-posts; invalid signatures are rejected. | `INT`, `CORE` | Idempotency |
| `REQ-BIL-022` | The service SHALL record cash (with receipt), bank transfer (manual match), and check (clearing status) payments. | `CAP-BIL-003` | P0 | 1 | **Given** a branch cash payment, **when** recorded, **then** a receipt is generated; bank/check payments capture matching/clearing state. | `DOC` | — |
| `REQ-BIL-023` | The service SHALL allocate payments to invoices (supporting partial payments) and generate receipts. | `CAP-BIL-003` | P0 | 1 | **Given** a payment, **when** allocated, **then** it applies to the target invoice(s); partial payments update balances correctly. | — | — |
| `REQ-BIL-024` | The service SHALL reverse a payment with reason (bounced check, failed mobile money, error). | `CAP-BIL-003` | P1 | 1 | **Given** a recorded payment, **when** reversed, **then** balances restore and the reversal is audited. | `AUD` | — |

#### 13.2.2 Installments, Refunds, Reconciliation & ERP

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-BIL-030` | The service SHALL configure installment plans (monthly/quarterly/semi-annual) and auto-generate a schedule with due dates and a configurable down payment on bind. | `CAP-BIL-004` | P0 | 1 | **Given** an installment product, **when** bound, **then** a schedule with down payment and dues is generated. | `POL` | — |
| `REQ-BIL-031` | The service SHALL send installment reminders before due dates and manage overdue installments with grace period and optional late fee. | `CAP-BIL-004` | P0 | 1 | **Given** an upcoming/overdue installment, **when** the job runs, **then** reminders are sent and overdue status set; `PaymentOverdue` emitted after grace. | `NOT`, `POL` | — |
| `REQ-BIL-032` | The service SHALL recalculate remaining installments upon endorsement premium adjustment. | `CAP-BIL-004` | P1 | 1 | **Given** an endorsement, **when** premium changes, **then** the remaining installment schedule is recalculated. | `POL` | — |
| `REQ-BIL-040` | The service SHALL calculate and process refunds (cancellation/endorsement reduction) via original method or bank transfer, with approval above threshold. | `CAP-BIL-005` | P0 | 1 | **Given** a return premium, **when** refunded, **then** the payout is processed and `RefundProcessed` emitted; refunds above threshold require finance approval. | `WFA`, `INT` | — |
| `REQ-BIL-050` | The service SHALL reconcile incoming payments against outstanding invoices, import bank statements (CSV/MT940), and manage unallocated payments. | `CAP-BIL-006` | P0 | 1 | **Given** payments and invoices, **when** reconciliation runs, **then** matches are made, unmatched flagged, and a reconciliation report produced daily. | `INT` | — |
| `REQ-BIL-060` | The service SHALL sync journal entries (premium income, claim payments, commissions, refunds) to the EIC ERP via events + daily reconciliation. | `CAP-BIL-007` | P0 | 1 | **Given** a financial event, **when** synced, **then** a balanced journal entry reaches the ERP; `JournalEntryCreated` emitted; daily reconciliation confirms completeness. | `INT`, `COM` | — |
| `REQ-BIL-070` | The service SHALL emit revenue-recognition-grade events sufficient for IFRS 17 measurement. | `CAP-BIL-A3` | P1 | 2 | **Given** premium/claim financial events, **when** emitted, **then** Reporting can construct IFRS 17 contract-group measurements from them. | `RPT`, `CMP` | Resolves PDD gap |

### 13.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Payment initiation | P95 ≤ 1s (excl. external gateway) |
| Idempotency | 100% — duplicate callbacks never double-post |
| Reconciliation completeness | 100% daily; unallocated tracked |
| Financial correctness | Balanced journals; zero-loss ledger |
| Availability | ≥ 99.9% (Tier-0) |

### 13.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` Policy | `REQ-BIL-001`, `REQ-BIL-031` | Payment confirmation & overdue → policy state |
| `CLM` Claims | `REQ-BIL-020` | Settlement disbursement |
| `COM` Commission | `REQ-BIL-060` | Commission journal sync |
| `RPT` Reporting | `REQ-BIL-050`, `REQ-BIL-070` | Collection & IFRS 17 |

### 13.5 Open Questions / Risks

- **OQ-BIL-1.** ERP integration contract — event-driven vs. batch API, and the GL journal schema — confirm with EIC ERP team.
- **OQ-BIL-2.** Late-payment fee/interest policy (`REQ-BIL-031`) — confirm NBE/EIC rules.
- **OQ-BIL-3.** Client-money segregation expectations under NBE — affects account modelling.

---

## 14. Module: Document Management (DOC)

> **Bounded Context:** [`BC-MDH-08`](./Medhen-Platform-Capability-Document.md#bc-mdh-08--document-management-pc-document-mgmt-svc) · **Service:** `pc-document-mgmt-svc` · **Database:** `pc_document_db` + MinIO
> **Capability source:** Cap doc Part III, BC-MDH-08 · **PDD source:** [§10 Module 8](./product_definition_document.md) · **Phase:** 1 (core)

### 14.0 Mission Statement

`DOC` generates, stores, and serves every platform document — bilingually (English + Amharic) from managed templates — and stores uploaded documents with metadata and secure retrieval.

### 14.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-DOC-001..011` | Document generation catalog | `REQ-DOC-001`…`REQ-DOC-004` | Phase 1 |
| `CAP-DOC-012` | Bilingual rendering | `REQ-DOC-010` | Phase 1 |
| `CAP-DOC-013` | Template CRUD | `REQ-DOC-011` | Phase 1 |
| `CAP-DOC-014` | Document storage | `REQ-DOC-012` | Phase 1 |
| `CAP-DOC-A3` | QR-verifiable certificates | `REQ-DOC-020` | Phase 2 |
| `CAP-DOC-A2` | Intelligent document processing | `REQ-DOC-021` | Phase 4 |

### 14.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-DOC-001` | The service SHALL generate the policy-document catalog: schedule, Certificate of Insurance, cover note, motor sticker, endorsement schedule, renewal/cancellation notice, quote summary, debit/credit notes, invoice/receipt, claim letters. | `CAP-DOC-001..011` | P0 | 1 | **Given** a triggering event with data, **when** generation runs, **then** the correct document type is produced from its template within 5s. | `POL`, `BIL`, `CLM`, `PRD` | — |
| `REQ-DOC-002` | The service SHALL generate the motor windshield sticker and COI meeting the NBE legal format. | `CAP-DOC-001..011` | P0 | 1 | **Given** a bound motor policy, **when** issued, **then** a compliant sticker + COI are produced. | `POL` | NBE Reg. 554/2024 |
| `REQ-DOC-010` | The service SHALL render all customer-facing documents in English and Amharic (Noto Sans Ethiopic) per user preference. | `CAP-DOC-012` | P0 | 1 | **Given** a document and locale preference, **when** generated, **then** it renders correctly in the chosen language with proper Ge'ez script. | `CORE` (i18n) | — |
| `REQ-DOC-011` | The service SHALL provide template CRUD (HTML/CSS) with merge-field definitions via admin portal. | `CAP-DOC-013` | P1 | 1 | **Given** an admin, **when** a template is edited, **then** subsequent generations use the new version; merge fields validate against the product. | `PRD`, `IAM` | — |
| `REQ-DOC-012` | The service SHALL store generated and uploaded documents in MinIO with metadata, entity linkage, versioning, and access-controlled download. | `CAP-DOC-014` | P0 | 1 | **Given** a document, **when** stored, **then** it is retrievable by authorized users only, with metadata and version. | MinIO, `IAM` | — |
| `REQ-DOC-020` | The service SHALL embed a verification QR code on COIs/stickers for third-party & police verification. | `CAP-DOC-A3` | P1 | 2 | **Given** a certificate, **when** generated, **then** it carries a QR resolving to a verification endpoint confirming validity. | — | Resolves PDD gap |
| `REQ-DOC-021` | The service SHALL extract structured data from uploaded KYC/claim documents (OCR/IDP). | `CAP-DOC-A2` | P3 | 4 | **Given** an uploaded document, **when** IDP runs, **then** key fields are extracted for human confirmation. | — | Resolves PDD gap (AI) |

### 14.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Document generation | < 5s (P95) |
| Bilingual fidelity | 100% of customer-facing docs |
| Storage durability | No document loss; versioned |
| Availability | ≥ 99.9% |

### 14.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL` / `BIL` / `CLM` | `REQ-DOC-001` | Document generation on lifecycle events |
| `NOT` Notifications | `REQ-DOC-012` | Document links for delivery |

### 14.5 Open Questions / Risks

- **OQ-DOC-1.** Confirm the exact NBE-mandated sticker/COI layout for `REQ-DOC-002`.
- **OQ-DOC-2.** E-signature (`CAP-DOC-A1`) scope/provider for proposal acceptance — Phase 3.

---

## 15. Module: Workflow & Approvals (WFA)

> **Bounded Context:** [`BC-MDH-09`](./Medhen-Platform-Capability-Document.md#bc-mdh-09--workflow--approvals-pc-workflow-svc) · **Service:** `pc-workflow-svc` · **Database:** `pc_workflow_db`
> **Capability source:** Cap doc Part III, BC-MDH-09 · **PDD source:** [§11 Module 9](./product_definition_document.md) · **Phase:** 1 (core)

### 15.0 Mission Statement

`WFA` is the reusable **maker-checker & approval engine** any module uses for approvals — routing by role/authority/branch, supporting multi-level and parallel approval, delegation, SLA tracking, and escalation, with complete approval audit.

### 15.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-WF-001/002` | Define & initiate workflow | `REQ-WFA-001`, `REQ-WFA-002` | Phase 1 |
| `CAP-WF-003/004` | Routing & decisions | `REQ-WFA-010`, `REQ-WFA-011` | Phase 1 |
| `CAP-WF-005/006` | Multi-level & parallel | `REQ-WFA-012` | Phase 1 |
| `CAP-WF-007` | Delegation | `REQ-WFA-013` | Phase 1 |
| `CAP-WF-008/009` | SLA & escalation | `REQ-WFA-014` | Phase 1 |
| `CAP-WF-010` | Approval history | `REQ-WFA-015` | Phase 1 |
| `CAP-WF-011/012` | My Tasks & dashboard | `REQ-WFA-016`, `REQ-WFA-017` | Phase 1 |

### 15.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-WFA-001` | The service SHALL define workflows (steps, conditions, assignee rules) and initiate instances linked to a business entity. | `CAP-WF-001/002` | P0 | 1 | **Given** a workflow definition, **when** an entity triggers it, **then** an instance is created bound to that entity and `WorkflowInitiated` emitted. | all BCs | — |
| `REQ-WFA-010` | The service SHALL route approval tasks by role, authority level, and branch, and support approve/reject/refer with comments. | `CAP-WF-003/004` | P0 | 1 | **Given** a task, **when** routed, **then** it reaches an eligible approver; a decision with comments transitions the workflow and emits `ApprovalCompleted`. | `IAM` | — |
| `REQ-WFA-012` | The service SHALL support sequential multi-level and parallel (all-must-approve) chains. | `CAP-WF-005/006` | P1 | 1 | **Given** a multi-level/parallel definition, **when** executed, **then** approvals proceed in the defined order/parallelism and complete only when satisfied. | — | — |
| `REQ-WFA-013` | The service SHALL support delegation of approval authority during absence. | `CAP-WF-007` | P1 | 1 | **Given** a delegation, **when** active, **then** tasks route to the delegate and the delegation is audited. | `IAM`, `AUD` | — |
| `REQ-WFA-014` | The service SHALL track SLAs per step and auto-escalate on breach. | `CAP-WF-008/009` | P1 | 1 | **Given** a task past its SLA, **when** breached, **then** it escalates to the next level and `SLABreached`/`WorkflowEscalated` emitted. | `NOT` | — |
| `REQ-WFA-015` | The service SHALL maintain a full approval audit (decisions, timestamps, comments). | `CAP-WF-010` | P0 | 1 | **Given** any decision, **when** made, **then** it is recorded immutably. | `AUD` | — |
| `REQ-WFA-016` | The service SHALL provide a per-user "My Tasks" inbox and a manager dashboard (open workflows, SLA status, bottlenecks). | `CAP-WF-011/012` | P0 | 1 | **Given** a user, **when** they open My Tasks, **then** their pending approvals list; managers see aggregate workflow health. | `RPT` | — |

### 15.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Task routing latency | < 1s |
| SLA breach alert | < 1 min |
| Approval audit completeness | 100% |
| Availability | ≥ 99.9% |

### 15.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `POL`, `CLM`, `BIL`, `UW`, `PRD` | `REQ-WFA-001`, `REQ-WFA-010` | Approval initiation & decision callback |

### 15.5 Open Questions / Risks

- **OQ-WFA-1.** Catalog of approval types & thresholds across modules — confirm with EIC to seed workflow definitions.

---

## 16. Module: Notifications (NOT)

> **Bounded Context:** [`BC-MDH-10`](./Medhen-Platform-Capability-Document.md#bc-mdh-10--notifications-pc-notification-svc) · **Service:** `pc-notification-svc` · **Database:** `pc_notification_db`
> **Capability source:** Cap doc Part III, BC-MDH-10 · **PDD source:** [§12 Module 10](./product_definition_document.md) · **Phase:** 1 (core)

### 16.0 Mission Statement

`NOT` is the event-driven, multi-channel notification hub (SMS, email, in-app, push) with per-event/channel/locale templates, delivery tracking, preferences, and scheduling.

### 16.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-NOT-001..004` | Channels | `REQ-NOT-001`, `REQ-NOT-002` | Phase 1 |
| `CAP-NOT-005/006` | Templates & event triggers | `REQ-NOT-010`, `REQ-NOT-011` | Phase 1 |
| `CAP-NOT-007` | Delivery tracking | `REQ-NOT-012` | Phase 1 |
| `CAP-NOT-008` | Preferences | `REQ-NOT-013` | Phase 1 |
| `CAP-NOT-009/010` | Scheduling & history | `REQ-NOT-014`, `REQ-NOT-015` | Phase 1 |

### 16.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-NOT-001` | The service SHALL send SMS via the Ethiopian gateway and email via SMTP/SendGrid. | `CAP-NOT-001..004` | P0 | 1 | **Given** a notification, **when** dispatched, **then** SMS/email is sent via `INT` and the attempt recorded. | `INT` | — |
| `REQ-NOT-002` | The service SHALL provide in-app notifications and (Phase 4) mobile push. | `CAP-NOT-001..004` | P1 | 1 | **Given** an in-app recipient, **when** an event fires, **then** a real-time in-app notification appears. | — | Push = Phase 4 |
| `REQ-NOT-010` | The service SHALL manage templates per event type, channel, and locale (en/am). | `CAP-NOT-005` | P0 | 1 | **Given** an event + channel + locale, **when** rendering, **then** the correct localized template is used. | `CORE` | — |
| `REQ-NOT-011` | The service SHALL auto-send on domain events per the trigger matrix (quote, policy issued, payment received/due/overdue, endorsement, renewal, claim submitted/updated/settled, approval required, cancellation). | `CAP-NOT-006` | P0 | 1 | **Given** a subscribed domain event, **when** received, **then** the mapped notifications dispatch within 5s. | all BCs | — |
| `REQ-NOT-012` | The service SHALL track delivery status (sent/delivered/failed) via gateway callbacks. | `CAP-NOT-007` | P1 | 1 | **Given** a dispatched notification, **when** a delivery report returns, **then** status updates and `NotificationSent`/`Failed` emitted. | `INT` | — |
| `REQ-NOT-013` | The service SHALL honor per-customer channel preferences (opt-in/opt-out) linked to consent. | `CAP-NOT-008` | P2 | 1 | **Given** a customer opted out of a channel, **when** a non-essential notification is due, **then** it is suppressed on that channel. | `PTY` | — |
| `REQ-NOT-014` | The service SHALL schedule future notifications (payment/renewal reminders) and keep a per-party history. | `CAP-NOT-009/010` | P1 | 1 | **Given** a scheduled notification, **when** the time arrives, **then** it dispatches; all sends are logged per party. | — | — |

### 16.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Dispatch latency | < 5s from triggering event |
| SMS volume | 10,000+/day |
| Delivery tracking | 100% status captured |
| Availability | ≥ 99.5% (Tier-2) |

### 16.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| All BCs | `REQ-NOT-011` | Event-triggered customer/staff messaging |
| `RPT` Reporting | `REQ-NOT-012` | Delivery statistics |

### 16.5 Open Questions / Risks

- **OQ-NOT-1.** SMS aggregator selection (PDD marks TBD) — affects `REQ-NOT-001` integration.
- **OQ-NOT-2.** Which notifications are "essential" (cannot be opted out) vs. marketing — confirm for `REQ-NOT-013`.

---

## 17. Module: Reporting & Analytics (RPT)

> **Bounded Context:** [`BC-MDH-11`](./Medhen-Platform-Capability-Document.md#bc-mdh-11--reporting--analytics-pc-reporting-svc) · **Service:** `pc-reporting-svc` · **Database:** `pc_reporting_db` (CQRS read-side)
> **Capability source:** Cap doc Part III, BC-MDH-11 · **PDD source:** [§13 Module 11](./product_definition_document.md) · **Phase:** 2 (foundations in 1)

### 17.0 Mission Statement

`RPT` is the primary **CQRS read-side** — consuming domain events to build read-optimized projections powering dashboards, operational/financial reports, and NBE regulatory returns; and the future home of the data-warehouse/BI layer.

### 17.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-RPT-001` | Dashboard KPIs | `REQ-RPT-001` | Phase 1–2 |
| `CAP-RPT-002` | Operational reports | `REQ-RPT-010` | Phase 1–2 |
| `CAP-RPT-003` | Financial reports | `REQ-RPT-020` | Phase 2 |
| `CAP-RPT-004` | Regulatory reports (NBE) | `REQ-RPT-030`, `REQ-RPT-031` | Phase 1–2 |
| `CAP-RPT-005` | Report features | `REQ-RPT-040` | Phase 1–2 |
| `CAP-RPT-A1/A4` | Data warehouse / IFRS 17 | `REQ-RPT-050`, `REQ-RPT-051` | Phase 2–3 |
| `CAP-RPT-A3` | Predictive analytics | `REQ-RPT-052` | Phase 4 |

### 17.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-RPT-001` | The service SHALL present dashboard KPIs (GWP, NWP, in-force, claims frequency/severity, loss ratio, combined ratio, collection rate, retention, outstanding reserves) with trends. | `CAP-RPT-001` | P0 | 1 | **Given** platform events, **when** the dashboard loads, **then** KPIs render branch/role-scoped within P95 ≤ 2s from < 5s-lagged projections. | all BCs | — |
| `REQ-RPT-010` | The service SHALL produce operational reports (production, claims, collection, agent performance, renewal pipeline, underwriting, endorsement, cancellation). | `CAP-RPT-002` | P0 | 1 | **Given** a period and filters, **when** run, **then** the report returns accurate figures reconciled to source events. | all BCs | — |
| `REQ-RPT-020` | The service SHALL produce financial reports (premium income, claims paid, aged outstanding premium, commission, reinsurance cession). | `CAP-RPT-003` | P1 | 2 | **Given** financial events, **when** run, **then** the report reconciles to Billing/Commission/Reinsurance ledgers. | `BIL`, `COM`, `RIN` | — |
| `REQ-RPT-030` | The service SHALL generate NBE regulatory returns: quarterly return, annual statutory, motor third-party (Reg. 554/2024), solvency. | `CAP-RPT-004` | P0 | 1 | **Given** a reporting period, **when** a return is generated, **then** it matches the NBE-prescribed format and reconciles to source data. | `CMP` | — |
| `REQ-RPT-040` | The service SHALL support date filters, drill-down, export (PDF/Excel/CSV), scheduled auto-generation & email, and a custom report builder. | `CAP-RPT-005` | P1 | 1 | **Given** a report, **when** exported/scheduled, **then** it delivers in the chosen format/schedule; drill-down navigates to detail. | `DOC`, `NOT` | Builder = P3 |
| `REQ-RPT-050` | The service SHALL provide a data warehouse / self-service BI layer beyond CQRS projections. | `CAP-RPT-A1` | P2 | 3 | **Given** analytical needs, **when** BI is used, **then** business users explore data without impacting transactional systems. | — | Resolves PDD gap |
| `REQ-RPT-051` | The service SHALL support IFRS 17 contract-group measurement (CSM, risk adjustment) and disclosures. | `CAP-RPT-A4` | P1 | 2 | **Given** IFRS 17 revenue events from Billing, **when** measured, **then** contract-group results and disclosures are produced. | `BIL`, `CMP` | Resolves PDD gap |
| `REQ-RPT-052` | The service SHALL provide predictive analytics (churn, loss-ratio forecasting, fraud propensity). | `CAP-RPT-A3` | P3 | 4 | **Given** historical data, **when** models run, **then** explainable predictions surface in dashboards. | `FIN` | Resolves PDD gap (AI) |

### 17.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Projection lag | < 5s from source event |
| Dashboard query | P95 ≤ 2s |
| Read isolation | Reporting load isolated from transactional DBs |
| Statutory return accuracy | 100% reconciled |

### 17.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| Executives / managers | `REQ-RPT-001`, `REQ-RPT-010` | Dashboards & operational MI |
| Compliance / NBE | `REQ-RPT-030`, `REQ-RPT-051` | Regulatory & IFRS 17 |

### 17.5 Open Questions / Risks

- **OQ-RPT-1.** Exact NBE return templates & submission mechanism — confirm to build `REQ-RPT-030`.
- **OQ-RPT-2.** IFRS 17 measurement approach (GMM/PAA) per product — confirm with actuarial/finance for `REQ-RPT-051`.

---

## 18. Module: Reinsurance & Coinsurance (RIN)

> **Bounded Context:** [`BC-MDH-12`](./Medhen-Platform-Capability-Document.md#bc-mdh-12--reinsurance--coinsurance-pc-reinsurance-svc) · **Service:** `pc-reinsurance-svc` · **Phase:** 2

### 18.0 Mission Statement

`RIN` manages risk transfer — treaty & facultative reinsurance plus coinsurance — computing cessions on bind, maintaining the cession register, generating bordereaux, and tracking recoveries and settlements.

### 18.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-RI-001/002` | Treaty configuration & terms | `REQ-RIN-001` | Phase 2 |
| `CAP-RI-003/005` | Automatic cession & register | `REQ-RIN-002`, `REQ-RIN-003` | Phase 2 |
| `CAP-RI-004` | Facultative placement | `REQ-RIN-004` | Phase 2 |
| `CAP-RI-006` | Bordereaux generation | `REQ-RIN-005` | Phase 2 |
| `CAP-RI-007/008` | Recovery & settlement | `REQ-RIN-006` | Phase 2 |
| `CAP-RI-009..012` | Reinsurer mgmt & reporting | `REQ-RIN-007` | Phase 2 |
| `CAP-RI-A1` | Coinsurance management | `REQ-RIN-010` | Phase 3 |

### 18.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-RIN-001` | The service SHALL configure treaties (quota-share, surplus, excess-of-loss) with retention, cession %, layer limits, and reinstatements. | `CAP-RI-001/002` | P1 | 2 | **Given** treaty terms, **when** configured, **then** they are versioned and applied to matching risks. | — | — |
| `REQ-RIN-002` | The service SHALL compute and record automatic cessions on policy bind/endorsement per applicable treaties. | `CAP-RI-003` | P1 | 2 | **Given** `PolicyBound`, **when** consumed, **then** the cession is computed and `CessionRecorded` emitted. | `POL` | — |
| `REQ-RIN-003` | The service SHALL maintain a cession register of all ceded risks. | `CAP-RI-005` | P1 | 2 | **Given** cessions, **when** queried, **then** the register lists all ceded risks with amounts. | — | — |
| `REQ-RIN-004` | The service SHALL record facultative placements for individual large risks. | `CAP-RI-004` | P2 | 3 | **Given** a large risk, **when** placed facultatively, **then** the placement and reinsurer are recorded. | `INT` | — |
| `REQ-RIN-005` | The service SHALL generate periodic bordereaux for reinsurers. | `CAP-RI-006` | P2 | 2 | **Given** a period, **when** bordereaux are generated, **then** they reconcile 100% to policy/claim data. | `DOC` | — |
| `REQ-RIN-006` | The service SHALL track reinsurance recoveries on claims and premium settlements. | `CAP-RI-007/008` | P1 | 2 | **Given** `ClaimSettled`, **when** consumed, **then** the recoverable is computed and `RecoveryRecorded` emitted. | `CLM`, `BIL` | — |
| `REQ-RIN-007` | The service SHALL manage reinsurer records (EthioRe, international) and treaty renewal tracking. | `CAP-RI-009..012` | P2 | 2 | **Given** reinsurers/treaties, **when** managed, **then** records and renewal timelines are tracked. | — | — |
| `REQ-RIN-010` | The service SHALL record coinsurance shares (lead/follow, share %) and split settlements on large risks. | `CAP-RI-A1` | P2 | 3 | **Given** a coinsured risk, **when** a claim settles, **then** each co-insurer's share is computed. | `POL`, `CLM` | Resolves PDD gap |

### 18.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Cession computation | On bind, < 5s (warm path) |
| Bordereaux accuracy | 100% reconciled |
| Availability | ≥ 99.5% |

### 18.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `RPT` Reporting | `REQ-RIN-002`, `REQ-RIN-006` | NWP & cession reporting |
| `BIL` Billing | `REQ-RIN-006` | Recovery & settlement accounting |

### 18.5 Open Questions / Risks

- **OQ-RIN-1.** EIC's actual treaty structure is TBD (PDD note) — detailed config deferred until EIC provides terms.

---

## 19. Module: Complaints & Disputes (CPL) · *new*

> **Bounded Context:** [`BC-MDH-13`](./Medhen-Platform-Capability-Document.md#bc-mdh-13--complaints--disputes-pc-complaints-svc--new) · **Service:** `pc-complaints-svc` · **Phase:** 1
> *New module — addresses PDD gap; delivers UK-aligned complaints handling (FCA DISP).*

### 19.0 Mission Statement

`CPL` delivers mandatory conduct-grade complaints handling: capture, classify, route, enforce acknowledgement/resolution SLAs, manage redress, escalate to Ombudsman/NBE, and produce complaints MI for Consumer-Duty evidence.

### 19.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-CPL-001` | Complaint intake | `REQ-CPL-001` | Phase 1 |
| `CAP-CPL-002` | Classification & routing | `REQ-CPL-002` | Phase 1 |
| `CAP-CPL-003` | SLA management | `REQ-CPL-003` | Phase 1 |
| `CAP-CPL-004` | Investigation & redress | `REQ-CPL-004` | Phase 1 |
| `CAP-CPL-005` | Escalation | `REQ-CPL-005` | Phase 1 |
| `CAP-CPL-006` | Complaints MI | `REQ-CPL-006` | Phase 1 |

### 19.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-CPL-001` | The service SHALL log complaints from any channel (portal, phone, branch, email) linked to party/policy/claim. | `CAP-CPL-001` | P0 | 1 | **Given** a complaint, **when** logged, **then** it is recorded with complainant, linked entity, and channel; `ComplaintLogged` emitted. | `PTY`, `POL`, `CLM` | — |
| `REQ-CPL-002` | The service SHALL classify complaints (category, severity, root-cause taxonomy) and route to an owner. | `CAP-CPL-002` | P0 | 1 | **Given** a complaint, **when** classified, **then** it routes to the responsible owner via workflow. | `WFA` | — |
| `REQ-CPL-003` | The service SHALL enforce acknowledgement and final-response SLAs (DISP timescales) with breach alerting. | `CAP-CPL-003` | P0 | 1 | **Given** a complaint, **when** SLA clocks run, **then** acknowledgement/final-response deadlines are tracked and breaches alerted. | `NOT` | DISP timescales |
| `REQ-CPL-004` | The service SHALL record investigation, decision (uphold/reject), and redress (compensation/apology), issuing redress payment where due. | `CAP-CPL-004` | P0 | 1 | **Given** an investigated complaint, **when** decided, **then** the outcome and any redress are recorded; redress payment routes to Billing; `ComplaintResolved` emitted. | `BIL` | — |
| `REQ-CPL-005` | The service SHALL track escalation to Ombudsman/NBE and regulator correspondence. | `CAP-CPL-005` | P1 | 1 | **Given** an unresolved/escalated complaint, **when** referred, **then** the external referral and correspondence are tracked; `ComplaintEscalated` emitted. | `INT` | — |
| `REQ-CPL-006` | The service SHALL produce complaints MI (root-cause trends, Consumer-Duty/fair-value evidence, regulatory complaints return). | `CAP-CPL-006` | P1 | 1 | **Given** complaint data, **when** MI is run, **then** trends and regulatory returns are produced and feed product governance. | `RPT`, `PRD` | — |

### 19.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Acknowledgement / final-response SLA | Per DISP; breach alerting |
| MI completeness | 100% complaints captured & categorized |
| Availability | ≥ 99.5% |

### 19.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `RPT` Reporting | `REQ-CPL-006` | Complaints MI & regulatory return |
| `PRD` Product | `REQ-CPL-006` | Root-cause feedback to product governance |

### 19.5 Open Questions / Risks

- **OQ-CPL-1.** Confirm applicable complaints timescales & external body (NBE consumer protection / Ombudsman equivalent) for Ethiopia vs. UK DISP.

---

## 20. Module: Financial Crime — Fraud / AML / Sanctions (FIN) · *new*

> **Bounded Context:** [`BC-MDH-14`](./Medhen-Platform-Capability-Document.md#bc-mdh-14--financial-crime--fraud--aml--sanctions-pc-fincrime-svc--new) · **Service:** `pc-fincrime-svc` · **Phase:** 2 (screening hooks in 1)
> *New module — addresses PDD gap (single "fraud flag"); delivers screening + AML + fraud SIU.*

### 20.0 Mission Statement

`FIN` provides sanctions/PEP screening, AML/CFT monitoring, and insurance fraud detection (SIU case management), producing regulator-defensible controls for NBE/FIS and FATF/ESAAMLG, reusing shared audit/workflow rather than duplicating them.

### 20.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-FIN-001` | Sanctions / PEP screening | `REQ-FIN-001` | Phase 1–2 |
| `CAP-FIN-002` | AML/CFT monitoring | `REQ-FIN-002` | Phase 2 |
| `CAP-FIN-003` | Fraud detection & indicators | `REQ-FIN-003` | Phase 2 |
| `CAP-FIN-004` | SIU case management | `REQ-FIN-004` | Phase 2 |
| `CAP-FIN-005` | Suspicious-transaction reporting | `REQ-FIN-005` | Phase 2 |
| `CAP-FIN-006` | Watchlist / list management | `REQ-FIN-006` | Phase 2 |
| `CAP-FIN-A1` | ML fraud scoring | `REQ-FIN-010` | Phase 4 |

### 20.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-FIN-001` | The service SHALL screen parties against sanctions/PEP lists at onboarding and material change (Amharic-aware matching) and disposition hits. | `CAP-FIN-001` | P1 | 1 | **Given** a party, **when** screened, **then** a result is recorded; a positive hit creates a case and gates binding until dispositioned; `ScreeningResult` emitted. | `PTY`, `INT` | Hook in Phase 1 |
| `REQ-FIN-002` | The service SHALL monitor premium/claim flows for suspicious patterns and risk-rate customers. | `CAP-FIN-002` | P1 | 2 | **Given** transaction flows, **when** monitored, **then** suspicious patterns raise alerts with a risk score. | `BIL`, `CLM` | — |
| `REQ-FIN-003` | The service SHALL apply configurable fraud indicators and score claims, routing high scores to SIU. | `CAP-FIN-003` | P1 | 2 | **Given** a claim, **when** scored, **then** a fraud score + indicators are returned; high scores create an SIU case; `ClaimScore` emitted. | `CLM` | — |
| `REQ-FIN-004` | The service SHALL manage SIU investigation cases (evidence, workflow, outcome). | `CAP-FIN-004` | P1 | 2 | **Given** a flagged claim/party, **when** an SIU case is opened, **then** it follows an investigation workflow with evidence and outcome recorded. | `WFA`, `DOC` | — |
| `REQ-FIN-005` | The service SHALL draft and file suspicious-transaction reports (STR/SAR) to FIS with audited decisions. | `CAP-FIN-005` | P1 | 2 | **Given** a reportable case, **when** an STR is filed, **then** it transmits via `INT` to FIS and the decision is audited immutably. | `INT`, `AUD` | — |
| `REQ-FIN-006` | The service SHALL ingest and version sanctions/PEP watchlists. | `CAP-FIN-006` | P1 | 2 | **Given** a list update, **when** ingested, **then** the versioned list is available to screening and re-screening runs against it. | `INT` | — |
| `REQ-FIN-010` | The service SHALL provide explainable ML fraud scoring and (long-horizon) graph/network fraud detection. | `CAP-FIN-A1` | P3 | 4 | **Given** historical fraud data, **when** the model scores a claim, **then** an explainable propensity is produced for analyst review. | `RPT` | Resolves PDD gap (AI) |

### 20.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Screening latency (inline) | P95 ≤ 2s |
| Screening recall | ≥ 99% on labelled set |
| Disposition audit | 100% immutable |
| Availability | ≥ 99.9% (Tier-0; gates binding) |

### 20.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `PTY` Party | `REQ-FIN-001` | Screening result & bind gate |
| `CLM` Claims | `REQ-FIN-003` | Fraud score |
| Compliance / FIS | `REQ-FIN-005` | STR filing |

### 20.5 Open Questions / Risks

- **OQ-FIN-1.** Sanctions/PEP list provider(s) & FIS filing interface — confirm for `INT` connectors.
- **OQ-FIN-2.** Binding-gate policy on screening-unavailable (block vs. post-bind screen) — confirm risk appetite with Compliance.

---

## 21. Module: Commission Management (COM) · *new*

> **Bounded Context:** [`BC-MDH-15`](./Medhen-Platform-Capability-Document.md#bc-mdh-15--commission-management-pc-commission-svc--new) · **Service:** `pc-commission-svc` · **Phase:** 1
> *New module — formalizes PDD commission mentions into a first-class capability.*

### 21.0 Mission Statement

`COM` owns commission scheme configuration, accurate calculation on production, clawbacks on cancellation, statements, and payout — a tier-1 distribution-management necessity.

### 21.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-COM-001` | Commission scheme configuration | `REQ-COM-001` | Phase 1 |
| `CAP-COM-002` | Commission calculation | `REQ-COM-002` | Phase 1 |
| `CAP-COM-003` | Clawback & adjustment | `REQ-COM-003` | Phase 1 |
| `CAP-COM-004` | Commission statements | `REQ-COM-004` | Phase 1 |
| `CAP-COM-005` | Payout & ERP sync | `REQ-COM-005` | Phase 1 |

### 21.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-COM-001` | The service SHALL configure commission schemes per product/LOB/producer (rates, tiers, overrides, caps), effective-dated. | `CAP-COM-001` | P1 | 1 | **Given** a scheme, **when** configured, **then** it is versioned and applied to matching production. | `PRD` | — |
| `REQ-COM-002` | The service SHALL calculate commission on bind/renewal/endorsement (base vs. override, tiered volume). | `CAP-COM-002` | P1 | 1 | **Given** `PolicyBound`/`Renewed`/`Endorsed`, **when** consumed, **then** commission is computed per the effective scheme; `CommissionEarned` emitted. | `POL` | — |
| `REQ-COM-003` | The service SHALL reverse/adjust commission on cancellation, refund, or non-payment (clawback). | `CAP-COM-003` | P1 | 1 | **Given** `PolicyCancelled`/refund, **when** consumed, **then** the corresponding commission is clawed back; `CommissionClawedBack` emitted. | `POL`, `BIL` | — |
| `REQ-COM-004` | The service SHALL generate per-producer commission statements and payout runs. | `CAP-COM-004` | P1 | 1 | **Given** a period, **when** statements run, **then** each producer's earned/clawed-back/net commission is stated. | `DOC` | — |
| `REQ-COM-005` | The service SHALL approve and sync commission payouts to the ERP. | `CAP-COM-005` | P1 | 1 | **Given** an approved payout, **when** synced, **then** a commission journal reaches the ERP; `CommissionPaid` emitted. | `BIL`, `INT`, `WFA` | — |

### 21.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Calculation accuracy | 100% reconciled to production |
| Statement generation | Scheduled per period |
| Availability | ≥ 99.5% |

### 21.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `BIL`/ERP | `REQ-COM-005` | Commission payout journals |
| `RPT` Reporting | `REQ-COM-002`, `REQ-COM-003` | Commission & producer reporting |

### 21.5 Open Questions / Risks

- **OQ-COM-1.** Commission schemes (rates, tiers, overrides) per product/producer — confirm with EIC distribution.

---

## 22. Module: Identity & Access Management (IAM)

> **Bounded Context:** [`BC-MDH-16`](./Medhen-Platform-Capability-Document.md#bc-mdh-16--identity--access-management-keycloak--pc-iam-svc) · **Service:** Keycloak + `pc-iam-svc` · **Phase:** 1

### 22.0 Mission Statement

`IAM` is the single identity & access authority — OIDC authentication, RBAC + ABAC authorization on every endpoint/event, user lifecycle for staff and external users, branch/product-scoped access, and login/security auditing.

### 22.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-IAM-001/002` | Registration & authentication | `REQ-IAM-001`, `REQ-IAM-002` | Phase 1 |
| `CAP-IAM-003/004` | RBAC & ABAC | `REQ-IAM-003`, `REQ-IAM-004` | Phase 1 |
| `CAP-IAM-005/006` | User & role management | `REQ-IAM-005` | Phase 1 |
| `CAP-IAM-007` | Branch restriction | `REQ-IAM-006` | Phase 1 |
| `CAP-IAM-008/009` | Session & password policy | `REQ-IAM-007` | Phase 1 |
| `CAP-IAM-010` | Login history audit | `REQ-IAM-008` | Phase 1 |

### 22.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-IAM-001` | The service SHALL register internal staff and external customers/agents/brokers, linking each to a party. | `CAP-IAM-001` | P0 | 1 | **Given** a new user, **when** registered, **then** an account is created and linked to a party; `UserLinked` emitted. | `PTY` | — |
| `REQ-IAM-002` | The service SHALL authenticate via username/password with optional MFA and SSO (OIDC/Keycloak). | `CAP-IAM-002` | P0 | 1 | **Given** credentials, **when** authenticated, **then** an OIDC token is issued; MFA enforced where configured. | Keycloak | — |
| `REQ-IAM-003` | The service SHALL enforce RBAC on every API endpoint and UI element. | `CAP-IAM-003` | P0 | 1 | **Given** a request, **when** authorized, **then** access is granted only if the user's role permits the action; denials are audited. | all BCs | — |
| `REQ-IAM-004` | The service SHALL enforce fine-grained ABAC permissions scoped by module, action, branch, and product line. | `CAP-IAM-004` | P0 | 1 | **Given** a request, **when** evaluated, **then** attribute conditions (branch/product/authority) are enforced, not just role. | all BCs | Resolves PDD RBAC→ABAC gap |
| `REQ-IAM-005` | The service SHALL provide a user-management admin portal and custom role/permission-set creation. | `CAP-IAM-005/006` | P0 | 1 | **Given** an admin, **when** managing users/roles, **then** changes take effect and are audited. | `AUD` | — |
| `REQ-IAM-006` | The service SHALL restrict access by branch. | `CAP-IAM-007` | P1 | 1 | **Given** a branch-scoped user, **when** querying, **then** only their branch's data is visible unless explicitly cross-branch authorized. | — | — |
| `REQ-IAM-007` | The service SHALL enforce configurable session timeouts, concurrent-session limits, and password policy (complexity/expiry/history). | `CAP-IAM-008/009` | P1 | 1 | **Given** policy config, **when** applied, **then** sessions and passwords conform; violations are blocked. | — | — |
| `REQ-IAM-008` | The service SHALL audit login history (success/failure with IP, device, timestamp). | `CAP-IAM-010` | P1 | 1 | **Given** a login attempt, **when** it occurs, **then** it is recorded with context; `LoginRecorded` emitted. | `AUD` | — |

### 22.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Token validation | P95 ≤ 20ms (cached JWKS) |
| Authorization decision | P95 ≤ 30ms |
| Availability | ≥ 99.95% (Tier-0; gates all access) |

### 22.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| All BCs | `REQ-IAM-003`, `REQ-IAM-004` | AuthN/AuthZ on every request/event |
| `WFA` Workflow | `REQ-IAM-004` | Authority for routing |

### 22.5 Open Questions / Risks

- **OQ-IAM-1.** Fayda-backed customer login (`CAP-IAM-A2`) — Phase 3 dependency on Fayda OIDC availability.

---

## 23. Module: Audit & Compliance (AUD)

> **Bounded Context:** [`BC-MDH-17`](./Medhen-Platform-Capability-Document.md#bc-mdh-17--audit--compliance-pc-audit-svc) · **Service:** `pc-audit-svc` · **Phase:** 1

### 23.0 Mission Statement

`AUD` is the single immutable audit authority — an append-only, hash-chained record of every data change and privileged action, searchable and exportable for examination, with retention enforced (≥ 7 years financial).

### 23.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-AUD-001/002` | Data-change & action audit | `REQ-AUD-001`, `REQ-AUD-002` | Phase 1 |
| `CAP-AUD-003` | Immutable append-only log | `REQ-AUD-003` | Phase 1 |
| `CAP-AUD-004/005` | Search & export | `REQ-AUD-004` | Phase 1 |
| `CAP-AUD-A1` | Legal hold | `REQ-AUD-010` | Phase 2 |

### 23.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-AUD-001` | The service SHALL record every data change (who, what, when, old→new). | `CAP-AUD-001` | P0 | 1 | **Given** any mutation across BCs, **when** committed, **then** an audit record captures actor, entity, and before/after values. | all BCs | — |
| `REQ-AUD-002` | The service SHALL record every privileged action (login, approve, decline, issue, settle). | `CAP-AUD-002` | P0 | 1 | **Given** a privileged action, **when** performed, **then** it is recorded with actor and context. | all BCs | — |
| `REQ-AUD-003` | The service SHALL store audit records in an immutable, append-only, hash-chained log. | `CAP-AUD-003` | P0 | 1 | **Given** the audit log, **when** tampering is attempted, **then** hash-chain verification detects it; records cannot be altered/deleted. | — | — |
| `REQ-AUD-004` | The service SHALL provide audit search/filter and compliance export. | `CAP-AUD-004/005` | P1 | 1 | **Given** search criteria, **when** queried by Compliance, **then** matching records return and can be exported for examination. | `RPT` | — |
| `REQ-AUD-010` | The service SHALL support legal hold suspending deletion of records under investigation. | `CAP-AUD-A1` | P1 | 2 | **Given** a legal hold, **when** applied, **then** affected records are exempt from retention deletion until released. | `CMP` | Resolves PDD gap |

### 23.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Audit write durability | 100% — no dropped records |
| Immutability | Tamper-evident hash chain |
| Retention | ≥ 7 years financial |
| Availability | ≥ 99.95% (Tier-0) |

### 23.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| All BCs | `REQ-AUD-001`, `REQ-AUD-002` | Audit emission |
| Compliance / NBE | `REQ-AUD-004` | Examination search & export |

### 23.5 Open Questions / Risks

- **OQ-AUD-1.** Retention classes vs. Ethiopian DPP erasure — coordinate with `REQ-PTY-060`/`CMP` on the lawful-basis matrix.

---

## 24. Module: Integration & Anti-Corruption Layer (INT) · *new*

> **Bounded Context:** [`BC-MDH-18`](./Medhen-Platform-Capability-Document.md#bc-mdh-18--integration--anti-corruption-layer-pc-integration-svc--new) · **Service:** `pc-integration-svc` · **Phase:** 1
> *New module — consolidates PDD's scattered integrations into one ACL.*

### 24.0 Mission Statement

`INT` is the single anti-corruption layer for all external integrations — payment gateways, ERP, Fayda, SMS, reinsurers, government systems — isolating the core from external quirks and enforcing idempotency/retry/circuit-breaking.

### 24.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-INT-001/002` | Payment connectors & callbacks | `REQ-INT-001`, `REQ-INT-002` | Phase 1 |
| `CAP-INT-003` | ERP integration | `REQ-INT-003` | Phase 1 |
| `CAP-INT-004` | Fayda national ID | `REQ-INT-004` | Phase 1 |
| `CAP-INT-005` | SMS gateway | `REQ-INT-005` | Phase 1 |
| `CAP-INT-006` | Reinsurer/broker exchange | `REQ-INT-006` | Phase 2 |
| `CAP-INT-007` | Resilience primitives | `REQ-INT-007` | Phase 1 |

### 24.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-INT-001` | The service SHALL provide payment-gateway connectors (Telebirr, CBE Birr, Amole) supporting initiate, status, refund. | `CAP-INT-001` | P0 | 1 | **Given** a payment operation, **when** invoked, **then** the connector translates to/from the gateway protocol and returns canonical results. | `BIL` | — |
| `REQ-INT-002` | The service SHALL handle gateway webhooks/callbacks idempotently with signature verification. | `CAP-INT-002` | P0 | 1 | **Given** a callback (incl. duplicate), **when** received, **then** signature is verified and the event processed exactly once. | `CORE` | — |
| `REQ-INT-003` | The service SHALL integrate with the EIC ERP via events + REST sync for journals. | `CAP-INT-003` | P0 | 1 | **Given** a journal event, **when** synced, **then** it reaches the ERP; failures buffer and replay; daily reconciliation confirms. | `BIL`, `COM` | — |
| `REQ-INT-004` | The service SHALL integrate Fayda national ID for identity verification. | `CAP-INT-004` | P1 | 1 | **Given** a verification request, **when** Fayda is available, **then** identity is verified; on outage, a degraded/manual path is signalled. | `PTY` | — |
| `REQ-INT-005` | The service SHALL integrate the SMS gateway (send + delivery-report callback). | `CAP-INT-005` | P0 | 1 | **Given** an SMS send, **when** dispatched, **then** it transmits and delivery reports update status. | `NOT` | — |
| `REQ-INT-006` | The service SHALL exchange bordereaux & settlements with reinsurers/brokers. | `CAP-INT-006` | P2 | 2 | **Given** a bordereaux/settlement, **when** exchanged, **then** it transmits in the agreed format and is acknowledged. | `RIN` | — |
| `REQ-INT-007` | The service SHALL provide resilience primitives (retry, circuit breaker, idempotency, DLQ, sandbox routing) for all connectors. | `CAP-INT-007` | P0 | 1 | **Given** an external failure, **when** it occurs, **then** the circuit breaks without cascading into the core; retries/DLQ apply. | `OBS` | Resolves PDD gap |

### 24.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Callback idempotency | 100% |
| External-failure isolation | No cascade into core |
| Reconciliation | Daily with each gateway/ERP |
| Availability | ≥ 99.9% (Tier-0) |

### 24.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `BIL` Billing | `REQ-INT-001`, `REQ-INT-002`, `REQ-INT-003` | Payments & ERP |
| `PTY` / `NOT` / `RIN` / `FIN` | `REQ-INT-004`, `REQ-INT-005`, `REQ-INT-006` | Fayda, SMS, reinsurer, list/STR |

### 24.5 Open Questions / Risks

- **OQ-INT-1.** Confirm each external system's API contract, sandbox availability, and SLAs (Telebirr, CBE Birr, Amole, ERP, Fayda, SMS aggregator).

---

## 25. Module: Observability & Telemetry (OBS) · *new*

> **Bounded Context:** [`BC-MDH-19`](./Medhen-Platform-Capability-Document.md#bc-mdh-19--observability--telemetry-pc-observability) · **Service:** `pc-observability` · **Phase:** 1
> *New module — addresses PDD gap (no observability).*

### 25.0 Mission Statement

`OBS` is the observability spine — structured logging, metrics, and distributed tracing (OpenTelemetry) — with golden-signal dashboards, SLO/burn-rate alerting, and event-flow/DLQ monitoring.

### 25.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-OBS-001` | Structured logging | `REQ-OBS-001` | Phase 1 |
| `CAP-OBS-002` | Metrics & golden signals | `REQ-OBS-002` | Phase 1 |
| `CAP-OBS-003` | Distributed tracing | `REQ-OBS-003` | Phase 1 |
| `CAP-OBS-004` | SLO & alerting | `REQ-OBS-004` | Phase 1 |
| `CAP-OBS-005` | Event-flow & DLQ monitoring | `REQ-OBS-005` | Phase 1 |

### 25.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-OBS-001` | Every service SHALL emit correlation-id-linked, PII-masked structured logs. | `CAP-OBS-001` | P0 | 1 | **Given** any request, **when** processed, **then** logs carry a correlation id and no unmasked PII. | all BCs | — |
| `REQ-OBS-002` | Every service SHALL emit golden-signal metrics (latency, throughput, errors, saturation). | `CAP-OBS-002` | P0 | 1 | **Given** a running service, **when** scraped, **then** golden-signal series exist with standard labels. | all BCs | — |
| `REQ-OBS-003` | The platform SHALL provide end-to-end distributed tracing (OpenTelemetry). | `CAP-OBS-003` | P0 | 1 | **Given** a cross-service flow, **when** traced, **then** a single trace spans all involved services. | all BCs | — |
| `REQ-OBS-004` | The platform SHALL define SLOs with burn-rate alerts and on-call routing. | `CAP-OBS-004` | P0 | 1 | **Given** an SLO burn, **when** the threshold is crossed, **then** an alert fires and routes to on-call within 1 min. | — | — |
| `REQ-OBS-005` | The platform SHALL monitor Kafka consumer lag, DLQ depth, and saga health. | `CAP-OBS-005` | P1 | 1 | **Given** rising lag or DLQ growth, **when** it exceeds thresholds, **then** an alert fires before user impact. | — | — |

### 25.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Telemetry coverage | 100% of services |
| Alert latency | < 1 min on SLO burn |
| Availability | ≥ 99.9% |

### 25.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| SRE / Ops | all `REQ-OBS-*` | Dashboards, alerts, traces |

### 25.5 Open Questions / Risks

- **OQ-OBS-1.** Observability stack selection (e.g., OTel + Prometheus/Grafana/Tempo/Loki) — confirm with platform engineering.

---

## 26. Module: Shared-Core Kernel & Product Extensibility (CORE) · *new*

> **Bounded Context:** [`BC-MDH-20`](./Medhen-Platform-Capability-Document.md#bc-mdh-20--shared-core-kernel--product-extensibility-pc-platform-kernel) · **Service:** `pc-platform-kernel` (library/SDK + control-plane) · **Phase:** 1
> *New module — operationalizes the shared-core promise (design principle "Shared Core, Product Extensions" & PRD §5).*

### 26.0 Mission Statement

`CORE` is the kernel & SDK that lets one platform core serve many product lines through configuration and plugins, not code forks. It owns the product-extension contract, the dynamic risk-schema engine, the rating/rules execution engines, and shared primitives (idempotency, i18n, Ethiopian calendar, money/tax).

### 26.1 Capability Scope

| CAP-ID | Capability | REQ-IDs | Status |
|:---|:---|:---|:---|
| `CAP-CORE-001` | Product-extension contract | `REQ-CORE-001` | Phase 1 |
| `CAP-CORE-002` | Dynamic risk-schema engine | `REQ-CORE-002` | Phase 1 |
| `CAP-CORE-003` | Rating execution engine | `REQ-CORE-003` | Phase 1 |
| `CAP-CORE-004` | Rules execution engine | `REQ-CORE-004` | Phase 1 |
| `CAP-CORE-005` | Idempotency SDK | `REQ-CORE-005` | Phase 1 |
| `CAP-CORE-006` | Internationalization primitives | `REQ-CORE-006` | Phase 1 |
| `CAP-CORE-007` | Money & tax primitives | `REQ-CORE-007` | Phase 1 |
| `CAP-CORE-A1` | Low-code extension toolkit | `REQ-CORE-010` | Phase 3 |

### 26.2 Functional Requirements

| REQ ID | Requirement | Capabilities | Priority | Phase | Acceptance Criteria | Dependencies | Notes |
|:---|:---|:---|:---|:---|:---|:---|:---|
| `REQ-CORE-001` | The kernel SHALL define a stable product-extension contract (risk schema, rating adapter, UW-rule bindings, claim taxonomy, document merge-fields) that a new LOB implements without core code change. | `CAP-CORE-001` | P0 | 1 | **Given** a new LOB implementing the contract, **when** registered & activated, **then** quote/bind/claims operate for it with zero modification to core aggregates. | `PRD` | Core of shared-core goal |
| `REQ-CORE-002` | The kernel SHALL provide a JSON-Schema-driven risk-schema engine to define, validate, version, and render LOB insured-item data. | `CAP-CORE-002` | P0 | 1 | **Given** an LOB risk schema, **when** risk data is entered, **then** it is validated against the schema and rendered dynamically in the quote wizard. | `POL` | — |
| `REQ-CORE-003` | The kernel SHALL provide a deterministic, versioned rating execution engine (rate tables/factors/loadings/discounts). | `CAP-CORE-003` | P0 | 1 | **Given** rate config + risk data, **when** executed, **then** the result is deterministic and reproducible per version. | `RAT` | — |
| `REQ-CORE-004` | The kernel SHALL provide a versioned rules execution engine for underwriting/eligibility decisions. | `CAP-CORE-004` | P0 | 1 | **Given** rule config, **when** evaluated, **then** accept/refer/decline is produced deterministically per version. | `UW` | — |
| `REQ-CORE-005` | The kernel SHALL provide a shared idempotency-key store and helpers. | `CAP-CORE-005` | P0 | 1 | **Given** a keyed operation, **when** retried, **then** it executes at-most-once effect. | `BIL`, `INT` | — |
| `REQ-CORE-006` | The kernel SHALL provide i18n primitives: en/am bilingual, Noto Sans Ethiopic, Ethiopian (Ge'ez) calendar with Gregorian toggle, ETB/locale formatting. | `CAP-CORE-006` | P0 | 1 | **Given** a locale, **when** rendering dates/numbers/text, **then** Amharic/Ge'ez and ETB format correctly; calendar conversions are accurate. | `DOC`, all UIs | Consolidates PDD i18n |
| `REQ-CORE-007` | The kernel SHALL provide money & tax primitives (ETB money type, VAT/stamp-duty, rounding, pro-rata-temporis). | `CAP-CORE-007` | P0 | 1 | **Given** monetary computations, **when** performed, **then** rounding, VAT/stamp-duty, and pro-rata are consistent platform-wide. | `RAT`, `BIL` | — |
| `REQ-CORE-010` | The kernel SHALL provide a low-code toolkit to author risk schemas, rules, and rate tables for new LOBs. | `CAP-CORE-A1` | P2 | 3 | **Given** a product author, **when** they use the toolkit, **then** a new LOB extension can be defined and tested without engineering. | — | Resolves PDD gap |

### 26.3 Module-Level NFRs

| Dimension | Target |
|:---|:---|
| Schema resolution | P95 ≤ 20ms (cached) |
| Backward compatibility | Additive-only; existing LOBs never break |
| New-LOB onboarding | Variant = config-only; new LOB = extension, zero core change |
| Determinism | Rating/rules reproducible per version |

### 26.4 Consumer Coupling

| Consumer | REQs consumed | Coupling Surface |
|:---|:---|:---|
| `PRD`/`POL`/`RAT`/`UW`/`CLM`/`DOC` | `REQ-CORE-001`, `REQ-CORE-002` | Extension contract & schema engine |
| `BIL`/`INT` | `REQ-CORE-005`, `REQ-CORE-007` | Idempotency, money/tax |
| All UIs/`DOC` | `REQ-CORE-006` | i18n & calendar |

### 26.5 Open Questions / Risks

- **OQ-CORE-1.** Rules/rating engine build-vs-adopt (embedded DSL vs. external decision engine) — architectural decision (future ADR).
- **OQ-CORE-2.** Ethiopian-calendar authority for contract dates (see OQ-POL-2) — affects `REQ-CORE-006` semantics.

---

## 27. Cross-Cutting Non-Functional Requirements (NFR)

| REQ ID | Requirement | Priority | Phase | Acceptance Criteria |
|:---|:---|:---|:---|:---|
| `REQ-NFR-001` | Standard-query API latency SHALL be P95 < 500ms; complex queries P95 < 2s. | P0 | 1 | Load test at target concurrency confirms P95 targets. |
| `REQ-NFR-002` | Premium calculation SHALL complete in < 1s; document generation in < 5s. | P0 | 1 | Measured under load. |
| `REQ-NFR-003` | Domain-event propagation (warm path) SHALL be < 5s end-to-end. | P0 | 1 | Event-lag telemetry (`OBS`) within target. |
| `REQ-NFR-004` | The platform SHALL support 500+ concurrent internal users. | P0 | 1 | Load test sustains target with SLOs met. |
| `REQ-NFR-005` | Tier-0 BCs SHALL target ≥ 99.9% availability (IAM/Audit ≥ 99.95%); overall business-hours uptime ≥ 99.9%. | P0 | 1 | Availability measured over rolling 30 days. |
| `REQ-NFR-006` | RTO SHALL be ≤ 4h and RPO ≤ 1h. | P0 | 1 | DR drill achieves targets. |
| `REQ-NFR-007` | Each microservice SHALL scale horizontally and be stateless (state in DB/Redis). | P0 | 1 | Scaling test shows linear throughput gain. |
| `REQ-NFR-008` | All services SHALL emit OpenTelemetry telemetry per `REQ-OBS-*`. | P0 | 1 | 100% telemetry coverage verified. |

> NFR targets are modest baselines aligned to the PDD; §22 of the Cap doc benchmarks tier-1 aspirations. Availability was raised from the PDD's 99.5% to a 99.9% target for Tier-0 to reflect the "tier-1" ambition — confirm with EIC as an SLA commitment.

## 28. Security Requirements (SEC)

| REQ ID | Requirement | Priority | Phase | Acceptance Criteria |
|:---|:---|:---|:---|:---|
| `REQ-SEC-001` | Authentication SHALL use OAuth2/OIDC via Keycloak. | P0 | 1 | All access flows through OIDC; no bypass. |
| `REQ-SEC-002` | Authorization SHALL be enforced via RBAC + ABAC at gateway and service level. | P0 | 1 | Endpoint tests confirm deny-by-default. |
| `REQ-SEC-003` | Data in transit SHALL use TLS 1.3 externally and mTLS internally. | P0 | 1 | TLS scan + mesh policy verified. |
| `REQ-SEC-004` | PII and financial data SHALL be encrypted at rest (AES-256), with field-level encryption for national ID & bank accounts. | P0 | 1 | DB inspection confirms encryption; keys in KMS/HSM. |
| `REQ-SEC-005` | PII SHALL be masked in logs and access-controlled. | P0 | 1 | Log inspection shows no unmasked PII. |
| `REQ-SEC-006` | Secrets SHALL be managed centrally (e.g., Vault); no secrets in code/config. | P0 | 1 | Secret scan passes; secrets sourced from vault. |
| `REQ-SEC-007` | APIs SHALL enforce rate limiting, input validation, and injection prevention. | P0 | 1 | Security test suite passes. |
| `REQ-SEC-008` | Security events SHALL be immutably logged (`AUD`). | P0 | 1 | Security events present & tamper-evident. |

## 29. Data Requirements (DAT)

| REQ ID | Requirement | Priority | Phase | Acceptance Criteria |
|:---|:---|:---|:---|:---|
| `REQ-DAT-001` | Each service SHALL own its database (no cross-service DB coupling). | P0 | 1 | No shared schemas across services. |
| `REQ-DAT-002` | Cross-service consistency SHALL be eventual via the Outbox pattern; all event consumers idempotent. | P0 | 1 | Crash-recovery test shows no lost/duplicated effect. |
| `REQ-DAT-003` | Event schemas SHALL be registered with backward-compatibility enforced in CI. | P0 | 1 | CI rejects non-backward-compatible schema change. |
| `REQ-DAT-004` | Policy data SHALL be bi-temporal (business-time + system-time) with no data loss on amendment. | P0 | 1 | Any-point-in-time reconstruction verified. |
| `REQ-DAT-005` | Backups SHALL be daily full + hourly incremental with ≥ 30-day retention and tested restores. | P0 | 1 | Restore drill succeeds within RTO/RPO. |
| `REQ-DAT-006` | Reporting SHALL read from replicas/projections isolated from transactional stores. | P1 | 2 | Reporting load does not affect transactional SLOs. |
| `REQ-DAT-007` | Reference data (admin units, banks, ISO codes) SHALL be centrally managed. | P1 | 1 | Reference datasets versioned & validated. |
| `REQ-DAT-008` | A data-migration & legacy-coexistence capability SHALL support phased onboarding with reconciliation & rollback. | P0 | 1 | Migration wave reconciles 100% before cutover. |

## 30. Compliance & Regulatory Requirements (CMP)

| REQ ID | Requirement | Priority | Phase | Driver | Acceptance Criteria |
|:---|:---|:---|:---|:---|:---|
| `REQ-CMP-001` | The platform SHALL comply with NBE insurance regulations and produce required statutory returns. | P0 | 1 | NBE | Returns generated per `REQ-RPT-030` and accepted. |
| `REQ-CMP-002` | Motor policies SHALL enforce compulsory third-party cover and issue COI + sticker. | P0 | 1 | NBE Reg. 554/2024 | Compliant documents issued (`REQ-DOC-002`). |
| `REQ-CMP-003` | The platform SHALL comply with the Ethiopian Personal Data Protection Proclamation (consent, DSR, residency). | P0 | 1 | Ethiopian DPP | Consent/DSR (`REQ-PTY-060`) and residency (`REQ-DAT`) enforced. |
| `REQ-CMP-004` | Financial records SHALL be retained ≥ 7 years with immutable audit for examination. | P0 | 1 | Ethiopian law | Retention & hash-chained audit (`REQ-AUD-003`) verified. |
| `REQ-CMP-005` | The platform SHALL support UK-aligned conduct: Consumer Duty (fair value, vulnerable customers), ICOBS/IDD demands-and-needs, Insurance Act 2015 fair-presentation, DISP complaints. | P0 | 1–2 | UK conduct | Traces to `REQ-PRD-051`, `REQ-PTY-061`, `REQ-UW-030`, `REQ-CPL-*`. |
| `REQ-CMP-006` | The platform SHALL support AML/CFT obligations: screening, monitoring, STR filing. | P0 | 1–2 | AML law; FATF | Traces to `REQ-FIN-001/002/005`. |
| `REQ-CMP-007` | The platform SHALL be IFRS 17/9 ready for financial reporting. | P1 | 2 | IFRS | Traces to `REQ-BIL-070`, `REQ-RPT-051`. |
| `REQ-CMP-008` | Tax computation SHALL apply VAT (15%) and stamp duty per Ethiopian law. | P0 | 1 | Ethiopian tax | Traces to `REQ-RAT-006`. |

## 31. Acceptance-Criteria Framework

- Every **functional** `REQ` states criteria in **Given-When-Then** so QA derives contract/integration tests directly.
- Every **NFR** states a **measurable target** (latency, throughput, availability) verified by load/DR testing.
- **Definition of Done** per REQ: implemented; unit + integration + contract tests pass; acceptance criteria demonstrated; telemetry & audit emitted; traceability (REQ↔CAP↔BC) recorded; docs updated.
- **Traceability is validated** as coverage matrices (Appendices) — no orphan CAPs, no untraceable REQs, every regulatory driver mapped.

## 32. Risks & Mitigations

| ID | Risk | Impact | Mitigation |
|:---|:---|:---|:---|
| R-1 | External payment/ERP/Fayda APIs unstable or undocumented | Delays; failed transactions | ACL isolation (`INT`), sandbox-first, circuit breakers, idempotency, reconciliation |
| R-2 | Legacy data migration complexity | Cutover risk, data quality | Phased strangler-fig, reconciliation gates, rollback per wave (`REQ-DAT-008`) |
| R-3 | EIC product content (rates/rules/wordings) not ready | Blocks product launch | Config-driven engine; EIC content workstream in parallel; shadow testing |
| R-4 | UK-conduct vs. NBE regulatory divergence | Compliance ambiguity | Compliance SME sign-off; `CMP` maps both; configurable timescales |
| R-5 | Scope creep across many LOBs | Timeline risk | Phased delivery (Motor first); shared-core proven before LOB expansion |
| R-6 | Amharic/Ge'ez calendar edge cases | Incorrect dates/premiums | Kernel primitives with dedicated test suite (`REQ-CORE-006`) |
| R-7 | Under-specified reinsurance structure | Rework | Defer detailed treaty config until EIC provides terms (OQ-RIN-1) |

## 33. Dependencies & Constraints

**External dependencies:** Telebirr / CBE Birr / Amole APIs; EIC ERP; Fayda; SMS aggregator; sanctions/PEP list & FIS filing (Phase 2); reinsurer/broker interfaces (Phase 2).
**Business content dependencies (EIC):** rate tables, underwriting rules, product wordings, authority limits, commission schemes, treaty terms, NBE return templates.
**Constraints:** Ethiopian data residency; bilingual (en/am); compulsory motor TP; 7-year retention; ETB + VAT/stamp duty; Ethiopian calendar support.

## 34. Phased-Delivery View

| Phase | Modules / REQ scope | Exit criteria |
|:---|:---|:---|
| **Phase 0 — Pilot MVP** | Thin Motor slice across PTY, PRD (seed), RAT, UW (STP), POL (quote→bind→issue), BIL (Telebirr sandbox), DOC (schedule+COI+sticker), CLM (FNOL→fast-track), NOT, IAM (basic), AUD (basic), CORE (i18n/money/schema) — each a multi-repo microservice on the production stack. | Full buy→claim demo runs on synthetic data, bilingual, Telebirr sandbox, deployed to a demo env; shared-core story demonstrable |
| **Phase 1 — Core + Motor (Production)** | Full core (PTY, PRD, POL, RAT, UW, CLM, BIL, DOC, WFA, NOT, CPL, COM, IAM, AUD, INT, OBS, CORE); Motor LOB; FIN screening hook; migration | Motor quote→claim→settlement live in production; NBE motor reporting; audit/IAM/observability operational; ≥ 80% motor STP |
| **Phase 2 — Life** | Life extension; RPT depth; FIN full; RIN; IFRS 17 events; conduct enhancements (PRD-051, PTY-061, UW-030, POL-090) | Life live; NBE returns automated; screening & monitoring live |
| **Phase 3 — Commercial** | Property, Marine, Liability, Engineering, WC via extension contract; coinsurance; data warehouse | Commercial LOBs live with zero core change |
| **Phase 4 — Specialty + AI** | Bonds, A&H, Travel; AI/ML (IDP, damage estimation, fraud ML, predictive analytics); agentic assist | AI capabilities in production with guardrails |

### 34.1 Phase 0 — Pilot MVP (Detail)

**Objective.** Deliver a live, clickable, end-to-end **Motor** demonstration — on synthetic data with sandbox/mocked integrations — that proves the three winning themes (end-to-end automation, Ethiopian localization, one-core-many-products) to EIC decision-makers within ~10–12 weeks, so the platform can be shown while contract terms are discussed.

**Demo storyboard (the pitch narrative).**

1. Agent/customer builds a **Motor quote** in Amharic → itemized premium (base + factors + VAT 15% + stamp duty).
2. **Auto-underwriting (STP)** approves a standard risk instantly.
3. Pay via **Telebirr (sandbox)** → receipt.
4. **Policy issued** → download **bilingual policy schedule + Certificate of Insurance + QR windshield sticker**.
5. **File a claim from mobile** (photo + GPS) → **fast-track settlement** → SMS confirmation.
6. Open the **product-config + Motor risk-schema screen** → *"the same core runs Life/Property; only this configuration changes"* (shared-core proof).
7. Show the **immutable audit trail + a KPI tile** (GWP, policies in-force) (governance/tier-1 proof).

**In-scope REQ mapping (happy path only).**

| Area | Pilot slice | REQ-IDs |
|:---|:---|:---|
| Party | Register individual + basic KYC upload (no full verification) | `REQ-PTY-001`, `-020` |
| Product | One seeded Motor product (config, not full lifecycle UI) | `REQ-PRD-001/010/020` (seed) |
| Rating | Motor premium + VAT/stamp duty | `REQ-RAT-001/002/003/006` |
| Underwriting | Auto-accept STP for standard risk | `REQ-UW-001/002` |
| Policy | Quote wizard → bind → issue | `REQ-POL-001/002/003/020/021/030` |
| Billing | Telebirr sandbox payment + receipt (single premium) | `REQ-BIL-001/010/020/023` |
| Documents | Schedule + COI + QR sticker, bilingual | `REQ-DOC-001/002/010/020` |
| Claims | Mobile FNOL → fast-track settlement | `REQ-CLM-001/002/011/040/050` |
| Notifications | SMS/email on bind + settle | `REQ-NOT-001/011` |
| Core/i18n | am/en, ETB money, Ethiopian calendar, Motor risk-schema | `REQ-CORE-002/006/007` |
| IAM + Audit | Basic auth + roles; immutable trail | `REQ-IAM-002/003`, `REQ-AUD-001/003` |

**Explicitly OUT (deferred to Phase 1+).** Endorsements/renewals/cancellations, installments, reconciliation, ERP sync, Fayda (mocked), full KYC workflow, UW referral workbench/authority matrix, reserves/recovery/total-loss, reinsurance, commission, complaints, fin-crime (mocked screening), NBE returns, multi-branch, ABAC depth.

**Architecture approach.** The pilot is built on the **production architecture** — **multi-repo microservices** with DDD, hexagonal/clean architecture, EDA (Kafka), CQRS, the outbox pattern, and saga orchestration — not a monolith. Phase 0 simply **scopes the service set** to those the Motor demo exercises (party, product, policy, rating, underwriting, claims, billing, document, notification, gateway + Keycloak), each in its own repo. A shared Go library and a service template make standing these up in the window feasible; **zero throwaway** — the same services are hardened and extended in Phase 1. See [Cap doc §5.1–5.4](./Medhen-Platform-Capability-Document.md#5-architecture-style--ddd--hexagonal--eda--cqrs).

**Milestones (~10–12 weeks; assumes ~4–6 engineers).**

| Milestone | ~Weeks | Outcome |
|:---|:---|:---|
| **M0 Foundations** | 1–2 | Per-service repos scaffolded (hexagonal template) + CI, shared-go library, basic IAM, i18n/ETB/calendar primitives, seeded Motor product + rate table + risk schema; walking skeleton (party → empty quote) |
| **M1 Buy journey** | 3–6 | Rating, STP UW, Telebirr sandbox, bind, bilingual schedule + COI + sticker, notifications — purchase demo works end-to-end |
| **M2 Claim + story** | 7–10 | Mobile FNOL → fast-track settle, audit trail, KPI tile, shared-core config teaser, seeded demo data + script, deployed demo environment |
| **M3 Polish** | 11–12 | Hardening, demo rehearsal, buffer |

**Success criteria.** The full buy→claim journey runs on synthetic data, bilingually, with Telebirr sandbox, on a reachable demo environment; the shared-core and governance stories are demonstrable; the [Cap doc](./Medhen-Platform-Capability-Document.md) + this PRD serve as credibility artifacts.

**EIC / input dependencies.** Representative Motor rate table & factors (or synthetic); Telebirr sandbox credentials (or payment mock); NBE sticker/COI layout (or approximation); a demo hosting target.

## 35. Appendices

### Appendix A — REQ → CAP Forward Matrix (summary)

Each module's §x.1 capability-scope table is the authoritative per-module forward map (CAP → REQ-IDs). Consolidated: every `REQ-{DOMAIN}-*` lists its `CAP-*` in the `Capabilities` column of its requirement row.

### Appendix B — CAP → REQ Reverse Coverage

| CAP prefix | BC | Covered by REQ domain | Coverage |
|:---|:---|:---|:---|
| `CAP-PARTY-*` | BC-MDH-01 | `REQ-PTY-*` | Full |
| `CAP-PROD-*` | BC-MDH-02 | `REQ-PRD-*` | Full |
| `CAP-POL-*` | BC-MDH-03 | `REQ-POL-*` | Full |
| `CAP-RATE-*` | BC-MDH-04 | `REQ-RAT-*` | Full |
| `CAP-UW-*` | BC-MDH-05 | `REQ-UW-*` | Full |
| `CAP-CLM-*` | BC-MDH-06 | `REQ-CLM-*` | Full |
| `CAP-BIL-*` | BC-MDH-07 | `REQ-BIL-*` | Full |
| `CAP-DOC-*` | BC-MDH-08 | `REQ-DOC-*` | Full |
| `CAP-WF-*` | BC-MDH-09 | `REQ-WFA-*` | Full |
| `CAP-NOT-*` | BC-MDH-10 | `REQ-NOT-*` | Full |
| `CAP-RPT-*` | BC-MDH-11 | `REQ-RPT-*` | Full |
| `CAP-RI-*` | BC-MDH-12 | `REQ-RIN-*` | Full |
| `CAP-CPL-*` | BC-MDH-13 | `REQ-CPL-*` | Full |
| `CAP-FIN-*` | BC-MDH-14 | `REQ-FIN-*` | Full |
| `CAP-COM-*` | BC-MDH-15 | `REQ-COM-*` | Full |
| `CAP-IAM-*` | BC-MDH-16 | `REQ-IAM-*` | Full |
| `CAP-AUD-*` | BC-MDH-17 | `REQ-AUD-*` | Full |
| `CAP-INT-*` | BC-MDH-18 | `REQ-INT-*` | Full |
| `CAP-OBS-*` | BC-MDH-19 | `REQ-OBS-*` | Full |
| `CAP-CORE-*` | BC-MDH-20 | `REQ-CORE-*` | Full |

*(No orphan capabilities. Advanced `-A#` capabilities map to enhancement REQs at P2/P3 or later phases.)*

### Appendix C — Regulatory → REQ Reverse Matrix

| Driver | REQ-IDs |
|:---|:---|
| NBE Reg. 554/2024 (compulsory motor TP) | `REQ-POL-030`, `REQ-DOC-002`, `REQ-RPT-030`, `REQ-CMP-002` |
| Ethiopian Personal Data Protection Proclamation | `REQ-PTY-060`, `REQ-SEC-004/005`, `REQ-DAT-*`, `REQ-CMP-003` |
| Financial records / 7-year retention | `REQ-AUD-003`, `REQ-AUD-010`, `REQ-CMP-004` |
| AML/CFT; FATF/ESAAMLG | `REQ-PTY-062`, `REQ-FIN-001/002/005`, `REQ-CMP-006` |
| FCA Consumer Duty / ICOBS / IDD / Insurance Act 2015 / DISP | `REQ-PRD-051`, `REQ-PTY-061`, `REQ-UW-030`, `REQ-POL-090`, `REQ-CPL-*`, `REQ-CMP-005` |
| IFRS 17 / IFRS 9 | `REQ-BIL-070`, `REQ-RPT-051`, `REQ-CMP-007` |
| Ethiopian tax (VAT/stamp duty) | `REQ-RAT-006`, `REQ-CMP-008` |

### Appendix D — NFR / SLO Catalog

Consolidated in §27 and each module's §x.3; mirrors Cap doc Appendix H.

### Appendix E — Glossary

Inherits the [Cap doc Appendix J](./Medhen-Platform-Capability-Document.md#appendix-j--glossary) and PDD §23.

---

*End of Medhen Platform PRD — Version 1.0. Authoritative capability source: [Medhen Platform Capability Document](./Medhen-Platform-Capability-Document.md).*
