CREATE TABLE approval_history (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    instance_id UUID REFERENCES workflow_instances(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    action VARCHAR(30) NOT NULL,
    actor_id VARCHAR(100) NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_instance ON approval_history(instance_id);
CREATE INDEX idx_audit_task ON approval_history(task_id);
