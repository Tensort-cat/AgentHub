package rabbitmq

import (
    "fmt"
    "time"
)

const (
    QueueWorkflowRun    = "workflow_run_queue"
    QueueWorkflowResult = "workflow_result_queue"
    TaskStatusSuccess   = "success"
    TaskStatusFailed    = "failed"
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
    TaskID     string    `json:"task_id"`
    WorkflowID int64     `json:"workflow_id"`
    UserID     int64     `json:"user_id"`
    SessionID  int64     `json:"session_id"`
    Status     string    `json:"status"`
    Result     string    `json:"result"`
    Error      string    `json:"error,omitempty"`
    FinishedAt time.Time `json:"finished_at"`
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
