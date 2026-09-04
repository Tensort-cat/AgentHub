package request

import message_enum "AgentHub/pkg/enum/message"

type MessageCreateReq struct {
	SessionID int64                    `json:"session_id" binding:"required"`
	Content   string                   `json:"content" binding:"required"`
	Type      message_enum.MessageType `json:"type" binding:"required"`
}
