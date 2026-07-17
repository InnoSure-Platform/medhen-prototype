package workflow

import (
	"errors"
	"time"
)

// WorkflowStatus defines the high-level state of an instance.
type WorkflowStatus string

const (
	StatusRunning    WorkflowStatus = "RUNNING"
	StatusApproved   WorkflowStatus = "APPROVED"
	StatusRejected   WorkflowStatus = "REJECTED"
	StatusTerminated WorkflowStatus = "TERMINATED"
)

// TaskStatus defines the state of a single approval task.
type TaskStatus string

const (
	TaskPending   TaskStatus = "PENDING"
	TaskCompleted TaskStatus = "COMPLETED"
	TaskCancelled TaskStatus = "CANCELLED"
	TaskEscalated TaskStatus = "ESCALATED"
)

// DecisionOutcome represents the user's decision.
type DecisionOutcome string

const (
	DecisionApprove DecisionOutcome = "APPROVE"
	DecisionReject  DecisionOutcome = "REJECT"
	DecisionRefer   DecisionOutcome = "REFER"
)

var (
	ErrInvalidTransition = errors.New("invalid workflow state transition")
	ErrMakerChecker      = errors.New("maker-checker violation: cannot approve own workflow")
	ErrTaskNotPending    = errors.New("task is not pending")
	ErrUnauthorized      = errors.New("unauthorized to action this task")
)

// WorkflowDefinition represents a reusable graph template.
type WorkflowDefinition struct {
	ID           string
	TenantID     string
	Code         string
	Version      int
	GraphPayload []byte // JSONB representation of the AST/DAG
	IsActive     bool
	CreatedAt    time.Time
}

// WorkflowInstance is a running execution.
type WorkflowInstance struct {
	ID               string
	TenantID         string
	DefinitionID     string
	BusinessEntityID string
	InitiatorID      string
	Status           WorkflowStatus
	TemporalRunID    string
	ContextSnapshot  []byte // JSONB
	CreatedAt        time.Time
}

// Task represents an individual unit of work for an assignee.
type Task struct {
	ID               string
	TenantID         string
	InstanceID       string
	StepNodeID       string
	AssigneeID       string
	DelegatedFromID  string
	Status           TaskStatus
	DecisionOutcome  DecisionOutcome
	DecisionComment  string
	DecisionBy       string
	SlaBreachAt      time.Time
	ResolvedAt       *time.Time
	CreatedAt        time.Time
}

// Complete handles the domain logic for finishing a task.
func (t *Task) Complete(userID string, initiatorID string, outcome DecisionOutcome, comment string) error {
	if t.Status != TaskPending && t.Status != TaskEscalated {
		return ErrTaskNotPending
	}
	// Maker-Checker enforcement
	if userID == initiatorID {
		return ErrMakerChecker
	}
	
	t.Status = TaskCompleted
	t.DecisionOutcome = outcome
	t.DecisionComment = comment
	t.DecisionBy = userID
	now := time.Now().UTC()
	t.ResolvedAt = &now
	return nil
}

// DelegationRule handles absence authority routing.
type DelegationRule struct {
	ID           string
	TenantID     string
	DelegatorID  string
	DelegateID   string
	ValidFrom    time.Time
	ValidTo      time.Time
	IsRevoked    bool
}

// IsActive returns true if the delegation is currently valid.
func (d *DelegationRule) IsActive(now time.Time) bool {
	if d.IsRevoked {
		return false
	}
	return now.After(d.ValidFrom) && now.Before(d.ValidTo)
}

// --- DAG Types ---

type NodeType string

const (
	NodeSequential NodeType = "SEQUENTIAL"
	NodeParallel   NodeType = "PARALLEL" // ALL_MUST_APPROVE
)

type WorkflowGraph struct {
	Nodes []WorkflowNode `json:"nodes"`
}

type WorkflowNode struct {
	ID             string     `json:"id"`
	Type           NodeType   `json:"type"`
	RoleExpression string     `json:"role_expression"` // e.g. "SeniorUW"
	NextNodeID     string     `json:"next_node_id,omitempty"`
}

// --- Audit Types ---

type HistoryAction string

const (
	ActionAssigned   HistoryAction = "ASSIGNED"
	ActionReassigned HistoryAction = "REASSIGNED"
	ActionEscalated  HistoryAction = "ESCALATED"
	ActionApproved   HistoryAction = "APPROVED"
	ActionRejected   HistoryAction = "REJECTED"
)

type ApprovalHistory struct {
	ID         string
	TenantID   string
	InstanceID string
	TaskID     string
	Action     HistoryAction
	ActorID    string
	Timestamp  time.Time
}

