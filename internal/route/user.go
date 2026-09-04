package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitUser(api *gin.RouterGroup) {
	// 获取当前用户信息
	api.GET("/me", controller.Me)
}
