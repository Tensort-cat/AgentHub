package test

import (
	"AgentHub/internal/config"
	"testing"
)

func TestConfig(t *testing.T) {
	if err := config.InitConfig(); err != nil {
		t.Fatal("初始化配置信息失败:", err)
	}

	cfg := config.Cfg
	t.Log(cfg)
}
