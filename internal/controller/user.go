package controller

import (
	service "AgentHub/internal/service/user"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func Me(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
	}
	id, ok := userID.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(ctx, constant.InternalServerError, constant.Error, nil)
	}
	ret, msg, data := service.Me(id)
	JsonBack(ctx, ret, msg, data)
}
