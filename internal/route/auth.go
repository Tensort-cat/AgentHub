package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitAuth(api *gin.RouterGroup) {
	api.POST("/login", controller.Login)
	api.POST("/register", controller.Register)
	api.POST("/sendCaptcha", controller.SendCaptcha)
}
