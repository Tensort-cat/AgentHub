package response

import (
	workflow_enum "AgentHub/pkg/enum/workflow"
	"encoding/json"
)

type WorkflowDetailResp struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Nodes       []NodeMetaData `json:"nodes"`
	Edges       []EdgeMetaData `json:"edges"`
}

type NodeMetaData struct {
	ID        int64                          `json:"id"`
	Name      string                         `json:"name"`
	Type      workflow_enum.WorkflowNodeType `json:"type"`
	PositionX int                            `json:"position_x"`
	PositionY int                            `json:"position_y"`
	Config    json.RawMessage                `json:"config"`
}

type EdgeMetaData struct {
	ID           int64           `json:"id"`
	SourceNodeID int64           `json:"source_node_id"`
	TargetNodeID int64           `json:"target_node_id"`
	Config       json.RawMessage `json:"config"`
}
