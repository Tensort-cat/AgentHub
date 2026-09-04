package model

import (
	tool_enum "AgentHub/pkg/enum/tool"
	"time"
)

type UserTools struct {
	UserID    int64
	ToolID    int64
	Status    tool_enum.UserToolStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserTools) TableName() string {
	return "user_tools"
}
