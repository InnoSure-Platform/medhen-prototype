CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    partition_key VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE TABLE parties (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    dob DATE,
    gender VARCHAR(10),
    national_id_type VARCHAR(50),
    national_id_number VARCHAR(100),
    
    legal_name VARCHAR(200),
    registration_number VARCHAR(100),
    industry_code VARCHAR(50),
    
    tin VARCHAR(100),
    surviving_party_id UUID REFERENCES parties(id),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique index for non-merged parties
CREATE UNIQUE INDEX idx_parties_national_id ON parties(tenant_id, national_id_number) WHERE status != 'MERGED';

CREATE TABLE addresses (
    id UUID PRIMARY KEY,
    party_id UUID REFERENCES parties(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    region VARCHAR(100) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    woreda VARCHAR(100) NOT NULL,
    kebele VARCHAR(100),
    house_number VARCHAR(50),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE party_roles (
    id UUID PRIMARY KEY,
    party_id UUID REFERENCES parties(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    attributes JSONB,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    UNIQUE (party_id, role)
);
