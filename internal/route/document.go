package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitDocument(api *gin.RouterGroup) {
	// 上传文档并加入指定知识库, 并 embedding 到 Redis
	api.POST("", controller.DocUpload)

	// 下载文档
	api.GET(":id/download", controller.DocDownload)
}
