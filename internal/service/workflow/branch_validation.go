package service

import (
	"fmt"

	"AgentHub/internal/model"
	workflow_enum "AgentHub/pkg/enum/workflow"
)

// validateBranchTopology 校验前端可见的虚拟 Branch 是否能够无歧义地
// 编译为 Eino 挂载在上游节点上的 GraphBranch。
func validateBranchTopology(
	nodes []model.WorkflowNode,
	edges []model.WorkflowEdge,
) error {
	nodeByID := make(map[int64]model.WorkflowNode, len(nodes))
	incoming := make(map[int64][]model.WorkflowEdge, len(nodes))
	outgoing := make(map[int64][]model.WorkflowEdge, len(nodes))
	for _, node := range nodes {
		nodeByID[node.ID] = node
	}
	for _, edge := range edges {
		incoming[edge.TargetNodeID] = append(incoming[edge.TargetNodeID], edge)
		outgoing[edge.SourceNodeID] = append(outgoing[edge.SourceNodeID], edge)
	}

	for _, branchNode := range nodes {
		if branchNode.Type != workflow_enum.Branch {
			continue
		}

		// Branch 自身不会成为 Eino 节点，因此只能挂载到唯一上游的输出端。
		branchIncoming := incoming[branchNode.ID]
		if len(branchIncoming) != 1 {
			return fmt.Errorf("branch node %d must have exactly one incoming edge", branchNode.ID)
		}

		upstream := nodeByID[branchIncoming[0].SourceNodeID]
		if upstream.Type == workflow_enum.Branch {
			return fmt.Errorf("branch node %d cannot follow another branch", branchNode.ID)
		}
		// 如果上游还有其他出边，Eino 会同时执行普通边和条件分支，违背单路分支语义。
		if len(outgoing[upstream.ID]) != 1 {
			return fmt.Errorf("upstream node %d of branch %d cannot have other outgoing edges", upstream.ID, branchNode.ID)
		}

		var cfg BranchConfig
		if err := decodeNodeConfig(branchNode.Config, &cfg); err != nil {
			return fmt.Errorf("branch node %d config is invalid: %w", branchNode.ID, err)
		}

		// 每条规则和默认出口都必须恰好连接一条边。
		expected := make(map[string]struct{}, len(cfg.Rules)+1)
		for _, rule := range cfg.Rules {
			expected[rule.ID] = struct{}{}
		}
		expected[branchDefaultRuleID] = struct{}{}

		connected := make(map[string]struct{}, len(expected))
		for _, edge := range outgoing[branchNode.ID] {
			target := nodeByID[edge.TargetNodeID]
			if target.Type == workflow_enum.Branch {
				return fmt.Errorf("branch node %d cannot connect directly to branch node %d", branchNode.ID, target.ID)
			}

			var edgeCfg BranchEdgeConfig
			if err := decodeNodeConfig(edge.Config, &edgeCfg); err != nil {
				return fmt.Errorf("branch edge %d config is invalid: %w", edge.ID, err)
			}
			if _, exists := expected[edgeCfg.BranchRuleID]; !exists {
				return fmt.Errorf("branch edge %d references unknown rule %q", edge.ID, edgeCfg.BranchRuleID)
			}
			if _, exists := connected[edgeCfg.BranchRuleID]; exists {
				return fmt.Errorf("branch rule %q has more than one outgoing edge", edgeCfg.BranchRuleID)
			}
			connected[edgeCfg.BranchRuleID] = struct{}{}
		}

		for ruleID := range expected {
			if _, exists := connected[ruleID]; !exists {
				return fmt.Errorf("branch node %d rule %q has no outgoing edge", branchNode.ID, ruleID)
			}
		}
	}

	return nil
}
