package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitModel(api *gin.RouterGroup) {
	// 列出当前用户已配置的模型
	api.GET("", controller.ModelList)

	// 创建模型
	api.POST("", controller.ModelCreate)

	// 更新模型配置信息
	api.PUT("/:id", controller.ModelUpdate)

	// 删除模型
	api.DELETE("/:id", controller.DeleteModel)
}
