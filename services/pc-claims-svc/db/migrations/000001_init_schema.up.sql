-- 000001_init_schema.up.sql

CREATE TABLE claims (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    claim_number VARCHAR(50) NOT NULL UNIQUE,
    policy_id UUID NOT NULL,
    status VARCHAR(30) NOT NULL,
    version INT NOT NULL,
    date_of_loss TIMESTAMPTZ NOT NULL,
    loss_type VARCHAR(50) NOT NULL,
    fraud_score INT,
    is_vulnerable_customer BOOLEAN DEFAULT FALSE,
    sla_deadline TIMESTAMPTZ,
    loss_details JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_claims_policy ON claims(policy_id);
CREATE INDEX idx_claims_tenant_status ON claims(tenant_id, status);

CREATE TABLE reserve_ledger (
    id BIGSERIAL PRIMARY KEY,
    claim_id UUID REFERENCES claims(id) ON DELETE RESTRICT,
    reserve_type VARCHAR(20) NOT NULL, 
    transaction_type VARCHAR(20) NOT NULL, 
    amount NUMERIC(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    exchange_rate NUMERIC(10, 6) NOT NULL,
    base_amount_delta NUMERIC(19, 4) NOT NULL,
    running_balance_base NUMERIC(19, 4) NOT NULL,
    reason_code VARCHAR(50) NOT NULL,
    author_id UUID NOT NULL,
    sys_valid_from TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sys_valid_to TIMESTAMPTZ NOT NULL DEFAULT '9999-12-31 23:59:59Z',
    biz_valid_from TIMESTAMPTZ NOT NULL,
    biz_valid_to TIMESTAMPTZ NOT NULL DEFAULT '9999-12-31 23:59:59Z'
);

ALTER TABLE reserve_ledger ADD CONSTRAINT chk_positive_balance CHECK (running_balance_base >= 0);

CREATE TABLE settlements (
    id UUID PRIMARY KEY,
    claim_id UUID REFERENCES claims(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL,
    gross_loss_base NUMERIC(19, 4) NOT NULL,
    policy_deductible NUMERIC(19, 4) NOT NULL,
    salvage_value_base NUMERIC(19, 4) NOT NULL,
    net_settlement_base NUMERIC(19, 4) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    partition_key VARCHAR(100) NOT NULL,
    payload BYTEA NOT NULL,
    headers JSONB,
    processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_unprocessed ON outbox(processed) WHERE processed = FALSE;
