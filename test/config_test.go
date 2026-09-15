package test

import (
	"AgentHub/internal/config"
	"AgentHub/pkg/zlog"
	"testing"
)

func TestConfig(t *testing.T) {
	if err := config.InitConfig(); err != nil {
		t.Fatal("初始化配置信息失败:", err)
	}

	cfg := config.Cfg
	t.Log(cfg)
}

func TestLogger(t *testing.T) {
	zlog.Debug("这是测试")
}
