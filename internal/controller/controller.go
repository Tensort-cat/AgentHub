package controller

import (
	"AgentHub/pkg/constant"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JsonBack(ctx *gin.Context, ret constant.Code, msg constant.Msg, data any) {
	var httpStatus int
	switch ret {
	case constant.Success: // 业务正常走完流程返回的结果码
		httpStatus = http.StatusOK

	case constant.InternalServerError: // 系统错误导致未正常走完业务流程返回的结果码
		httpStatus = http.StatusInternalServerError

	case constant.BadRequest: // 业务数据问题导致未正常走完业务流程返回的结果码
		httpStatus = http.StatusBadRequest

	case constant.Unauthorized: // 权限不足
		httpStatus = http.StatusUnauthorized
	}

	ctx.JSON(httpStatus, gin.H{
		"code": ret,
		"msg":  msg,
		"data": data,
	})
}
