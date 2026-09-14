package service

import (
	"AgentHub/internal/model"
	message_enum "AgentHub/pkg/enum/message"
	"AgentHub/pkg/util"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// state.go

type RuntimeState struct {
	SessionID int64             // 每次运行对应一个新的 session
	Input     string            // 图的原始输入
	Messages  []*schema.Message // 运行过程中产生的信息
	Variables map[string]map[string]any

	mu sync.RWMutex // 读写锁
}

func (s *RuntimeState) SaveMsg2DB() error {
	s.mu.RLock()

	messages := make([]*model.Message, 0, len(s.Messages))

	hs := map[schema.RoleType]message_enum.MessageType{
		schema.User:      message_enum.MessageTypeUser,
		schema.Assistant: message_enum.MessageTypeAssistant,
		schema.Tool:      message_enum.MessageTypeTool,
		schema.System:    message_enum.MessageTypeSystem,
	}

	for _, msg := range s.Messages {
		messages = append(messages, &model.Message{
			BaseModel: model.BaseModel{
				ID:        util.GenID(),
				CreatedAt: time.Now(),
			},
			SessionID: s.SessionID,
			Type:      hs[msg.Role],
			Content:   msg.Content,
		})
	}

	s.mu.RUnlock()

	return saveMessages(messages)
}

func newRuntimeState(sessionID int64, input string) *RuntimeState {
	return &RuntimeState{
		SessionID: sessionID,
		Input:     input,
		Messages: []*schema.Message{
			schema.UserMessage(input),
		},
		Variables: make(map[string]map[string]any),
	}
}
