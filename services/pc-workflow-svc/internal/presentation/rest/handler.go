package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"go.temporal.io/sdk/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/medhen/pc-workflow-svc/internal/domain/workflow"
)

type Handler struct {
	logger         *slog.Logger
	repo           workflow.Repository
	temporalClient client.Client
}

func NewHandler(logger *slog.Logger, repo workflow.Repository, tc client.Client) *Handler {
	return &Handler{
		logger:         logger,
		repo:           repo,
		temporalClient: tc,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/api/pc-workflow/v1/tasks/me", otelhttp.NewHandler(http.HandlerFunc(h.GetMyTasks), "GetMyTasks"))
	mux.Handle("/api/pc-workflow/v1/tasks/", otelhttp.NewHandler(http.HandlerFunc(h.TaskDecision), "TaskDecision"))
	mux.Handle("/api/pc-workflow/v1/manager/tasks", otelhttp.NewHandler(http.HandlerFunc(h.GetManagerTasks), "GetManagerTasks"))
	mux.Handle("/api/pc-workflow/v1/instances/", otelhttp.NewHandler(http.HandlerFunc(h.GetInstanceProgress), "GetInstanceProgress"))
}

func (h *Handler) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	// Mock authentication extraction
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.repo.GetPendingTasksForAssignee(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": tasks,
		"meta": map[string]int{"total": len(tasks), "page": 1},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) TaskDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Task ID from path e.g., /api/pc-workflow/v1/tasks/tsk-123/decisions
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 || parts[6] != "decisions" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	taskID := parts[5]

	userID := r.Header.Get("X-User-Id")

	var payload struct {
		Decision   workflow.DecisionOutcome `json:"decision"`
		ReasonCode string                   `json:"reason_code"`
		Comments   string                   `json:"comments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// 1. Fetch Task
	task, err := h.repo.GetTask(ctx, taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// 2. Execute Domain Logic
	// For Maker-Checker, we need the instance initiator
	instance, err := h.repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := task.Complete(userID, instance.InitiatorID, payload.Decision, payload.Comments); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// 3. Save Task
	if err := h.repo.UpdateTask(ctx, task); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Signal Temporal Workflow
	err = h.temporalClient.SignalWorkflow(ctx, instance.TemporalRunID, "", "TaskDecision", string(payload.Decision))
	if err != nil {
		h.logger.Error("Failed to signal Temporal", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetManagerTasks(w http.ResponseWriter, r *http.Request) {
	// Mock authentication extraction
	managerID := r.Header.Get("X-User-Id")
	if managerID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.repo.GetTasksForManager(r.Context(), managerID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": tasks,
		"meta": map[string]int{"total": len(tasks)},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetInstanceProgress(w http.ResponseWriter, r *http.Request) {
	// Extract Instance ID from path e.g., /api/pc-workflow/v1/instances/inst-123/progress
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 || parts[6] != "progress" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	instanceID := parts[5]

	instance, err := h.repo.GetInstance(r.Context(), instanceID)
	if err != nil {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}
	
	// Example static progress mapping. In reality, we'd query tasks and the workflow graph.
	response := map[string]interface{}{
		"instance_id":     instance.ID,
		"status":          instance.Status,
		"total_steps":     3,
		"completed_steps": 1,
		"percent_complete": 33,
		"active_nodes": []string{"step-2"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
