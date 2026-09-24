package service

import (
	"encoding/json"
	"fmt"

	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	tool_enum "AgentHub/pkg/enum/tool"
	workflow_enum "AgentHub/pkg/enum/workflow"
)

func validateNodes(
	wf model.Workflow,
	nodes []model.WorkflowNode,
) error {

	if len(nodes) == 0 {
		return fmt.Errorf("工作流 %d 不存在节点", wf.ID)
	}

	nodeIDs := make(map[int64]struct{}, len(nodes))

	startCnt := 0
	endCnt := 0

	for _, node := range nodes {

		// 1. 节点必须属于当前工作流
		if node.WorkflowID != wf.ID {
			return fmt.Errorf(
				"节点 %d 不属于工作流 %d",
				node.ID,
				wf.ID,
			)
		}

		// 2. 节点 ID 不能重复
		if _, exists := nodeIDs[node.ID]; exists {
			return fmt.Errorf(
				"存在重复的节点 ID: %d",
				node.ID,
			)
		}

		nodeIDs[node.ID] = struct{}{}

		// 3. 节点类型必须合法
		if !isValidNodeType(node.Type) {
			return fmt.Errorf(
				"节点 %d 的类型非法: %v",
				node.ID,
				node.Type,
			)
		}

		// 4. 节点名称不能为空
		if node.Name == "" {
			return fmt.Errorf(
				"节点 %d 的名称不能为空",
				node.ID,
			)
		}

		// 5. 统计 Start / End
		switch node.Type {
		case workflow_enum.Start:
			startCnt++

		case workflow_enum.End:
			endCnt++
		}
	}

	// 6. 必须且只能存在一个 Start
	if startCnt != 1 {
		return fmt.Errorf(
			"Start 节点数量必须为 1，实际为 %d",
			startCnt,
		)
	}

	// 7. 必须且只能存在一个 End
	if endCnt != 1 {
		return fmt.Errorf(
			"End 节点数量必须为 1，实际为 %d",
			endCnt,
		)
	}

	return nil
}

func validateEdges(
	wf model.Workflow,
	nodes []model.WorkflowNode,
	edges []model.WorkflowEdge,
) error {

	// 建立 Node ID 集合
	nodeIDs := make(map[int64]struct{}, len(nodes))
	nodeByID := make(map[int64]model.WorkflowNode, len(nodes))

	var startID, endID int64

	for _, node := range nodes {
		nodeIDs[node.ID] = struct{}{}
		nodeByID[node.ID] = node

		switch node.Type {
		case workflow_enum.Start:
			startID = node.ID

		case workflow_enum.End:
			endID = node.ID
		}
	}

	// 用于检测重复边
	edgeSet := make(map[string]struct{}, len(edges))

	for _, edge := range edges {

		// 1. Edge 必须属于当前 Workflow
		if edge.WorkflowID != wf.ID {
			return fmt.Errorf(
				"边 %d 不属于工作流 %d",
				edge.ID,
				wf.ID,
			)
		}

		// 2. Source Node 必须存在
		if _, exists := nodeIDs[edge.SourceNodeID]; !exists {
			return fmt.Errorf(
				"边 %d 的 source 节点不存在: %d",
				edge.ID,
				edge.SourceNodeID,
			)
		}

		// 3. Target Node 必须存在
		if _, exists := nodeIDs[edge.TargetNodeID]; !exists {
			return fmt.Errorf(
				"边 %d 的 target 节点不存在: %d",
				edge.ID,
				edge.TargetNodeID,
			)
		}

		// 4. 不允许自环
		if edge.SourceNodeID == edge.TargetNodeID {
			return fmt.Errorf(
				"节点 %d 不允许连接到自身",
				edge.SourceNodeID,
			)
		}

		// 5. Start 不允许存在入边
		if edge.TargetNodeID == startID {
			return fmt.Errorf(
				"Start 节点不能存在入边: %d -> Start",
				edge.SourceNodeID,
			)
		}

		// 6. End 不允许存在出边
		if edge.SourceNodeID == endID {
			return fmt.Errorf(
				"End 节点不能存在出边: End -> %d",
				edge.TargetNodeID,
			)
		}

		// 7. 普通节点不允许重复 source/target；Branch 还需带上规则 ID。
		// 这样不同规则可以合法地连接到同一个目标节点。
		key := fmt.Sprintf("%d:%d", edge.SourceNodeID, edge.TargetNodeID)
		if nodeByID[edge.SourceNodeID].Type == workflow_enum.Branch {
			var cfg BranchEdgeConfig
			if err := decodeNodeConfig(edge.Config, &cfg); err != nil {
				return fmt.Errorf("branch edge %d config is invalid: %w", edge.ID, err)
			}
			key += ":" + cfg.BranchRuleID
		}

		if _, exists := edgeSet[key]; exists {
			return fmt.Errorf(
				"存在重复的边: %d -> %d",
				edge.SourceNodeID,
				edge.TargetNodeID,
			)
		}

		edgeSet[key] = struct{}{}
	}

	return nil
}

