package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	
	"github.com/medhen/pc-workflow-svc/internal/domain/workflow"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateInstance(ctx context.Context, instance *workflow.WorkflowInstance) error {
	ctxSnapshot, err := json.Marshal(instance.ContextSnapshot)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO workflow_instances (id, tenant_id, definition_id, business_entity_id, initiator_id, status, temporal_run_id, context_snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err = r.db.ExecContext(ctx, query,
		instance.ID, instance.TenantID, instance.DefinitionID, instance.BusinessEntityID,
		instance.InitiatorID, instance.Status, instance.TemporalRunID, ctxSnapshot, instance.CreatedAt)
	return err
}

func (r *Repository) UpdateInstanceStatus(ctx context.Context, instanceID string, status workflow.WorkflowStatus) error {
	query := `UPDATE workflow_instances SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, instanceID)
	return err
}

func (r *Repository) GetInstance(ctx context.Context, instanceID string) (*workflow.WorkflowInstance, error) {
	// Stub implementation
	return &workflow.WorkflowInstance{ID: instanceID, Status: workflow.StatusRunning}, nil
}

func (r *Repository) CreateTask(ctx context.Context, task *workflow.Task) error {
	query := `
		INSERT INTO tasks (id, tenant_id, instance_id, step_node_id, assignee_id, delegated_from_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.TenantID, task.InstanceID, task.StepNodeID,
		task.AssigneeID, task.DelegatedFromID, task.Status, task.CreatedAt)
	return err
}

func (r *Repository) UpdateTask(ctx context.Context, task *workflow.Task) error {
	query := `
		UPDATE tasks 
		SET status = $1, decision_outcome = $2, decision_comment = $3, decision_by = $4, resolved_at = $5
		WHERE id = $6`
		
	_, err := r.db.ExecContext(ctx, query,
		task.Status, task.DecisionOutcome, task.DecisionComment, task.DecisionBy, task.ResolvedAt, task.ID)
	return err
}

func (r *Repository) GetTask(ctx context.Context, taskID string) (*workflow.Task, error) {
	// Stub implementation
	return &workflow.Task{ID: taskID, Status: workflow.TaskPending}, nil
}

func (r *Repository) GetPendingTasksForAssignee(ctx context.Context, assigneeID string) ([]*workflow.Task, error) {
	// Stub implementation
	return []*workflow.Task{}, nil
}

func (r *Repository) GetDefinition(ctx context.Context, code string) (*workflow.WorkflowDefinition, error) {
	// Stub implementation
	return &workflow.WorkflowDefinition{ID: "def-123", Code: code}, nil
}

func (r *Repository) GetActiveDelegation(ctx context.Context, delegatorID string) (*workflow.DelegationRule, error) {
	// Stub implementation
	return nil, nil // No delegation
}

func (r *Repository) PublishEvent(ctx context.Context, topic string, partitionKey string, payload []byte) error {
	query := `
		INSERT INTO outbox (topic, partition_key, payload, headers)
		VALUES ($1, $2, $3, '{}')`
		
	_, err := r.db.ExecContext(ctx, query, topic, partitionKey, payload)
	return err
}

func (r *Repository) AppendAuditHistory(ctx context.Context, record *workflow.ApprovalHistory) error {
	query := `
		INSERT INTO approval_history (id, tenant_id, instance_id, task_id, action, actor_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
		
	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.TenantID, record.InstanceID, record.TaskID,
		record.Action, record.ActorID, record.Timestamp)
	return err
}

func (r *Repository) GetTasksForManager(ctx context.Context, managerID string) ([]*workflow.Task, error) {
	// In reality, this would join with IAM or filter by branch
	// For now, return a stub list
	return []*workflow.Task{}, nil
}

func (r *Repository) GetWorkloads(ctx context.Context, userIDs []string) (map[string]int, error) {
	// Stub implementation returning 0 for all users to prevent bottleneck assignment
	workloads := make(map[string]int)
	for _, id := range userIDs {
		workloads[id] = 0
	}
	return workloads, nil
}

