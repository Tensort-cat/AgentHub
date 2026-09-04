package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/knowledge_base"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func KbList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}
	id, ok := userID.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	ret, msg, data := service.List(id)
	JsonBack(c, ret, msg, data)
}

func KbCreate(c *gin.Context) {
	var req request.KbCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error("参数绑定失败", zap.Error(err))
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}
	id, ok := userID.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	ret, msg := service.Create(req.Name, req.Description, id, req.EmbedderID)
	JsonBack(c, ret, msg, nil)
}

func KbDetail(c *gin.Context) {
	kbID := c.Param("id")
	if kbID == "" {
		zlog.Error("知识库ID不能为空")
		JsonBack(c, constant.BadRequest, "知识库 ID 为空", nil)
		return
	}

	ret, msg, data := service.Detail(kbID)
	JsonBack(c, ret, msg, data)
}