func validateGraph(
	nodes []model.WorkflowNode,
	edges []model.WorkflowEdge,
) error {

	// =========================
	// 1. 找到 Start / End
	// =========================

	var startID, endID int64

	for _, node := range nodes {
		switch node.Type {
		case workflow_enum.Start:
			startID = node.ID

		case workflow_enum.End:
			endID = node.ID
		}
	}

	if startID == 0 {
		return fmt.Errorf("不存在 Start 节点")
	}

	if endID == 0 {
		return fmt.Errorf("不存在 End 节点")
	}

	// =========================
	// 2. 构建正向图
	// =========================

	graph := make(map[int64][]int64, len(nodes))

	for _, edge := range edges {
		graph[edge.SourceNodeID] =
			append(graph[edge.SourceNodeID], edge.TargetNodeID)
	}

	// =========================
	// 3. 从 Start DFS
	// =========================

	visitedFromStart := make(map[int64]bool, len(nodes))

	stack := []int64{startID}
	visitedFromStart[startID] = true

	for len(stack) > 0 {

		// pop
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, next := range graph[current] {

			if visitedFromStart[next] {
				continue
			}

			visitedFromStart[next] = true
			stack = append(stack, next)
		}
	}

	// 所有节点都必须能够从 Start 到达
	if len(visitedFromStart) != len(nodes) {

		for _, node := range nodes {
			if !visitedFromStart[node.ID] {
				return fmt.Errorf(
					"节点 %d (%s) 无法从 Start 节点到达",
					node.ID,
					node.Name,
				)
			}
		}
	}

	// =========================
	// 4. 构建反向图
	// =========================

	reverseGraph := make(map[int64][]int64, len(nodes))

	for _, edge := range edges {
		reverseGraph[edge.TargetNodeID] =
			append(reverseGraph[edge.TargetNodeID], edge.SourceNodeID)
	}

	// =========================
	// 5. 从 End 反向 DFS
	// =========================

	visitedToEnd := make(map[int64]bool, len(nodes))

	stack = []int64{endID}
	visitedToEnd[endID] = true

	for len(stack) > 0 {

		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, prev := range reverseGraph[current] {

			if visitedToEnd[prev] {
				continue
			}

			visitedToEnd[prev] = true
			stack = append(stack, prev)
		}
	}

	// 所有节点最终都必须能够到达 End
	if len(visitedToEnd) != len(nodes) {

		for _, node := range nodes {
			if !visitedToEnd[node.ID] {
				return fmt.Errorf(
					"节点 %d (%s) 无法到达 End 节点",
					node.ID,
					node.Name,
				)
			}
		}
	}

	return nil
}

