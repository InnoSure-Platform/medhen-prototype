package workflow

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
	"pc-notification-svc/internal/application/command"
)

type Activities struct {
	dispatchHandler *command.DispatchHandler
}

func NewActivities(dh *command.DispatchHandler) *Activities {
	return &Activities{dispatchHandler: dh}
}

func (a *Activities) ExecuteDispatch(ctx context.Context, cmd command.DispatchCommand) error {
	return a.dispatchHandler.Handle(ctx, cmd)
}

// ScheduledNotificationWorkflow sleeps until the scheduled time, then dispatches
func ScheduledNotificationWorkflow(ctx workflow.Context, delay time.Duration, cmd command.DispatchCommand) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ScheduledNotificationWorkflow started", "delay", delay)

	// Sleep durably in Temporal
	err := workflow.Sleep(ctx, delay)
	if err != nil {
		return err
	}

	// Trigger the activity to dispatch
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var activities *Activities
	return workflow.ExecuteActivity(ctx, activities.ExecuteDispatch, cmd).Get(ctx, nil)
}
