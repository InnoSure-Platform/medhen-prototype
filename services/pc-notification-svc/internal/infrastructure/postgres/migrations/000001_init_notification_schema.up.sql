CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY,
    code VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    category VARCHAR(50) NOT NULL,
    subject_template TEXT,
    body_template TEXT NOT NULL,
    version INT NOT NULL,
    UNIQUE(code, channel, locale, version)
);

CREATE TABLE IF NOT EXISTS routing_preferences (
    party_id UUID PRIMARY KEY,
    opted_out_sms BOOLEAN DEFAULT false,
    opted_out_email BOOLEAN DEFAULT false,
    opted_out_in_app BOOLEAN DEFAULT false,
    marketing_opt_in BOOLEAN DEFAULT false,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    party_id UUID NOT NULL,
    template_code VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    category VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL,
    recipient_address VARCHAR(255) NOT NULL,
    rendered_content TEXT,
    vendor_receipt_id VARCHAR(255),
    error_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_party ON notifications(party_id);
CREATE INDEX idx_notifications_status ON notifications(status);

CREATE TABLE IF NOT EXISTS inbox (
    event_id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    partition_key VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMPTZ
);
