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

		endpoint, err := addNode(ctx, graph, node)
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
) (nodeEndpoint, error) {

	switch node.Type {

	case workflow_enum.ChatTemplate:
		return addChatTemplateNode(graph, node)

	case workflow_enum.ChatModel:
		return addChatModelNode(ctx, graph, node)

	case workflow_enum.Tool:
		return addToolNode(ctx, graph, node)

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

	var sessionID int64

	if err := compose.ProcessState(
		ctx,
		func(_ context.Context, state *RuntimeState) error {
			sessionID = state.SessionID

			state.Variables[nodeKey] = map[string]any{
				"system_prompt": cfg.SystemPrompt,
			}

			return nil
		},
	); err != nil {
		return nodeEndpoint{}, err
	}

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
		ctx context.Context,
		output *schema.Message,
		state *RuntimeState,
	) (*schema.Message, error) {

		if output == nil {
			return nil, fmt.Errorf("ChatModel 输出为空")
		}

		return output, compose.ProcessState(
			ctx,
			func(_ context.Context, state *RuntimeState) error {
				state.Messages = append(
					state.Messages,
					output,
				)
				return nil
			},
		)
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
) (nodeEndpoint, error) {

	var cfg ToolConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	var sessionID int64

	if err := compose.ProcessState(
		ctx,
		func(_ context.Context, state *RuntimeState) error {
			sessionID = state.SessionID
			return nil
		},
	); err != nil {
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

	post := func(
		ctx context.Context,
		output string,
		state *RuntimeState,
	) (string, error) {

		return output, compose.ProcessState(
			ctx,
			func(_ context.Context, state *RuntimeState) error {

				state.Messages = append(
					state.Messages,
					&schema.Message{
						Role:    schema.Tool,
						Content: output,
					},
				)

				return nil
			},
		)
	}

	if err := graph.AddToolsNode(
		key,
		toolsNode,
		compose.WithStatePostHandler(post),
		compose.WithNodeName(node.Name),
	); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: key,
		Exit:  key,
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
func addBranchNode(
	ctx context.Context,
	graph *compose.Graph[string, string],
	node my_model.WorkflowNode,
	edges []model.WorkflowEdge,
	endpoints map[int64]nodeEndpoint,
) (nodeEndpoint, error) {

	var cfg BranchConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return nodeEndpoint{}, err
	}

	// Branch 没有真实的输入输出节点。
	// 它挂载在自己的上游节点 Exit 上。
	var startNode string

	for _, edge := range edges {
		if edge.TargetNodeID == node.ID {
			source, ok := endpoints[edge.SourceNodeID]
			if !ok {
				return nodeEndpoint{}, fmt.Errorf(
					"Branch 上游节点不存在: %d",
					edge.SourceNodeID,
				)
			}

			startNode = source.Exit
			break
		}
	}

	if startNode == "" {
		return nodeEndpoint{}, fmt.Errorf(
			"Branch 节点没有上游节点: %d",
			node.ID,
		)
	}

	endNodes := make(map[string]bool)

	for _, targetID := range cfg.Conditions {

		targetNodeID, err := strconv.ParseInt(
			targetID,
			10,
			64,
		)
		if err != nil {
			continue
		}

		target, ok := endpoints[targetNodeID]
		if !ok {
			return nodeEndpoint{}, fmt.Errorf(
				"Branch 目标节点不存在: %d",
				targetNodeID,
			)
		}

		endNodes[target.Entry] = true
	}

	branch := compose.NewGraphBranch(
		func(
			ctx context.Context,
			input string,
		) (string, error) {

			targetID, ok := cfg.Conditions[input]
			if !ok {
				return compose.END, nil
			}

			targetNodeID, err := strconv.ParseInt(
				targetID,
				10,
				64,
			)
			if err != nil {
				return "", fmt.Errorf(
					"非法 Branch 目标节点: %s",
					targetID,
				)
			}

			target, ok := endpoints[targetNodeID]
			if !ok {
				return "", fmt.Errorf(
					"Branch 目标节点不存在: %d",
					targetNodeID,
				)
			}

			return target.Entry, nil
		},
		endNodes,
	)

	if err := graph.AddBranch(startNode, branch); err != nil {
		return nodeEndpoint{}, err
	}

	return nodeEndpoint{
		Entry: startNode,
		Exit:  startNode,
	}, nil
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
