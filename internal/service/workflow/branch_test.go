package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"AgentHub/internal/model"
	workflow_enum "AgentHub/pkg/enum/workflow"

	"gorm.io/datatypes"
)

func TestBranchEvaluatorOperators(t *testing.T) {
	tests := []struct {
		name          string
		operator      string
		value         string
		input         string
		caseSensitive bool
	}{
		{name: "equals", operator: branchOperatorEquals, value: "APPROVE", input: "approve"},
		{name: "contains", operator: branchOperatorContains, value: "批准", input: "已经批准请求", caseSensitive: true},
		{name: "starts with", operator: branchOperatorStartsWith, value: "result:", input: "RESULT: ok"},
		{name: "ends with", operator: branchOperatorEndsWith, value: "done", input: "task DONE"},
		{name: "regex", operator: branchOperatorRegex, value: `^order-\d+$`, input: "ORDER-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := newBranchEvaluator(BranchConfig{
				Version:   branchConfigVersion,
				TrimSpace: true,
				Rules: []BranchRule{{
					ID:            "matched",
					Operator:      tt.operator,
					Value:         tt.value,
					CaseSensitive: tt.caseSensitive,
				}},
			})
			if err != nil {
				t.Fatalf("newBranchEvaluator() error = %v", err)
			}

			if got := evaluator.Match("  " + tt.input + "  "); got != "matched" {
				t.Fatalf("Match() = %q, want matched", got)
			}
		})
	}
}

func TestBranchEvaluatorUsesFirstMatchAndDefault(t *testing.T) {
	evaluator, err := newBranchEvaluator(BranchConfig{
		Version: branchConfigVersion,
		Rules: []BranchRule{
			{ID: "first", Operator: branchOperatorContains, Value: "ok"},
			{ID: "second", Operator: branchOperatorEquals, Value: "ok"},
		},
	})
	if err != nil {
		t.Fatalf("newBranchEvaluator() error = %v", err)
	}
	if got := evaluator.Match("ok"); got != "first" {
		t.Fatalf("Match() = %q, want first", got)
	}
	if got := evaluator.Match("no"); got != branchDefaultRuleID {
		t.Fatalf("Match() = %q, want %q", got, branchDefaultRuleID)
	}
}

