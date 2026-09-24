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

// BranchConfig 描述前端 Branch 节点的匹配规则。
// Rules 的数组顺序就是执行优先级，第一条命中的规则决定唯一的后续路径。
type BranchConfig struct {
	Version   int          `json:"version"`
	TrimSpace bool         `json:"trim_space"`
	Rules     []BranchRule `json:"rules"`
}

// BranchRule.ID 是规则与 Branch 出边之间的稳定关联键。
// Label 仅用于界面展示，不参与匹配。
type BranchRule struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Operator      string `json:"operator"`
	Value         string `json:"value"`
	CaseSensitive bool   `json:"case_sensitive"`
}

// BranchEdgeConfig 保存 Branch 规则与目标节点连线的关系。
// 默认出口使用保留值 "$default"，普通边的 config 仍为 {}。
type BranchEdgeConfig struct {
	BranchRuleID string `json:"branch_rule_id"`
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
