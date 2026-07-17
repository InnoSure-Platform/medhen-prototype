package workflow

import (
	"context"
)

// Repository defines the persistence interface for the workflow context.
type Repository interface {
	// Instance management
	CreateInstance(ctx context.Context, instance *WorkflowInstance) error
	UpdateInstanceStatus(ctx context.Context, instanceID string, status WorkflowStatus) error
	GetInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error)

	// Task management
	CreateTask(ctx context.Context, task *Task) error
	UpdateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, taskID string) (*Task, error)
	GetPendingTasksForAssignee(ctx context.Context, assigneeID string) ([]*Task, error)
	GetTasksForManager(ctx context.Context, managerID string) ([]*Task, error)
	GetWorkloads(ctx context.Context, userIDs []string) (map[string]int, error)

	// Definitions
	GetDefinition(ctx context.Context, code string) (*WorkflowDefinition, error)

	// Delegations
	GetActiveDelegation(ctx context.Context, delegatorID string) (*DelegationRule, error)
	
	// Audit History
	AppendAuditHistory(ctx context.Context, record *ApprovalHistory) error
	
	// Outbox
	PublishEvent(ctx context.Context, topic string, partitionKey string, payload []byte) error
}

// IAMClient defines the external dependency for resolving roles to concrete users.
type IAMClient interface {
	// ResolveRoleToUsers queries pc-iam-svc to return a list of eligible User IDs
	// for a given role expression and context payload.
	ResolveRoleToUsers(ctx context.Context, roleExpression string, contextPayload []byte) ([]string, error)
	
	// GetManager returns the manager of a given user for escalation.
	GetManager(ctx context.Context, userID string) (string, error)
}
