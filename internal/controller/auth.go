package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/user"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var req request.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}

	ret, msg, data := service.Login(req.Email, req.Password)
	JsonBack(ctx, ret, msg, data)
}

func Register(ctx *gin.Context) {
	var req request.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}

	ret, msg := service.Register(req.Name, req.Email, req.Password, req.Captcha)
	JsonBack(ctx, ret, msg, nil)
}

func SendCaptcha(ctx *gin.Context) {
	var req request.SendCaptchaReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(ctx, constant.BadRequest, constant.Error, nil)
		return
	}

	ret, msg := service.SendCaptcha(req.Email)
	JsonBack(ctx, ret, msg, nil)
}
