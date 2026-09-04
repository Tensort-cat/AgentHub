package client

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// NewClient 创建并初始化一个 SSE MCP Client。
func NewClient(ctx context.Context, serverURL string) (*client.Client, error) {
	c, err := client.NewSSEMCPClient(serverURL)
	if err != nil {
		return nil, fmt.Errorf("create MCP client: %w", err)
	}

	if err := c.Start(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("start MCP client: %w", err)
	}

	initRequest := mcp.InitializeRequest{
		Params: struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    mcp.ClientCapabilities `json:"capabilities"`
			ClientInfo      mcp.Implementation     `json:"clientInfo"`
		}{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "AgentHub",
				Version: "1.0.0",
			},
		},
	}

	if _, err := c.Initialize(ctx, initRequest); err != nil {
		c.Close()
		return nil, fmt.Errorf("initialize MCP client: %w", err)
	}

	return c, nil
}
