package request

import workflow_enum "AgentHub/pkg/enum/workflow"

type WorkflowUpdateReq struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Status      *workflow_enum.WorkflowStatus `json:"status"`
}
