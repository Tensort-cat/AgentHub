package request

import (
	workflow_enum "AgentHub/pkg/enum/workflow"
	"encoding/json"
)

type NodeConfigReq struct {
	Name      string                          `json:"name"`
	Type      *workflow_enum.WorkflowNodeType `json:"type"`
	PositionX *int                            `json:"position_x"`
	PositionY *int                            `json:"position_y"`
	Config    json.RawMessage                 `json:"config"`
}
