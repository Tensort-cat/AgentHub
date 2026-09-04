package server

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Start() error {
	// 1. 创建 MCP Server
	s := server.NewMCPServer(
		"AgentHub MCP Server",
		"1.0.0",
	)

	// 2. 注册天气工具
	weatherTool := mcp.NewTool(
		"get_weather",
		mcp.WithDescription("获取指定城市的天气信息"),
		mcp.WithString(
			"city",
			mcp.Required(),
			mcp.Description("城市名称"),
		),
	)

	s.AddTool(
		weatherTool,
		func(
			ctx context.Context,
			request mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {

			args, ok := request.Params.Arguments.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("arguments 必须是 map")
			}

			city, ok := args["city"].(string)
			if !ok {
				return nil, fmt.Errorf("city 参数必须是字符串")
			}

			// 这里只是测试，所以暂时返回模拟数据
			return mcp.NewToolResultText(
				fmt.Sprintf(
					"%s 当前天气：晴，25°C",
					city,
				),
			), nil
		},
	)

	// 3. 创建 SSE Server
	sseServer := server.NewSSEServer(
		s,
		server.WithBaseURL("http://localhost:12345"),
	)

	log.Println("MCP SSE Server listening on http://localhost:12345")

	// 4. 启动
	return sseServer.Start("localhost:12345")
}