package service

import (
	"context"
	"fmt"
	"strconv"

	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	my_model "AgentHub/internal/model"
	my_client "AgentHub/internal/service/mcp/client"
	model_enum "AgentHub/pkg/enum/model"
	tool_enum "AgentHub/pkg/enum/tool"
	workflow_enum "AgentHub/pkg/enum/workflow"

	eino_model "github.com/cloudwego/eino/components/model"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type nodeEndpoint struct {
	Entry string
	Exit  string
}

func BuildGraph(
	ctx context.Context,
	input string,
	nodes []my_model.WorkflowNode,
	edges []my_model.WorkflowEdge,
	sessionID int64,
) (*compose.Graph[string, string], error) {

	nodeMap := make(map[int64]my_model.WorkflowNode, len(nodes))
	endpoints := make(map[int64]nodeEndpoint, len(nodes))

	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	graph := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *RuntimeState {
			return newRuntimeState(sessionID, input)
		}),
	)

	GetMcpCliManager().Register(sessionID)

	// 1. 添加普通节点
	for _, node := range nodes {
		if node.Type == workflow_enum.Start ||
			node.Type == workflow_enum.End ||
			node.Type == workflow_enum.Branch {
			continue
		}

		endpoint, err := addNode(ctx, graph, node, sessionID)
		if err != nil {
			return nil, fmt.Errorf(
				"添加节点失败 node=%d type=%d: %w",
				node.ID,
				node.Type,
				err,
			)
		}

		endpoints[node.ID] = endpoint
	}

	// 2. 添加 Branch (分支节点在 eino 里比较特殊，单独用一个逻辑)
	for _, node := range nodes {
		if node.Type != workflow_enum.Branch {
			continue
		}

		endpoint, err := addBranchNode(
			ctx,
			graph,
			node,
			edges,
			nodeMap,
			endpoints,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"添加 Branch 失败 node=%d: %w",
				node.ID,
				err,
			)
		}

		endpoints[node.ID] = endpoint
	}

	// 3. 添加边
	for _, edge := range edges {
		if err := addEdge(
			graph,
			edge,
			nodeMap,
			endpoints,
		); err != nil {
			return nil, fmt.Errorf(
				"添加边失败 edge=%d: %w",
				edge.ID,
				err,
			)
		}
	}

	return graph, nil
}

func Sweep(sessionID int64) {
	GetMcpCliManager().Sweep(sessionID)
}

func addNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
	sessionID int64,
) (nodeEndpoint, error) {

	switch node.Type {

	case workflow_enum.ChatTemplate:
		return addChatTemplateNode(graph, node)

	case workflow_enum.ChatModel:
		return addChatModelNode(ctx, graph, node, sessionID)

	case workflow_enum.Tool:
		return addToolNode(ctx, graph, node, sessionID)

	case workflow_enum.Retriever:
		return addRetrieverNode(ctx, graph, node)

	case workflow_enum.Agent:
		return addAgentNode(ctx, graph, node)

	default:
		return nodeEndpoint{}, fmt.Errorf(
			"不支持的节点类型: %v",
			node.Type,
		)
	}
}

