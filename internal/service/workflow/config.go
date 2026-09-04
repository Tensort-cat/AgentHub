package service

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ChatModelConfig struct {
	ModelID      int64   `json:"model_id"`
	Temperature  float32 `json:"temperature"`
	MaxTokens    int     `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt"`

	ToolIDs []int64 `json:"tool_ids"`
}

type ChatTemplateConfig struct {
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
}

type BranchConfig struct {
	// 第一版具体字段根据你前端最终设计确定。
	//
	// 例如：
	// Conditions []BranchCondition `json:"conditions"`
}

type BranchCondition struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

type ToolConfig struct {
	ToolIDs []int64 `json:"tool_ids"`
}

type AgentConfig struct {
	ToolCallingModel int64 `json:"tool_calling_model"` // 发起工具调用的模型 ID

	// 待定
}

type RetrieverConfig struct {
	KbID int64 `json:"kb_id"`
	TopK int   `json:"top_k"`
}

func decodeNodeConfig(
	raw []byte,
	target any,
) error {

	if len(bytes.TrimSpace(raw)) == 0 {
		return fmt.Errorf("config 不能为空")
	}

	// 必须是 JSON object
	var object map[string]json.RawMessage

	if err := json.Unmarshal(raw, &object); err != nil {
		return fmt.Errorf("config 不是合法 JSON: %w", err)
	}

	if object == nil {
		return fmt.Errorf("config 必须是 JSON object")
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("config 字段非法: %w", err)
	}

	return nil
}
