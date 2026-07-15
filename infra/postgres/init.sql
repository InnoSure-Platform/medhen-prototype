-- Medhen Phase 0 — logical databases / schemas per BC (ADR-PC-020)
-- Single Postgres instance; schemas enforce isolation boundaries.

CREATE SCHEMA IF NOT EXISTS pc_party;
CREATE SCHEMA IF NOT EXISTS pc_product;
CREATE SCHEMA IF NOT EXISTS pc_policy;
CREATE SCHEMA IF NOT EXISTS pc_billing;
CREATE SCHEMA IF NOT EXISTS pc_document;
CREATE SCHEMA IF NOT EXISTS pc_claims;
CREATE SCHEMA IF NOT EXISTS pc_audit;
CREATE SCHEMA IF NOT EXISTS pc_notification;
CREATE SCHEMA IF NOT EXISTS pc_outbox;

GRANT ALL ON SCHEMA pc_party TO medhen;
GRANT ALL ON SCHEMA pc_product TO medhen;
GRANT ALL ON SCHEMA pc_policy TO medhen;
GRANT ALL ON SCHEMA pc_billing TO medhen;
GRANT ALL ON SCHEMA pc_document TO medhen;
GRANT ALL ON SCHEMA pc_claims TO medhen;
GRANT ALL ON SCHEMA pc_audit TO medhen;
GRANT ALL ON SCHEMA pc_notification TO medhen;
GRANT ALL ON SCHEMA pc_outbox TO medhen;
