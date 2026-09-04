package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitTool(api *gin.RouterGroup) {
	// 列出可用工具（卡片展示）
	api.GET("", controller.ToolList)

	// 工具详情，包含 provider、配置字段等。
	api.GET(":id", controller.ToolDetail)

	// 启用/禁用工具
	api.POST("", controller.ToolEnableOrDisable)
}

func InitUserTools(api *gin.RouterGroup) {
	// 用户订阅/取消订阅工具
	api.POST("", controller.ToolSubscribe)
}
