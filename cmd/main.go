package main

import (
	"AgentHub/internal/config"
	"AgentHub/internal/dao"
	"AgentHub/internal/dao/rabbitmq"
	"AgentHub/internal/route"
	workflowService "AgentHub/internal/service/workflow"
	"AgentHub/pkg/zlog"
	"context"
	"fmt"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// 初始化配置信息
	if err := config.InitConfig(); err != nil {
		zlog.Error("初始化配置信息失败", zap.Error(err))
		return
	}

	// 初始化 MySQL 连接
	if err := dao.InitMySQL(); err != nil {
		zlog.Error("初始化 MySQL 连接失败", zap.Error(err))
		return
	}
	defer dao.CloseMySQL()

	// 初始化 Redis 连接
	if err := dao.InitRedisCli(ctx); err != nil {
		zlog.Error("初始化 Redis 客户端失败", zap.Error(err))
		return
	}
	defer dao.CloseRedis()

	// 初始化 RabbitMQ 连接
	if err := rabbitmq.InitRabbitmq(); err != nil {
		zlog.Error("初始化 rabbitmq 连接失败", zap.Error(err))
		return
	}
	defer rabbitmq.Close()
	if err := rabbitmq.StartWorkflowResultConsumer(ctx, workflowService.HandleRunResult); err != nil {
		zlog.Error("启动工作流结果消费者失败", zap.Error(err))
		return
	}
	if err := rabbitmq.StartWorkflowConsumer(ctx, 3, workflowService.HandleRunTask); err != nil {
		zlog.Error("启动工作流消费者失败", zap.Error(err))
		return
	}

	// 开启 web 服务
	port := fmt.Sprintf(":%d", config.Cfg.MainConfig.Port)
	zlog.Info("backend port", zap.String("port", port))
	route.InitRoute()
	if err := route.GE.Run(port); err != nil {
		zlog.Error("启动 Web 服务失败", zap.String("error", err.Error()))
		return
	}
}
