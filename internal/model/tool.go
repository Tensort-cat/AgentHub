package model

import tool_enum "AgentHub/pkg/enum/tool"

type Tool struct {
	BaseModel

	Name        string             `gorm:"column:name;type:varchar(100);not null"`
	Description string             `gorm:"column:description;type:text"`
	Avatar      string             `gorm:"column:avatar;type:varchar(255)"`
	Type        tool_enum.ToolType `gorm:"column:type;not null;default:1"`

	MCPServerURL string `gorm:"column:mcp_server_url;type:varchar(500)"`
	MCPToolName  string `gorm:"column:mcp_tool_name;type:varchar(100)"`

	Status tool_enum.ToolStatus `gorm:"column:status;not null;default:1"`
}

func (Tool) TableName() string {
	return "tools"
}
