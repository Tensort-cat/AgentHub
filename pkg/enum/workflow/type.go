package workflow_enum

type WorkflowNodeType int8

const (
	ChatModel WorkflowNodeType = iota + 1
	ChatTemplate
	Branch
	Tool
	Retriever
	Agent
	Start
	End
)
