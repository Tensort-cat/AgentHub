package controller

import (
	service "AgentHub/internal/service/session"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SessionPage(c *gin.Context) {
	wfID, exists := c.GetQuery("wf_id")
	if !exists || wfID == "" {
		zlog.Error("缺少 wf_id 参数")
		JsonBack(c, constant.BadRequest, "需要工作流 ID !", nil)
		return
	}

	pageStr, exists := c.GetQuery("page")
	if !exists || pageStr == "" {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	sizeStr, exists := c.GetQuery("size")
	if !exists || sizeStr == "" {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	ret, msg, data := service.Page(wfID, page, size)
	JsonBack(c, ret, msg, data)
}

func SessionMsgPage(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		zlog.Error("缺少会话 ID 参数")
		JsonBack(c, constant.BadRequest, "缺少会话 ID 参数", nil)
		return
	}

	pageStr, exists := c.GetQuery("page")
	if !exists || pageStr == "" {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	sizeStr, exists := c.GetQuery("size")
	if !exists || sizeStr == "" {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	ret, msg, data := service.MsgPage(sessionID, page, size)
	JsonBack(c, ret, msg, data)
}
