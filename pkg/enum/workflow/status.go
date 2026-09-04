package workflow_enum

type WorkflowStatus int8

const (
	DRAFT WorkflowStatus = iota + 1
	ACTIVE
)
