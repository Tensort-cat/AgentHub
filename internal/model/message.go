package model

import message_enum "AgentHub/pkg/enum/message"

type Message struct {
	BaseModel

	SessionID int64                    `gorm:"column:session_id;index;not null"`
	Type      message_enum.MessageType `gorm:"column:type;not null"`
	Content   string                   `gorm:"column:content;type:longtext;not null"`
}

func (Message) TableName() string {
	return "messages"
}
