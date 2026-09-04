package response

import (
	message_enum "AgentHub/pkg/enum/message"
	"time"
)

type SessionMsgPageItem struct {
	ID        int64                    `json:"id"`
	Content   string                   `json:"content"`
	Type      message_enum.MessageType `json:"type"`
	CreatedAt time.Time                `json:"created_at"`
}
