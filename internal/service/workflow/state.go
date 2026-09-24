package service

import (
	"AgentHub/internal/model"
	message_enum "AgentHub/pkg/enum/message"
	"AgentHub/pkg/util"
	"slices"
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

// appendMessages 是 RuntimeState 消息写入的统一入口。
func (s *RuntimeState) appendMessages(messages ...*schema.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, message := range messages {
		if message != nil {
			s.Messages = append(s.Messages, message)
		}
	}
}

// snapshotMessages 返回当前有序消息历史的浅拷贝，防止调用方修改底层切片。
func (s *RuntimeState) snapshotMessages() []*schema.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]*schema.Message(nil), s.Messages...)
}

// latestToolCallMessage 返回最近一条带 ToolCalls 的 AssistantMessage。
func (s *RuntimeState) latestToolCallMessage() (*schema.Message, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, msg := range slices.Backward(s.Messages) {

		if msg != nil && msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			return msg, true
		}
	}

	return nil, false
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
