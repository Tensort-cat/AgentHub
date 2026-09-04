package request

import tool_enum "AgentHub/pkg/enum/tool"

type ToolSubscribeReq struct {
	ToolID int64                    `json:"tool_id"`
	Status tool_enum.UserToolStatus `json:"status"`
}