// ---------------------------------------------------------
// ChatTemplate
// string -> map[string]any -> []*schema.Message -> string
// ---------------------------------------------------------
func addChatTemplateNode(
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
) (nodeEndpoint, error) {

	var cfg ChatTemplateConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	tmp := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(cfg.SystemPrompt),
		schema.UserMessage(cfg.UserPrompt),
	)

	key := strconv.FormatInt(node.ID, 10)

	inputAdapter := key + "__input"
	componentNode := key + "__component"
	outputAdapter := key + "__output"

	if err := graph.AddLambdaNode(
		inputAdapter,
		compose.InvokableLambda(stringToTemplateParams),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddChatTemplateNode(
		componentNode,
		tmp,
		compose.WithNodeName(node.Name),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddLambdaNode(
		outputAdapter,
		compose.InvokableLambda(messagesToString),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(inputAdapter, componentNode); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(componentNode, outputAdapter); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: inputAdapter,
		Exit:  outputAdapter,
	}, nil
}

// ---------------------------------------------------------
// ChatModel
// string -> []*schema.Message -> *schema.Message -> string
// ---------------------------------------------------------
func addChatModelNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
	sessionID int64,
) (nodeEndpoint, error) {

	var cfg ChatModelConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	cm, err := newChatModel(ctx, cfg)
	if err != nil {
		return nodeEndpoint{}, err
	}

	nodeKey := strconv.FormatInt(node.ID, 10)

	if len(cfg.ToolIDs) > 0 {

		tools, err := getTools(ctx, cfg.ToolIDs, sessionID)
		if err != nil {
			return nodeEndpoint{}, err
		}

		infos := make([]*schema.ToolInfo, 0, len(tools))

		for _, t := range tools {
			info, err := t.Info(ctx)
			if err != nil {
				return nodeEndpoint{}, err
			}

			infos = append(infos, info)
		}

		// 告诉模型能够调用的工具信息
		cm, err = cm.WithTools(infos)
		if err != nil {
			return nodeEndpoint{}, err
		}
	}

	inputAdapter := nodeKey + "__input"
	componentNode := nodeKey + "__component"
	outputAdapter := nodeKey + "__output"

	// string -> []*schema.Message
	if err := graph.AddLambdaNode(
		inputAdapter,
		compose.InvokableLambda(stringToMessages),
	); err != nil {
		return nodeEndpoint{}, err
	}

	pre := func(
		ctx context.Context,
		input []*schema.Message,
		state *RuntimeState,
	) ([]*schema.Message, error) {

		if cfg.SystemPrompt == "" {
			return input, nil
		}

		// 把系统提示词放第一个
		msgs := make([]*schema.Message, 0, len(input)+1)
		msgs = append(
			msgs,
			schema.SystemMessage(cfg.SystemPrompt),
		)
		msgs = append(msgs, input...)

		return msgs, nil
	}

	post := func(
		_ context.Context,
		output *schema.Message,
		state *RuntimeState,
	) (*schema.Message, error) {

		if output == nil {
			return nil, fmt.Errorf("ChatModel 输出为空")
		}

		state.appendMessages(output)
		return output, nil
	}

	if err := graph.AddChatModelNode(
		componentNode,
		cm,
		compose.WithStatePreHandler(pre),
		compose.WithStatePostHandler(post),
		compose.WithNodeName(node.Name),
	); err != nil {
		return nodeEndpoint{}, err
	}

	// *schema.Message -> string
	if err := graph.AddLambdaNode(
		outputAdapter,
		compose.InvokableLambda(messageToString),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(inputAdapter, componentNode); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(componentNode, outputAdapter); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: inputAdapter,
		Exit:  outputAdapter,
	}, nil
}

func newChatModel(
	ctx context.Context,
	cfg ChatModelConfig,
) (eino_model.ToolCallingChatModel, error) {

	var model my_model.Model
	if err := dao.DB.
		First(&model, "id = ?", cfg.ModelID).Error; err != nil {
		return nil, fmt.Errorf("获取模型信息失败: %w", err)
	}

	switch model.Provider {
	case model_enum.ModelProviderOpenAI:
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:      model.APIKey,
			Model:       model.Name,
			Temperature: &cfg.Temperature,
			MaxTokens:   &cfg.MaxTokens,
		})

	case model_enum.ModelProviderArk:
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:      model.APIKey,
			Model:       model.Name,
			Temperature: &cfg.Temperature,
			MaxTokens:   &cfg.MaxTokens,
		})

	default:
		return nil, fmt.Errorf("不支持的 ChatModel Provider: %v", model.Provider)
	}
}

// ---------------------------------------------------------
// Retriever
// string -> []*schema.Document -> string
// ---------------------------------------------------------
func addRetrieverNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
) (nodeEndpoint, error) {

	var cfg RetrieverConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	kbRet, err := NewKBRetriever(
		ctx,
		cfg.KbID,
		cfg.TopK,
	)
	if err != nil {
		return nodeEndpoint{}, err
	}

	key := strconv.FormatInt(node.ID, 10)

	// Retriever 节点不需要输入适配器, 因为它需要的输入就是 string
	componentNode := key + "__component"
	outputAdapter := key + "__output"

	// string -> []*schema.Document
	if err := graph.AddRetrieverNode(
		componentNode,
		kbRet,
		compose.WithNodeName(node.Name),
	); err != nil {
		return nodeEndpoint{}, err
	}

	// []*schema.Document -> string
	if err := graph.AddLambdaNode(
		outputAdapter,
		compose.InvokableLambda(documentsToString),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(
		componentNode,
		outputAdapter,
	); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: componentNode,
		Exit:  outputAdapter,
	}, nil
}

