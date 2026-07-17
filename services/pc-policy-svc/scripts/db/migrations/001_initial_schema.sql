-- Create extension for range types and GiST indexes (Bi-temporal)
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Quotes Table
CREATE TABLE quotes (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    product_id UUID NOT NULL,
    party_id UUID NOT NULL,
    status VARCHAR(30) NOT NULL,
    risk_payload JSONB NOT NULL,
    premium NUMERIC(15, 2) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Policies Table
CREATE TABLE policies (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    policy_number VARCHAR(100) UNIQUE NOT NULL,
    product_id UUID NOT NULL,
    party_id UUID NOT NULL,
    status VARCHAR(30) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Policy Versions (Bi-temporal)
CREATE TABLE policy_versions (
    id UUID PRIMARY KEY,
    policy_id UUID REFERENCES policies(id) ON DELETE CASCADE,
    version_seq INT NOT NULL,
    risk_payload JSONB NOT NULL,
    total_premium NUMERIC(15, 2) NOT NULL,
    
    -- Bi-temporal axes
    effective_period tstzrange NOT NULL,
    system_period tstzrange NOT NULL DEFAULT tstzrange(CURRENT_TIMESTAMP, 'infinity'),
    
    -- Prevent overlapping effective dates for currently known system truth
    EXCLUDE USING gist (
        policy_id WITH =,
        effective_period WITH &&,
        system_period WITH &&
    )
);

-- Transactional Outbox
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ
);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY,
    actor_id VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
