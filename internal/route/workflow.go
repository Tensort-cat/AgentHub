package route

import (
	"AgentHub/internal/controller"

	"github.com/gin-gonic/gin"
)

func InitWorkflow(api *gin.RouterGroup) {
	// 分页查询工作流
	api.GET("", controller.WorkflowPage)

	// 创建工作流
	api.POST("", controller.WorkflowCreate)

	// 获取工作流详情
	api.GET("/:id", controller.WorkflowDetail)

	// 更新工作流元信息
	api.PUT("/:id", controller.WordflowUpdate)

	// 删除工作流
	api.DELETE("/:id", controller.WorkflowDelete)

	// 运行工作流
	api.POST("/run", controller.Run)
	api.GET("/run/:task_id/events", controller.WorkflowRunEvents)

	// 节点相关接口
	nodeGroup := api.Group("/:id/nodes")
	{
		// 创建节点
		nodeGroup.POST("", controller.NodeCreate)

		// 更新节点
		nodeGroup.PUT("/:node_id", controller.NodeUpdate)

		// 删除节点
		nodeGroup.DELETE("/:node_id", controller.NodeDelete)
	}

	// 边相关接口
	edgeGroup := api.Group("/:id/edges")
	{
		// 创建边
		edgeGroup.POST("", controller.EdgeCreate)

		// 删除边
		edgeGroup.DELETE("/:edge_id", controller.EdgeDelete)
	}
}