// ---------------------------------------------------------
// Tool
// ---------------------------------------------------------
func addToolNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
	sessionID int64,
) (nodeEndpoint, error) {

	var cfg ToolConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	tools, err := getTools(
		ctx,
		cfg.ToolIDs,
		sessionID,
	)
	if err != nil {
		return nodeEndpoint{}, err
	}

	toolsNode, err := compose.NewToolNode(
		ctx,
		&compose.ToolsNodeConfig{
			Tools: tools,
		},
	)
	if err != nil {
		return nodeEndpoint{}, err
	}

	key := strconv.FormatInt(node.ID, 10)
	inputAdapter := key + "__input"
	componentNode := key + "__component"
	outputAdapter := key + "__output"

	// string 控制信号 -> 包含 ToolCalls 的完整 AssistantMessage。
	if err := graph.AddLambdaNode(
		inputAdapter,
		compose.InvokableLambda(stringToToolMessage),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddToolsNode(
		componentNode,
		toolsNode,
		compose.WithNodeName(node.Name),
	); err != nil {
		return nodeEndpoint{}, err
	}

	// ToolMessage 列表写入 RuntimeState，并转回 string 控制信号。
	if err := graph.AddLambdaNode(
		outputAdapter,
		compose.InvokableLambda(toolMessagesToString),
	); err != nil {
		return nodeEndpoint{}, err
	}

	if err := graph.AddEdge(inputAdapter, componentNode); err != nil {
		return nodeEndpoint{}, err
	}
	if err := graph.AddEdge(componentNode, outputAdapter); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: inputAdapter,
		Exit:  outputAdapter,
	}, nil
}

func getTools(
	ctx context.Context,
	toolIDs []int64,
	sessionID int64,
) ([]tool.BaseTool, error) {

	if len(toolIDs) == 0 {
		return nil, nil
	}

	// 1. 查询工具
	var tools []model.Tool

	if err := dao.DB.
		Where("id IN ?", toolIDs).
		Find(&tools).Error; err != nil {
		return nil, fmt.Errorf("查询工具失败: %w", err)
	}

	if len(tools) != len(toolIDs) {
		return nil, fmt.Errorf("部分工具不存在")
	}

	// 2. 按 MCP Server 分组
	serverTools := make(map[string][]string)

	for _, t := range tools {

		if t.Status != tool_enum.Enable {
			return nil, fmt.Errorf(
				"工具 %d (%s) 已被禁用",
				t.ID,
				t.Name,
			)
		}

		if t.Type != tool_enum.ToolTypeMCP {
			return nil, fmt.Errorf(
				"工具 %d (%s) 暂不支持该工具类型",
				t.ID,
				t.Name,
			)
		}

		if t.MCPServerURL == "" {
			return nil, fmt.Errorf(
				"工具 %d (%s) 未配置 MCP Server",
				t.ID,
				t.Name,
			)
		}

		if t.MCPToolName == "" {
			return nil, fmt.Errorf(
				"工具 %d (%s) 未配置 MCP Tool Name",
				t.ID,
				t.Name,
			)
		}

		serverTools[t.MCPServerURL] =
			append(serverTools[t.MCPServerURL], t.MCPToolName)
	}

	// 3. 从 MCP Server 获取工具
	allTools := make([]tool.BaseTool, 0, len(toolIDs))

	for serverURL, toolNames := range serverTools {

		cli, err := my_client.NewClient(ctx, serverURL)
		if err != nil {
			return nil, err
		}

		// 登记 MCP Client
		GetMcpCliManager().Login(sessionID, cli)

		mcpTools, err := mcp.GetTools(ctx, &mcp.Config{
			Cli:          cli,
			ToolNameList: toolNames,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"获取 MCP Tools 失败: %w",
				err,
			)
		}

		allTools = append(allTools, mcpTools...)
	}

	return allTools, nil
}

