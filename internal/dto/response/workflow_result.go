package response

import (
	workflow_enum "AgentHub/pkg/enum/workflow"
	"time"
)

type WorkflowResult struct {
	SessionID int64  `json:"session_id"`
	Content   string `json:"content"`
}

// WorkflowRunAccepted is returned after a workflow run has been accepted by
// RabbitMQ. The actual result is produced asynchronously and is correlated by
// TaskID.
type WorkflowRunAccepted struct {
	TaskID string                          `json:"task_id"`
	Status workflow_enum.WorkflowRunStatus `json:"status"`
}

type WorkflowRunState struct {
	TaskID    string                          `json:"task_id"`
	Status    workflow_enum.WorkflowRunStatus `json:"status"`
	SessionID int64                           `json:"session_id,omitempty"`
	Result    string                          `json:"result,omitempty"`
	Error     string                          `json:"error,omitempty"`
	UpdatedAt time.Time                       `json:"updated_at"`
}
