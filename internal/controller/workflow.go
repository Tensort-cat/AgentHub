package controller

import (
	"AgentHub/internal/dto/request"
	service "AgentHub/internal/service/workflow"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func WorkflowPage(c *gin.Context) {
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
	}

	pageStr, ok := c.GetQuery("page")
	if !ok {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	sizeStr, ok := c.GetQuery("size")
	if !ok {
		zlog.Info("参数缺失")
		JsonBack(c, constant.BadRequest, "参数缺失", nil)
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		zlog.Error("字符串转整型错误")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	ret, msg, data := service.Page(id, page, size)
	JsonBack(c, ret, msg, data)
}

func WorkflowCreate(c *gin.Context) {
	var req request.WorkflowCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
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
	}

	ret, msg := service.Create(id, req.Name, req.Description)
	JsonBack(c, ret, msg, nil)
}

func WorkflowDetail(c *gin.Context) {
	wfID := c.Param("id")

	ret, msg, data := service.Detail(wfID)
	JsonBack(c, ret, msg, data)
}

func WordflowUpdate(c *gin.Context) {
	var req request.WorkflowUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	wfID := c.Param("id")
	ret, msg := service.Update(wfID, req.Name, req.Description, req.Status)
	JsonBack(c, ret, msg, nil)
}

func WorkflowDelete(c *gin.Context) {
	wfID := c.Param("id")
	ret, msg := service.Delete(wfID)
	JsonBack(c, ret, msg, nil)
}

func Run(c *gin.Context) {
	var req request.RunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
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

	ret, msg, data := service.SubmitRun(c.Request.Context(), req.WfID, req.Input, id)
	JsonBack(c, ret, msg, data)
}

func WorkflowRunEvents(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}
	id, ok := userID.(int64)
	if !ok {
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	taskID := c.Param("task_id")
	ret, msg, state := service.GetRunState(c.Request.Context(), taskID, id)
	if ret != constant.Success {
		JsonBack(c, ret, msg, nil)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	sendState := func() {
		c.SSEvent("workflow_run", state)
		c.Writer.Flush()
	}
	sendState()
	if state.Status.IsTerminal() {
		return
	}

	pollTicker := time.NewTicker(time.Second)
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer pollTicker.Stop()
	defer heartbeatTicker.Stop()

	lastUpdatedAt := state.UpdatedAt
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeatTicker.C:
			_, _ = c.Writer.WriteString(": ping\n\n")
			c.Writer.Flush()
		case <-pollTicker.C:
			ret, msg, state = service.GetRunState(c.Request.Context(), taskID, id)
			if ret != constant.Success {
				c.SSEvent("error", gin.H{"code": ret, "msg": msg})
				c.Writer.Flush()
				return
			}
			if !state.UpdatedAt.Equal(lastUpdatedAt) {
				sendState()
				lastUpdatedAt = state.UpdatedAt
			}
			if state.Status.IsTerminal() {
				return
			}
		}
	}
}

func NodeCreate(c *gin.Context) {
	var req request.NodeConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	wfIDStr := c.Param("id")
	wfID, err := strconv.Atoi(wfIDStr)
	if err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}
	ret, msg := service.NodeCreate(int64(wfID), req.Name, req.Type, req.PositionX, req.PositionY, req.Config)
	JsonBack(c, ret, msg, nil)
}

func NodeUpdate(c *gin.Context) {
	var req request.NodeConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	nodeID := c.Param("node_id")
	ret, msg := service.NodeUpdate(nodeID, req.Name, req.Type, req.PositionX, req.PositionY, req.Config)
	JsonBack(c, ret, msg, nil)
}

func NodeDelete(c *gin.Context) {
	nodeID := c.Param("node_id")
	ret, msg := service.NodeDelete(nodeID)
	JsonBack(c, ret, msg, nil)
}

func EdgeCreate(c *gin.Context) {
	var req request.EdgeConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.BadRequest, constant.Error, nil)
		return
	}

	wfIDStr := c.Param("id")
	wfID, err := strconv.Atoi(wfIDStr)
	if err != nil {
		zlog.Error(err.Error())
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}
	ret, msg := service.EdgeCreate(int64(wfID), req.SourceNodeID, req.TargetNodeID, req.Config)
	JsonBack(c, ret, msg, nil)
}

func EdgeDelete(c *gin.Context) {
	edgeID := c.Param("edge_id")
	ret, msg := service.EdgeDelete(edgeID)
	JsonBack(c, ret, msg, nil)
}
