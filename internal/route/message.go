package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitMessage(api *gin.RouterGroup) {
	// 创建新消息
	api.POST("", controller.MessageCreate)
}
