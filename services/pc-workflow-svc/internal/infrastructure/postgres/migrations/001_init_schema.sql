CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    code VARCHAR(100) NOT NULL,
    version INT NOT NULL,
    graph_payload JSONB NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, code, version)
);

CREATE TABLE workflow_instances (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    definition_id UUID REFERENCES workflow_definitions(id),
    business_entity_id VARCHAR(100) NOT NULL,
    initiator_id VARCHAR(100) NOT NULL,
    status VARCHAR(30) NOT NULL,
    temporal_run_id VARCHAR(100) NOT NULL,
    context_snapshot JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wfi_entity ON workflow_instances(business_entity_id);

CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    instance_id UUID REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_node_id VARCHAR(100) NOT NULL,
    assignee_id VARCHAR(100),
    delegated_from_id VARCHAR(100),
    status VARCHAR(30) NOT NULL,
    decision_outcome VARCHAR(30),
    decision_comment TEXT,
    decision_by VARCHAR(100),
    sla_breach_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id) WHERE status = 'PENDING';

CREATE TABLE delegations (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    delegator_id VARCHAR(100) NOT NULL,
    delegate_id VARCHAR(100) NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN DEFAULT false
);

CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    partition_key TEXT NOT NULL,
    payload BYTEA NOT NULL,
    headers JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);
