package response

import (
	tool_enum "AgentHub/pkg/enum/tool"
	"time"
)

type ToolDetailResp struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      tool_enum.ToolStatus `json:"status"`
	Avatar      string               `json:"avatar"`
	CreateAt    time.Time            `json:"created_at"`
}
