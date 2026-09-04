package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/model"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func ModelList(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}
	id, ok := userID.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(ctx, constant.InternalServerError, constant.Error, nil)
		return
	}
	ret, msg, data := service.List(id)
	JsonBack(ctx, ret, msg, data)
}

func ModelCreate(ctx *gin.Context) {
	var req request.ModelConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}
	id, ok := userID.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(ctx, constant.InternalServerError, constant.Error, nil)
		return
	}

	ret, msg := service.Create(req.Name, req.BaseUrl, req.ApiKey, id, req.Type)
	JsonBack(ctx, ret, msg, nil)
}

func ModelUpdate(ctx *gin.Context) {
	var req request.ModelConfigReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}

	modelID := ctx.Param("id")
	ret, msg := service.Update(modelID, req.Name, req.BaseUrl, req.ApiKey)
	JsonBack(ctx, ret, msg, nil)
}

func DeleteModel(ctx *gin.Context) {
	modelID := ctx.Param("id")
	ret, msg := service.Delete(modelID)
	JsonBack(ctx, ret, msg, nil)
}
