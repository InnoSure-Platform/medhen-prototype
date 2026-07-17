package temporal

import (
	"encoding/json"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/medhen/pc-workflow-svc/internal/domain/workflow"
)

// WorkflowParams holds initiation arguments.
type WorkflowParams struct {
	InstanceID             string
	WorkflowDefinitionCode string
	ContextSnapshot        []byte
}

// ApprovalWorkflow orchestrates the dynamic DAG defined in the Definition.
func ApprovalWorkflow(ctx workflow.Context, params WorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ApprovalWorkflow started", "InstanceID", params.InstanceID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Parse Graph from context
	var graph workflow.WorkflowGraph
	if err := json.Unmarshal(params.ContextSnapshot, &graph); err != nil {
		// Fallback to a single step if no graph is provided for simpler legacy payloads
		graph = workflow.WorkflowGraph{
			Nodes: []workflow.WorkflowNode{
				{ID: "step-1", Type: workflow.NodeSequential, RoleExpression: "SeniorUW"},
			},
		}
	}

	signalChan := workflow.GetSignalChannel(ctx, "TaskDecision")
	finalDecision := "APPROVE"

	for _, node := range graph.Nodes {
		assignParams := AssignTaskParams{
			InstanceID:      params.InstanceID,
			StepNodeID:      node.ID,
			RoleExpression:  node.RoleExpression,
			ContextSnapshot: params.ContextSnapshot,
		}

		if node.Type == workflow.NodeSequential {
			decision, err := executeNode(ctx, assignParams, signalChan)
			if err != nil {
				return err
			}
			if decision == "REJECT" {
				finalDecision = "REJECT"
				break
			}
		} else if node.Type == workflow.NodeParallel {
			// For parallel, we simulate dispatching multiple tasks to a quorum
			// e.g. "ALL_MUST_APPROVE". We will launch 2 concurrent tasks as an example.
			var futures []workflow.Future
			for i := 0; i < 2; i++ {
				f := workflow.ExecuteActivity(ctx, "AssignTaskActivity", assignParams)
				futures = append(futures, f)
			}
			
			// Wait for all assignments to complete
			for _, f := range futures {
				var taskID string
				if err := f.Get(ctx, &taskID); err != nil {
					return err
				}
				// Normally we would wait for all signals for all tasks,
				// simplified here to one wait per node level
				decision, err := waitForDecisionOrEscalate(ctx, taskID, signalChan)
				if err != nil {
					return err
				}
				if decision == "REJECT" {
					finalDecision = "REJECT"
				}
			}
			
			if finalDecision == "REJECT" {
				break
			}
		}
	}

	// Publish Completion
	err := workflow.ExecuteActivity(ctx, "NotifyWorkflowCompletedActivity", params.InstanceID, finalDecision).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to notify completion", "Error", err)
		return err
	}

	logger.Info("ApprovalWorkflow completed", "Outcome", finalDecision)
	return nil
}

func executeNode(ctx workflow.Context, params AssignTaskParams, signalChan workflow.ReceiveChannel) (string, error) {
	var taskID string
	err := workflow.ExecuteActivity(ctx, "AssignTaskActivity", params).Get(ctx, &taskID)
	if err != nil {
		return "", err
	}
	return waitForDecisionOrEscalate(ctx, taskID, signalChan)
}

func waitForDecisionOrEscalate(ctx workflow.Context, taskID string, signalChan workflow.ReceiveChannel) (string, error) {
	var decision string
	timerFuture := workflow.NewTimer(ctx, 24*time.Hour)
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &decision)
	})

	selector.AddFuture(timerFuture, func(f workflow.Future) {
		_ = workflow.ExecuteActivity(ctx, "EscalateTaskActivity", taskID).Get(ctx, nil)
		signalChan.Receive(ctx, &decision) // Wait again after escalation
	})

	selector.Select(ctx)
	return decision, nil
}
