package request

import tool_enum "AgentHub/pkg/enum/tool"

type ToolEnableOrDisableReq struct {
	ToolID int64                `json:"tool_id"`
	Status tool_enum.ToolStatus `json:"status"`
}
