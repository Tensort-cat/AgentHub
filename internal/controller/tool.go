package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/tool"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func ToolList(c *gin.Context) {
	ret, msg, data := service.List()
	JsonBack(c, ret, msg, data)
}

func ToolDetail(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		zlog.Warn("缺少必要参数")
		JsonBack(c, constant.BadRequest, "缺少必要参数", nil)
	}

	ret, msg, data := service.Detail(toolID)
	JsonBack(c, ret, msg, data)
}

func ToolSubscribe(c *gin.Context) {
	var req request.ToolSubscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	userIDAny, exists := c.Get("user_id")
	if !exists {
		zlog.Warn("用户未登录")
		JsonBack(c, constant.Unauthorized, "用户未登录", nil)
		return
	}

	userID, ok := userIDAny.(int64)
	if !ok {
		zlog.Warn("断言失败")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	ret, msg := service.Subscribe(userID, req.ToolID, req.Status)
	JsonBack(c, ret, msg, nil)
}

func ToolEnableOrDisable(c *gin.Context) {
	var req request.ToolEnableOrDisableReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	ret, msg := service.EnableOrDisable(req.ToolID, req.Status)
	JsonBack(c, ret, msg, nil)
}