func validateNodeConfigs(nodes []model.WorkflowNode, userID int64) error {

	for _, node := range nodes {

		switch node.Type {

		case workflow_enum.Start:
			if err := validateStartConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.End:
			if err := validateEndConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.ChatTemplate:
			if err := validateChatTemplateConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.ChatModel:
			if err := validateChatModelConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.Branch:
			if err := validateBranchConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.Tool:
			if err := validateToolConfig(node.Config, userID); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		case workflow_enum.Retriever:
			if err := validateRetrieverConfig(node.Config); err != nil {
				return fmt.Errorf(
					"节点 %d (%s) Config 不合法: %w",
					node.ID,
					node.Name,
					err,
				)
			}

		// case workflow_enum.Agent:
		// 	if err := validateAgentConfig(node.Config); err != nil {
		// 		return fmt.Errorf(
		// 			"节点 %d (%s) Config 不合法: %w",
		// 			node.ID,
		// 			node.Name,
		// 			err,
		// 		)
		// 	}

		default:
			return fmt.Errorf(
				"节点 %d 类型未知: %v",
				node.ID,
				node.Type,
			)
		}
	}

	return nil
}

func validateStartConfig(raw []byte) error {
	return validateEmptyConfig(raw)
}

func validateEndConfig(raw []byte) error {
	return validateEmptyConfig(raw)
}

func validateEmptyConfig(raw []byte) error {

	var config map[string]json.RawMessage

	if err := json.Unmarshal(raw, &config); err != nil {
		return fmt.Errorf("config 不是合法 JSON: %w", err)
	}

	if config == nil {
		return fmt.Errorf("config 必须为 {}")
	}

	if len(config) != 0 {
		return fmt.Errorf("该节点不允许配置字段")
	}

	return nil
}

func validateChatModelConfig(raw []byte) error {

	var cfg ChatModelConfig

	if err := decodeNodeConfig(raw, &cfg); err != nil {
		return err
	}

	if cfg.ModelID <= 0 {
		return fmt.Errorf("model_id 必须大于 0")
	}

	if cfg.Temperature < 0 || cfg.Temperature > 2 {
		return fmt.Errorf(
			"temperature 必须在 [0, 2] 范围内",
		)
	}

	if cfg.MaxTokens <= 0 {
		return fmt.Errorf(
			"max_tokens 必须大于 0",
		)
	}

	return nil
}

func validateChatTemplateConfig(raw []byte) error {

	var cfg ChatTemplateConfig

	if err := decodeNodeConfig(raw, &cfg); err != nil {
		return err
	}

	if cfg.SystemPrompt == "" &&
		cfg.UserPrompt == "" {
		return fmt.Errorf(
			"system_prompt 和 user_prompt 不能同时为空",
		)
	}

	if cfg.UserPrompt == "" {
		return fmt.Errorf(
			"user_prompt 不能为空",
		)
	}

	return nil
}

func validateBranchConfig(raw []byte) error {
	var cfg BranchConfig
	if err := decodeNodeConfig(raw, &cfg); err != nil {
		return err
	}

	_, err := newBranchEvaluator(cfg)
	return err
}

func validateRetrieverConfig(raw []byte) error {
	var cfg RetrieverConfig

	if err := decodeNodeConfig(raw, &cfg); err != nil {
		return err
	}

	if cfg.KbID <= 0 {
		return fmt.Errorf(
			"kb_id 必须大于 0",
		)
	}

	if cfg.TopK < 1 {
		return fmt.Errorf(
			"top_k 必须大于 0",
		)
	}
	return nil

}

func validateToolConfig(raw []byte, userID int64) error {
	var cfg ToolConfig

	if err := decodeNodeConfig(raw, &cfg); err != nil {
		return err
	}

	// 用户只能使用自己订阅过的工具
	// 查询工具订阅表，检查 config.ToolIDs 是否在用户的订阅列表中
	var userTools []model.UserTools
	if err := dao.DB.
		Model(&model.UserTools{}).
		Where("user_id = ? and status = ?", userID, tool_enum.Subscribed).
		Find(&userTools).Error; err != nil {
		return err
	}
	check := make(map[int64]struct{}, len(userTools))
	for _, tool := range userTools {
		check[tool.ToolID] = struct{}{}
	}

	for _, toolID := range cfg.ToolIDs {
		if _, exists := check[toolID]; !exists {
			return fmt.Errorf("工具 ID %d 未在用户的订阅列表中", toolID)
		}
	}

	return nil
}
