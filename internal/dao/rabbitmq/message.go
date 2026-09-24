package rabbitmq

import (
	workflow_enum "AgentHub/pkg/enum/workflow"
	"fmt"
	"time"
)

const (
	QueueWorkflowRun    = "workflow_run_queue"
	QueueWorkflowResult = "workflow_result_queue"
)

type WorkflowRunTask struct {
	TaskID     string    `json:"task_id"`
	WorkflowID int64     `json:"workflow_id"`
	UserID     int64     `json:"user_id"`
	SessionID  int64     `json:"session_id,omitempty"`
	Input      string    `json:"input"`
	CreatedAt  time.Time `json:"created_at"`
}

type WorkflowRunResult struct {
	TaskID     string                          `json:"task_id"`
	WorkflowID int64                           `json:"workflow_id"`
	UserID     int64                           `json:"user_id"`
	SessionID  int64                           `json:"session_id"`
	Status     workflow_enum.WorkflowRunStatus `json:"status"`
	Result     string                          `json:"result"`
	Error      string                          `json:"error,omitempty"`
	FinishedAt time.Time                       `json:"finished_at"`
}

func NewWorkflowRunTask(workflowID int64, input string, userID int64) WorkflowRunTask {
	return WorkflowRunTask{
		TaskID:     fmt.Sprintf("wf-%d-%d", time.Now().UnixNano(), workflowID),
		WorkflowID: workflowID,
		UserID:     userID,
		Input:      input,
		CreatedAt:  time.Now().UTC(),
	}
}