// ---------------------------------------------------------
// Branch
// ---------------------------------------------------------
// addBranchNode 将数据库中的可视化 Branch 节点编译为 Eino GraphBranch。
// Branch 没有真实的执行节点：它挂载在上游 endpoint 上，并把原始 string
// 直接转发到命中规则对应的下游 endpoint。
func addBranchNode(
	_ context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
	edges []model.WorkflowEdge,
	nodeMap map[int64]my_model.WorkflowNode,
	endpoints map[int64]nodeEndpoint,
) (nodeEndpoint, error) {
	var cfg BranchConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}
	evaluator, err := newBranchEvaluator(cfg)
	if err != nil {
		return nodeEndpoint{}, err
	}

	// 可视化的 upstream -> Branch 边只用于定位 GraphBranch 的挂载点，
	// 不会作为普通边添加到 Eino graph。
	var incoming *model.WorkflowEdge
	for i := range edges {
		if edges[i].TargetNodeID != node.ID {
			continue
		}
		if incoming != nil {
			return nodeEndpoint{}, fmt.Errorf("branch node %d has more than one incoming edge", node.ID)
		}
		incoming = &edges[i]
	}
	if incoming == nil {
		return nodeEndpoint{}, fmt.Errorf("branch node %d has no incoming edge", node.ID)
	}

	source, exists := nodeMap[incoming.SourceNodeID]
	if !exists {
		return nodeEndpoint{}, fmt.Errorf("branch upstream node does not exist: %d", incoming.SourceNodeID)
	}
	// Start/End 在 Eino 中是保留 endpoint，不存在于 endpoints 映射中。
	startNode := compose.START
	if source.Type != workflow_enum.Start {
		endpoint, ok := endpoints[source.ID]
		if !ok {
			return nodeEndpoint{}, fmt.Errorf("branch upstream endpoint does not exist: %d", source.ID)
		}
		startNode = endpoint.Exit
	}

	allowedRuleIDs := make(map[string]struct{}, len(cfg.Rules)+1)
	for _, rule := range cfg.Rules {
		allowedRuleIDs[rule.ID] = struct{}{}
	}
	allowedRuleIDs[branchDefaultRuleID] = struct{}{}

	// targetByRule 用于运行时选择目标；endNodes 是 Eino 要求预先声明的目标白名单。
	targetByRule := make(map[string]string, len(cfg.Rules)+1)
	endNodes := make(map[string]bool, len(cfg.Rules)+1)
	for _, edge := range edges {
		if edge.SourceNodeID != node.ID {
			continue
		}

		var edgeCfg BranchEdgeConfig
		if err := decodeNodeConfig(edge.Config, &edgeCfg); err != nil {
			return nodeEndpoint{}, fmt.Errorf("branch edge %d config is invalid: %w", edge.ID, err)
		}
		if _, ok := allowedRuleIDs[edgeCfg.BranchRuleID]; !ok {
			return nodeEndpoint{}, fmt.Errorf("branch edge %d references unknown rule %q", edge.ID, edgeCfg.BranchRuleID)
		}
		if _, duplicate := targetByRule[edgeCfg.BranchRuleID]; duplicate {
			return nodeEndpoint{}, fmt.Errorf("branch rule %q has more than one target", edgeCfg.BranchRuleID)
		}

		target, ok := nodeMap[edge.TargetNodeID]
		if !ok {
			return nodeEndpoint{}, fmt.Errorf("branch target node does not exist: %d", edge.TargetNodeID)
		}
		targetEndpoint := compose.END
		if target.Type != workflow_enum.End {
			endpoint, ok := endpoints[target.ID]
			if !ok {
				return nodeEndpoint{}, fmt.Errorf("branch target endpoint does not exist: %d", target.ID)
			}
			targetEndpoint = endpoint.Entry
		}

		targetByRule[edgeCfg.BranchRuleID] = targetEndpoint
		endNodes[targetEndpoint] = true
	}

	for _, rule := range cfg.Rules {
		if _, ok := targetByRule[rule.ID]; !ok {
			return nodeEndpoint{}, fmt.Errorf("branch rule %q has no target", rule.ID)
		}
	}
	if _, ok := targetByRule[branchDefaultRuleID]; !ok {
		return nodeEndpoint{}, fmt.Errorf("branch default rule has no target")
	}

	branch := compose.NewGraphBranch(
		func(_ context.Context, input string) (string, error) {
			// evaluator 只返回路由键；Eino 会把未修改的 input 交给选中的下游节点。
			ruleID := evaluator.Match(input)
			target, ok := targetByRule[ruleID]
			if !ok {
				return "", fmt.Errorf("branch rule %q has no target", ruleID)
			}
			return target, nil
		},
		endNodes,
	)
	if err := graph.AddBranch(startNode, branch); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{Entry: startNode, Exit: startNode}, nil
}

// ---------------------------------------------------------
// Agent
// ---------------------------------------------------------
func addAgentNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
) (nodeEndpoint, error) {

	// TODO: Agent
	return nodeEndpoint{}, fmt.Errorf("Agent 节点暂未实现")
}

// ---------------------------------------------------------
// Edge
// ---------------------------------------------------------
func addEdge(
	graph *compose.Graph[string, string],
	edge my_model.WorkflowEdge,
	nodeMap map[int64]my_model.WorkflowNode,
	endpoints map[int64]nodeEndpoint,
) error {

	source, ok := nodeMap[edge.SourceNodeID]
	if !ok {
		return fmt.Errorf(
			"源节点不存在: %d",
			edge.SourceNodeID,
		)
	}

	target, ok := nodeMap[edge.TargetNodeID]
	if !ok {
		return fmt.Errorf(
			"目标节点不存在: %d",
			edge.TargetNodeID,
		)
	}

	// Branch 的边由 Branch 自己管理。
	if target.Type == workflow_enum.Branch ||
		source.Type == workflow_enum.Branch {
		return nil
	}

	sourceEndpoint := compose.START
	targetEndpoint := compose.END

	if source.Type != workflow_enum.Start {
		sourceEndpoint = endpoints[source.ID].Exit
	}

	if target.Type != workflow_enum.End {
		targetEndpoint = endpoints[target.ID].Entry
	}

	return graph.AddEdge(
		sourceEndpoint,
		targetEndpoint,
	)
}
