package route

import (
	"AgentHub/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var GE *gin.Engine

func InitRoute() {
	GE = gin.Default()

	// 跨域配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "Authorization"}
	GE.Use(cors.New(corsConfig))

	// 初始化路由
	api := GE.Group("/api/v1")
	{
		// 认证相关
		authRouter := api.Group("/auth")
		InitAuth(authRouter)
	}
	{
		// 文件与静态资源相关
		staticRouter := api.Group("/static")
		staticRouter.Use(middleware.Auth())
		InitStatic(staticRouter)
	}
	{
		// 用户相关
		userRouter := api.Group("/users")
		userRouter.Use(middleware.Auth())
		InitUser(userRouter)
	}
	{
		// 模型相关
		modelRouter := api.Group("/models")
		modelRouter.Use(middleware.Auth())
		InitModel(modelRouter)
	}
	{
		// 工作流相关
		workflowRouter := api.Group("/workflows")
		workflowRouter.Use(middleware.Auth())
		InitWorkflow(workflowRouter)
	}
	{
		// 会话相关
		sessionRouter := api.Group("/sessions")
		sessionRouter.Use(middleware.Auth())
		InitSession(sessionRouter)
	}
	{
		// 消息相关
		messageRouter := api.Group("/messages")
		messageRouter.Use(middleware.Auth())
		InitMessage(messageRouter)
	}
	{
		// 知识库相关
		knowledgeRouter := api.Group("/kb")
		knowledgeRouter.Use(middleware.Auth())
		InitKnowledgeBase(knowledgeRouter)
	}
	{
		// 文档相关
		documentRouter := api.Group("/docs")
		documentRouter.Use(middleware.Auth())
		InitDocument(documentRouter)
	}
	{
		// 工具中心相关
		toolRouter := api.Group("/tools")
		toolRouter.Use(middleware.Auth())
		InitTool(toolRouter)
	}
	{
		// 用户工具相关
		userToolsRouter := api.Group("/user_tools")
		userToolsRouter.Use(middleware.Auth())
		InitUserTools(userToolsRouter)
	}
}
