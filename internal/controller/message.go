package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/message"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func MessageCreate(c *gin.Context) {
	var req request.MessageCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, "参数绑定错误", nil)
		return
	}

	ret, msg := service.Create(req.SessionID, req.Content, req.Type)
	JsonBack(c, ret, msg, nil)
}
