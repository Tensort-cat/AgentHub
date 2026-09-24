package workflow_enum

type WorkflowStatus int8

const (
	DRAFT WorkflowStatus = iota + 1
	ACTIVE
)

type WorkflowRunStatus string

const (
	RunStatusQueued  WorkflowRunStatus = "queued"
	RunStatusRunning WorkflowRunStatus = "running"
	RunStatusSuccess WorkflowRunStatus = "success"
	RunStatusFailed  WorkflowRunStatus = "failed"
)

func (s WorkflowRunStatus) IsTerminal() bool {
	return s == RunStatusSuccess || s == RunStatusFailed
}
