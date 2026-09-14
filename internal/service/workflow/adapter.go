package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// stringToMessages: string -> []*schema.Message
func stringToMessages(
	ctx context.Context,
	input string,
) ([]*schema.Message, error) {
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

	return input.Content, nil
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