func TestValidateBranchConfig(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "unknown field", raw: `{"version":1,"trim_space":true,"rules":[{"id":"ok","label":"","operator":"equals","value":"ok","case_sensitive":false}],"extra":true}`},
		{name: "invalid id", raw: `{"version":1,"rules":[{"id":"bad id","operator":"equals","value":"ok"}]}`},
		{name: "duplicate id", raw: `{"version":1,"rules":[{"id":"same","operator":"equals","value":"a"},{"id":"same","operator":"equals","value":"b"}]}`},
		{name: "reserved id", raw: `{"version":1,"rules":[{"id":"$default","operator":"equals","value":"ok"}]}`},
		{name: "invalid operator", raw: `{"version":1,"rules":[{"id":"ok","operator":"greater_than","value":"ok"}]}`},
		{name: "invalid regex", raw: `{"version":1,"rules":[{"id":"ok","operator":"regex","value":"["}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBranchConfig([]byte(tt.raw)); err == nil {
				t.Fatal("validateBranchConfig() error = nil")
			}
		})
	}
}

func TestValidateBranchTopology(t *testing.T) {
	nodes := branchTestNodes(t, false)
	validEdges := []model.WorkflowEdge{
		branchTestEdge(t, 11, 1, 2, ""),
		branchTestEdge(t, 12, 2, 3, "approve"),
		branchTestEdge(t, 13, 2, 3, branchDefaultRuleID),
	}

	tests := []struct {
		name  string
		edges []model.WorkflowEdge
	}{
		{name: "missing incoming", edges: validEdges[1:]},
		{name: "duplicate incoming", edges: append(append([]model.WorkflowEdge{}, validEdges...), branchTestEdge(t, 14, 3, 2, ""))},
		{name: "missing default", edges: validEdges[:2]},
		{name: "missing rule", edges: []model.WorkflowEdge{validEdges[0], validEdges[2]}},
		{name: "duplicate handle", edges: append(append([]model.WorkflowEdge{}, validEdges...), branchTestEdge(t, 14, 2, 3, "approve"))},
		{name: "unknown handle", edges: []model.WorkflowEdge{validEdges[0], validEdges[1], branchTestEdge(t, 14, 2, 3, "unknown")}},
		{name: "upstream also bypasses branch", edges: append(append([]model.WorkflowEdge{}, validEdges...), branchTestEdge(t, 14, 1, 3, ""))},
	}

	if err := validateBranchTopology(nodes, validEdges); err != nil {
		t.Fatalf("valid topology error = %v", err)
	}
	if err := validateEdges(model.Workflow{BaseModel: model.BaseModel{ID: 99}}, nodes, validEdges); err != nil {
		t.Fatalf("same-target branch edges should be valid: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBranchTopology(nodes, tt.edges); err == nil {
				t.Fatal("validateBranchTopology() error = nil")
			}
		})
	}
}

func TestValidateBranchTopologyRejectsAdjacentBranches(t *testing.T) {
	nodes := branchTestNodes(t, false)
	nodes = append(nodes, model.WorkflowNode{
		BaseModel:  model.BaseModel{ID: 4},
		WorkflowID: 99,
		Name:       "second branch",
		Type:       workflow_enum.Branch,
		Config:     nodes[1].Config,
	})
	edges := []model.WorkflowEdge{
		branchTestEdge(t, 11, 1, 2, ""),
		branchTestEdge(t, 12, 2, 4, "approve"),
		branchTestEdge(t, 13, 2, 3, branchDefaultRuleID),
	}

	if err := validateBranchTopology(nodes, edges); err == nil {
		t.Fatal("validateBranchTopology() error = nil")
	}
}

func TestBuildGraphBranchRoutesAndPreservesInput(t *testing.T) {
	nodes := branchTestNodes(t, true)
	edges := []model.WorkflowEdge{
		branchTestEdge(t, 11, 1, 2, ""),
		branchTestEdge(t, 12, 2, 3, "approve"),
		branchTestEdge(t, 13, 2, 4, branchDefaultRuleID),
		branchTestEdge(t, 14, 3, 5, ""),
		branchTestEdge(t, 15, 4, 5, ""),
	}

	graph, err := BuildGraph(context.Background(), "", nodes, edges, 9001)
	if err != nil {
		t.Fatalf("BuildGraph() error = %v", err)
	}
	defer Sweep(9001)

	runner, err := graph.Compile(context.Background())
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	matched, err := runner.Invoke(context.Background(), "  APPROVE  ")
	if err != nil {
		t.Fatalf("matched Invoke() error = %v", err)
	}
	if !strings.Contains(matched, "matched:  APPROVE  ") {
		t.Fatalf("matched output %q did not preserve original input", matched)
	}

	fallback, err := runner.Invoke(context.Background(), "REJECT")
	if err != nil {
		t.Fatalf("default Invoke() error = %v", err)
	}
	if !strings.Contains(fallback, "default:REJECT") {
		t.Fatalf("default output = %q", fallback)
	}
}

func branchTestNodes(t *testing.T, withTargets bool) []model.WorkflowNode {
	t.Helper()
	workflowID := int64(99)
	nodes := []model.WorkflowNode{
		{BaseModel: model.BaseModel{ID: 1}, WorkflowID: workflowID, Name: "start", Type: workflow_enum.Start, Config: datatypes.JSON(`{}`)},
		{BaseModel: model.BaseModel{ID: 2}, WorkflowID: workflowID, Name: "branch", Type: workflow_enum.Branch, Config: branchTestJSON(t, BranchConfig{
			Version:   branchConfigVersion,
			TrimSpace: true,
			Rules: []BranchRule{{
				ID:       "approve",
				Label:    "approve",
				Operator: branchOperatorEquals,
				Value:    "APPROVE",
			}},
		})},
	}
	if withTargets {
		nodes = append(nodes,
			model.WorkflowNode{BaseModel: model.BaseModel{ID: 3}, WorkflowID: workflowID, Name: "matched", Type: workflow_enum.ChatTemplate, Config: branchTestJSON(t, ChatTemplateConfig{UserPrompt: "matched:{input}"})},
			model.WorkflowNode{BaseModel: model.BaseModel{ID: 4}, WorkflowID: workflowID, Name: "default", Type: workflow_enum.ChatTemplate, Config: branchTestJSON(t, ChatTemplateConfig{UserPrompt: "default:{input}"})},
			model.WorkflowNode{BaseModel: model.BaseModel{ID: 5}, WorkflowID: workflowID, Name: "end", Type: workflow_enum.End, Config: datatypes.JSON(`{}`)},
		)
	} else {
		nodes = append(nodes,
			model.WorkflowNode{BaseModel: model.BaseModel{ID: 3}, WorkflowID: workflowID, Name: "end", Type: workflow_enum.End, Config: datatypes.JSON(`{}`)},
		)
	}
	return nodes
}

func branchTestEdge(t *testing.T, id, source, target int64, ruleID string) model.WorkflowEdge {
	t.Helper()
	config := datatypes.JSON(`{}`)
	if ruleID != "" {
		config = branchTestJSON(t, BranchEdgeConfig{BranchRuleID: ruleID})
	}
	return model.WorkflowEdge{
		BaseModel:    model.BaseModel{ID: id},
		WorkflowID:   99,
		SourceNodeID: source,
		TargetNodeID: target,
		Config:       config,
	}
}

func branchTestJSON(t *testing.T, value any) datatypes.JSON {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return datatypes.JSON(raw)
}
