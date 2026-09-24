package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	// ToolCallSignal 是 ChatModel 输出 ToolCalls 时在 string 图中的内部路由信号。
	// Branch 应通过 equals 规则把该信号对应的路径连接到 Tool 节点。
	ToolCallSignal = "__agenthub_tool_call__"

	// toolResultSignal 表示 ToolsNode 已执行完毕。下一轮 ChatModel 收到该信号时，
	// 会从 RuntimeState 恢复包含 ToolCall 和 ToolMessage 的完整对话。
	toolResultSignal = "__agenthub_tool_result__"
)

// stringToMessages: string -> []*schema.Message
func stringToMessages(
	ctx context.Context,
	input string,
) ([]*schema.Message, error) {
	if input == toolResultSignal {
		var messages []*schema.Message
		if err := compose.ProcessState(
			ctx,
			func(_ context.Context, state *RuntimeState) error {
				messages = state.snapshotMessages()
				return nil
			},
		); err != nil {
			return nil, fmt.Errorf("restore messages after tool execution: %w", err)
		}

		if len(messages) == 0 || messages[len(messages)-1].Role != schema.Tool {
			return nil, fmt.Errorf("tool result signal has no corresponding tool message")
		}
		return messages, nil
	}

	return []*schema.Message{
		schema.UserMessage(input),
	}, nil
}

// messageToString: *schema.Message -> string
func messageToString(
	ctx context.Context,
	input *schema.Message,
) (string, error) {
	if input == nil {
		return "", fmt.Errorf("message is nil")
	}
	if len(input.ToolCalls) > 0 {
		return ToolCallSignal, nil
	}

	return input.Content, nil
}

// stringToToolMessage 从 RuntimeState 恢复 ChatModel 产生的完整 AssistantMessage。
// string 只承担路由信号，ToolCalls、Arguments 和 CallID 不做有损序列化。
func stringToToolMessage(
	ctx context.Context,
	input string,
) (*schema.Message, error) {
	if input != ToolCallSignal {
		return nil, fmt.Errorf("tool node requires signal %q, got %q", ToolCallSignal, input)
	}

	var message *schema.Message
	if err := compose.ProcessState(
		ctx,
		func(_ context.Context, state *RuntimeState) error {
			var ok bool
			message, ok = state.latestToolCallMessage()
			if !ok {
				return fmt.Errorf("no assistant message containing tool calls")
			}
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("restore tool call message: %w", err)
	}

	return message, nil
}

// toolMessagesToString 保存 ToolsNode 返回的原始 ToolMessage。
// ToolCallID 必须保留，否则下一轮模型无法把结果关联到对应的 ToolCall。
func toolMessagesToString(
	ctx context.Context,
	input []*schema.Message,
) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("tool node returned no messages")
	}

	for i, msg := range input {
		if msg == nil {
			return "", fmt.Errorf("tool message %d is nil", i)
		}
		if msg.Role != schema.Tool {
			return "", fmt.Errorf("tool message %d has unexpected role %q", i, msg.Role)
		}
		if msg.ToolCallID == "" {
			return "", fmt.Errorf("tool message %d has empty tool_call_id", i)
		}
	}

	if err := compose.ProcessState(
		ctx,
		func(_ context.Context, state *RuntimeState) error {
			state.appendMessages(input...)
			return nil
		},
	); err != nil {
		return "", fmt.Errorf("save tool messages: %w", err)
	}

	return toolResultSignal, nil
}

// messagesToString: []*schema.Message -> string
func messagesToString(
	ctx context.Context,
	input []*schema.Message,
) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	contents := make([]string, 0, len(input))
	for _, msg := range input {
		if msg == nil || msg.Content == "" {
			continue
		}
		contents = append(contents, msg.Content)
	}

	return strings.Join(contents, "\n\n"), nil
}

// stringToTemplateParams: string -> map[string]any
//
// ChatTemplate 统一使用 "input" 作为用户输入变量。
func stringToTemplateParams(
	ctx context.Context,
	input string,
) (map[string]any, error) {
	return map[string]any{
		"input": input,
	}, nil
}

// documentsToString: []*schema.Document -> string
func documentsToString(
	ctx context.Context,
	input []*schema.Document,
) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	contents := make([]string, 0, len(input))
	for _, doc := range input {
		if doc == nil || doc.Content == "" {
			continue
		}
		contents = append(contents, doc.Content)
	}

	return strings.Join(contents, "\n\n"), nil
}

// func stringToMap(
// 	ctx context.Context,
// 	input string,
// ) (map[string]string, error) {
// 	if len(input) == 0 {
// 		return nil, nil
// 	}

// }
