package service

import (
	"context"
	"fmt"
	"sync"

	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	my_model "AgentHub/internal/model"
	my_client "AgentHub/internal/service/mcp/client"
	model_enum "AgentHub/pkg/enum/model"
	tool_enum "AgentHub/pkg/enum/tool"
	workflow_enum "AgentHub/pkg/enum/workflow"
	"AgentHub/pkg/zlog"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	eino_model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
)

type RuntimeState struct {
	SessionID int64

	Input string

	Messages []*schema.Message

	Variables map[string]any

	LastOutput string
}

func newRuntimeState(sessionID int64, input string) *RuntimeState {
	return &RuntimeState{
		SessionID: sessionID,
		Input:     input,
		Messages:  make([]*schema.Message, 0),
		Variables: make(map[string]any),
	}
}

// 用于管理一次工作流运行的 MCP Client
// 采用单例模式
type MCPCliManager struct {
	clis map[int64][]*client.Client
	mu   sync.Mutex
}

func (mcm *MCPCliManager) Register(sessionID int64) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	if _, exists := mcm.clis[sessionID]; !exists {
		mcm.clis[sessionID] = make([]*client.Client, 0)
	}
}

func (mcm *MCPCliManager) Login(sessionID int64, cli *client.Client) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	mcm.clis[sessionID] = append(mcm.clis[sessionID], cli)
}

func (mcm *MCPCliManager) Sweep(sessionID int64) {
	mcm.mu.Lock()

	clis, ok := mcm.clis[sessionID]
	if ok {
		delete(mcm.clis, sessionID)
	}

	mcm.mu.Unlock()

	if !ok {
		return
	}

	for _, cli := range clis {
		if err := cli.Close(); err != nil {
			zlog.Error(
				fmt.Sprintf(
					"关闭 MCP Client 失败 session=%d: %v",
					sessionID,
					err,
				),
			)
		}
	}
}

var (
	mcpCliManager *MCPCliManager
	once          sync.Once
)

func GetMcpCliManager() *MCPCliManager {
	once.Do(func() {
		mcpCliManager = &MCPCliManager{
			clis: make(map[int64][]*client.Client),
		}
	})

	return mcpCliManager
}

type RuntimeInput struct {
	msg *schema.Message
}

type RuntimeOutput struct {
	msg *schema.Message
}

// BuildGraph 将数据库中的 WorkflowNode / WorkflowEdge
// 转换为 Eino Graph。
func BuildGraph(
	ctx context.Context,
	nodes []my_model.WorkflowNode,
	edges []my_model.WorkflowEdge,
	sessionID int64,
) (*compose.Graph[RuntimeInput, RuntimeOutput], error) {

	nodeMap := make(map[int64]my_model.WorkflowNode)

	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	graph := compose.NewGraph[RuntimeInput, RuntimeOutput](
		compose.WithGenLocalState(func(ctx context.Context) *RuntimeState {
			return newRuntimeState(sessionID, "")
		}),
	)

	// 在 mcpCliManager 登记该次 session
	GetMcpCliManager().Register(sessionID)

	// 添加节点
	for _, node := range nodes {
		if err := addNode(ctx, graph, node, sessionID); err != nil {
			return nil, fmt.Errorf(
				"添加节点失败 node=%d type=%d: %w",
				node.ID,
				node.Type,
				err,
			)
		}
	}

	// 添加边
	for _, edge := range edges {
		if err := addEdge(graph, edge, nodeMap); err != nil {
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
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
	sessionID int64,
) error {
	switch node.Type {

	case workflow_enum.Start:
		return nil

	case workflow_enum.End:
		return nil

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

	case workflow_enum.Branch:
		return addBranchNode(ctx, graph, node)

	default:
		return fmt.Errorf("不支持的节点类型: %v", node.Type)
	}
}

func addChatTemplateNode(
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
) error {
	// 提取节点配置信息
	var cfg ChatTemplateConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return err
	}
	// 创建 ChatTemplate
	tmp := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(cfg.SystemPrompt),
		schema.UserMessage(cfg.UserPrompt),
	)

	// 将节点加入 graph
	return graph.AddChatTemplateNode(node.Name, tmp)
}

func addChatModelNode(
	ctx context.Context,
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
	sessionID int64,
) error {
	// 从节点获取配置信息
	var cfg ChatModelConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return err
	}

	cm, err := newChatModel(ctx, cfg)
	if err != nil {
		return err
	}

	if len(cfg.ToolIDs) > 0 {
		tools, err := getTools(ctx, cfg.ToolIDs, sessionID)
		if err != nil {
			return err
		}
		infos := make([]*schema.ToolInfo, 0, len(tools))

		for _, t := range tools {
			info, err := t.Info(ctx)
			if err != nil {
				return err
			}

			infos = append(infos, info)
		}

		cm, err = cm.WithTools(infos)
		if err != nil {
			return err
		}
	}

	// 为节点上下文添加系统提示词
	preHandler := func(
		ctx context.Context,
		input map[string]any,
		state *RuntimeState,
	) (map[string]any, error) {
		input["system_prompt"] = cfg.SystemPrompt
		return input, nil
	}

	return graph.AddChatModelNode(
		node.Name,
		cm,
		compose.WithStatePreHandler(preHandler),
	)
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

// TODO: 分支节点
func addBranchNode(
	ctx context.Context,
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
) error {

	// 将节点加入 graph

	return nil
}

// TODO: 工具节点
func addToolNode(
	ctx context.Context,
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
	sessionID int64,
) error {
	// 从节点获取配置信息
	var cfg ToolConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return err
	}

	tools, err := getTools(ctx, cfg.ToolIDs, sessionID)
	if err != nil {
		return err
	}
	toolsNode, err := compose.NewToolNode(ctx,
		&compose.ToolsNodeConfig{Tools: tools},
	)
	if err != nil {
		return err
	}

	return graph.AddToolsNode(node.Name, toolsNode)
}

func addRetrieverNode(
	ctx context.Context,
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
) error {

	// 1. 解析节点配置
	var cfg RetrieverConfig
	if err := decodeNodeConfig(node.Config, &cfg); err != nil {
		return err
	}

	// 2. 创建绑定指定知识库的 Retriever
	kbRet, err := NewKBRetriever(ctx, cfg.KbID, cfg.TopK)
	if err != nil {
		return err
	}

	// 3. 添加到 Graph
	return graph.AddRetrieverNode(node.Name, kbRet)
}

// TODO: Agent 节点
func addAgentNode(
	ctx context.Context,
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	node my_model.WorkflowNode,
) error {
	return nil
}

func addEdge(
	graph *compose.Graph[RuntimeInput, RuntimeOutput],
	edge my_model.WorkflowEdge,
	nodeMap map[int64]my_model.WorkflowNode,
) error {
	// 将两条边连接
	source := nodeMap[edge.SourceNodeID]
	target := nodeMap[edge.TargetNodeID]

	return graph.AddEdge(source.Name, target.Name)
}
