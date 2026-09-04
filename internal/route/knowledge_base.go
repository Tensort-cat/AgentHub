package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitKnowledgeBase(api *gin.RouterGroup) {
	// 列出知识库
	api.GET("", controller.KbList)

	// 创建知识库
	api.POST("", controller.KbCreate)

	// 获取知识库详情与文件列表
	api.GET("/:id", controller.KbDetail)
}
