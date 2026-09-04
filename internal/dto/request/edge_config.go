package request

type EdgeConfigReq struct {
	SourceNodeID int64 `json:"source_node_id" binding:"required"`
	TargetNodeID int64 `json:"target_node_id" binding:"required"`
}
