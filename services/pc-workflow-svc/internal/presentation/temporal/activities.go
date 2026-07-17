package temporal

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
	
	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/medhen/pc-workflow-svc/internal/domain/workflow"
)

// Activities encapsulates dependencies for Temporal activities.
type Activities struct {
	Repo      workflow.Repository
	IAMClient workflow.IAMClient
}

// AssignTaskParams are the inputs for assigning a task.
type AssignTaskParams struct {
	InstanceID      string
	StepNodeID      string
	RoleExpression  string
	ContextSnapshot []byte
}

// AssignTaskActivity resolves the assignee via IAM, applies Load-Balanced Routing, and persists the Task in DB.
func (a *Activities) AssignTaskActivity(ctx context.Context, params AssignTaskParams) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing AssignTaskActivity", "instanceID", params.InstanceID)

	// 1. Resolve eligible users via IAM
	users, err := a.IAMClient.ResolveRoleToUsers(ctx, params.RoleExpression, params.ContextSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to resolve IAM role: %w", err)
	}
	
	if len(users) == 0 {
		return "", fmt.Errorf("no eligible users found for role: %s", params.RoleExpression)
	}
	
	// 2. Intelligent Routing: Sort by Workload
	workloads, err := a.Repo.GetWorkloads(ctx, users)
	if err != nil {
		return "", fmt.Errorf("failed to get workloads: %w", err)
	}
	
	sort.SliceStable(users, func(i, j int) bool {
		return workloads[users[i]] < workloads[users[j]]
	})
	
	assigneeID := users[0] // Select least-loaded user
	
	// 3. Check for active delegations
	delegation, err := a.Repo.GetActiveDelegation(ctx, assigneeID)
	if err != nil {
		return "", fmt.Errorf("failed to check delegations: %w", err)
	}
	
	var delegatedFrom string
	if delegation != nil {
		delegatedFrom = assigneeID
		assigneeID = delegation.DelegateID
	}

	// 4. Create Task in DB
	taskID := "tsk-" + uuid.NewString()
	task := &workflow.Task{
		ID:              taskID,
		TenantID:        "t-system",
		InstanceID:      params.InstanceID,
		StepNodeID:      params.StepNodeID,
		AssigneeID:      assigneeID,
		DelegatedFromID: delegatedFrom,
		Status:          workflow.TaskPending,
		CreatedAt:       time.Now().UTC(),
	}

	if err := a.Repo.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to persist task: %w", err)
	}

	// 5. Append Immutable Audit Log
	action := workflow.ActionAssigned
	if delegatedFrom != "" {
		action = workflow.ActionReassigned
	}
	auditLog := &workflow.ApprovalHistory{
		ID:         "aud-" + uuid.NewString(),
		TenantID   : "t-system",
		InstanceID: params.InstanceID,
		TaskID:     taskID,
		Action:     action,
		ActorID:    "system-assigner",
		Timestamp:  time.Now().UTC(),
	}
	_ = a.Repo.AppendAuditHistory(ctx, auditLog)

	// 6. Publish Event to Outbox
	eventPayload, _ := json.Marshal(map[string]string{
		"event_type": "TaskAssignedEvent",
		"task_id":    taskID,
		"assignee":   assigneeID,
	})
	if err := a.Repo.PublishEvent(ctx, "platform.workflow.task.v1", taskID, eventPayload); err != nil {
		return "", fmt.Errorf("failed to publish outbox event: %w", err)
	}

	return taskID, nil
}

// EscalateTaskActivity handles SLA breaches by reassigning the task and auditing it.
func (a *Activities) EscalateTaskActivity(ctx context.Context, taskID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing EscalateTaskActivity", "taskID", taskID)

	task, err := a.Repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != workflow.TaskPending {
		return nil
	}

	managerID, err := a.IAMClient.GetManager(ctx, task.AssigneeID)
	if err != nil {
		return err
	}

	task.Status = workflow.TaskEscalated
	task.AssigneeID = managerID
	
	if err := a.Repo.UpdateTask(ctx, task); err != nil {
		return err
	}

	// Append Audit Log
	auditLog := &workflow.ApprovalHistory{
		ID:         "aud-" + uuid.NewString(),
		TenantID:   task.TenantID,
		InstanceID: task.InstanceID,
		TaskID:     taskID,
		Action:     workflow.ActionEscalated,
		ActorID:    "system-sla-monitor",
		Timestamp:  time.Now().UTC(),
	}
	_ = a.Repo.AppendAuditHistory(ctx, auditLog)

	eventPayload, _ := json.Marshal(map[string]string{
		"event_type": "TaskEscalatedEvent",
		"task_id":    taskID,
		"new_assignee": managerID,
	})
	return a.Repo.PublishEvent(ctx, "platform.workflow.task.v1", taskID, eventPayload)
}

// NotifyWorkflowCompletedActivity publishes the final outcome to the downstream BC.
func (a *Activities) NotifyWorkflowCompletedActivity(ctx context.Context, instanceID string, outcome string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing NotifyWorkflowCompletedActivity", "instanceID", instanceID)
	
	status := workflow.StatusApproved
	if outcome == "REJECT" {
		status = workflow.StatusRejected
	}
	
	if err := a.Repo.UpdateInstanceStatus(ctx, instanceID, status); err != nil {
		return err
	}

	eventPayload, _ := json.Marshal(map[string]string{
		"event_type": "WorkflowCompletedEvent",
		"instance_id": instanceID,
		"outcome": outcome,
	})
	return a.Repo.PublishEvent(ctx, "platform.workflow.instance.v1", instanceID, eventPayload)
}
