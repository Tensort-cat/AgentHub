package service

import (
	"AgentHub/pkg/zlog"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/client"
)

// 用于管理一次工作流运行产生的 MCP Client
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
	mcmOnce       sync.Once
)

func GetMcpCliManager() *MCPCliManager {
	mcmOnce.Do(func() {
		mcpCliManager = &MCPCliManager{
			clis: make(map[int64][]*client.Client),
		}
	})

	return mcpCliManager
}
