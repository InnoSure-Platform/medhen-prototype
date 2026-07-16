CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE products (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    code VARCHAR(50) NOT NULL,
    lob VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL,
    version INT NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    require_fair_value BOOLEAN DEFAULT false,
    schema_payload JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, code, version)
);

CREATE TABLE product_coverages (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_mandatory BOOLEAN DEFAULT false,
    min_limit NUMERIC(15, 2),
    max_limit NUMERIC(15, 2),
    deductible_config JSONB DEFAULT '{}',
    parent_coverage_code VARCHAR(50),
    UNIQUE (product_id, code)
);

CREATE TABLE rate_tables (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    version INT NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    dimensions JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name, version)
);

CREATE TABLE rate_rows (
    id BIGSERIAL PRIMARY KEY,
    rate_table_id UUID REFERENCES rate_tables(id) ON DELETE CASCADE,
    dimension_bounds JSONB NOT NULL,
    factor NUMERIC(10, 4) NOT NULL
);

CREATE INDEX idx_rate_rows_table ON rate_rows(rate_table_id);

CREATE TABLE uw_rule_sets (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    version INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, product_id, version)
);

CREATE TABLE uw_rules (
    id UUID PRIMARY KEY,
    rule_set_id UUID REFERENCES uw_rule_sets(id) ON DELETE CASCADE,
    priority INT NOT NULL,
    condition JSONB NOT NULL,
    action VARCHAR(50) NOT NULL,
    message TEXT NOT NULL
);

CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unpublished ON outbox(created_at) WHERE published_at IS NULL;
