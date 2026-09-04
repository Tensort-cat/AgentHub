package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitSession(api *gin.RouterGroup) {

	// 分页查询会话
	api.GET("", controller.SessionPage)

	// 分页获取会话的消息列表
	api.GET("/:id", controller.SessionMsgPage)

}
