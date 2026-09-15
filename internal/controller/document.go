package controller

import (
	"AgentHub/internal/config"
	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	service "AgentHub/internal/service/document"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

var basePath = config.Cfg.StaticSrcConfig.StaticFilePath

func DocUpload(c *gin.Context) {
	kbIDStr := c.PostForm("knowledge_base_id")
	kbID, err := strconv.ParseInt(kbIDStr, 10, 64)
	if err != nil {
		zlog.Warn(err.Error())
		JsonBack(c, constant.BadRequest, "缺少参数信息或格式错误", nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		zlog.Warn(err.Error())
		JsonBack(c, constant.BadRequest, "缺少参数信息", nil)
		return
	}

	ret, msg, data := service.DocUpload(kbID, file)
	JsonBack(c, ret, msg, data)
}

func DocDownload(c *gin.Context) {
	docID := c.Param("docID")

	// 1. 获取当前用户 ID
	userIDAny, exists := c.Get("user_id")
	if !exists {
		zlog.Error("不存在参数 user_id")
		JsonBack(c, constant.BadRequest, constant.Error, nil)
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		zlog.Error("断言失败")
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
	}

	// 2. 查询文档
	var doc model.Document
	if err := dao.DB.First(&doc, "id = ?", docID).Error; err != nil {
		// 返回文档不存在
		zlog.Debug(err.Error())
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	// 3. 查询知识库，并验证归属
	var kb model.KnowledgeBase
	if err := dao.DB.
		Where("id = ? AND user_id = ?", doc.KnowledgeBaseID, userID).
		First(&kb).Error; err != nil {
		// 没有权限
		zlog.Debug(err.Error())
		JsonBack(c, constant.InternalServerError, constant.Error, nil)
		return
	}

	// 4. 构造文件路径
	filePath := filepath.Join(
		basePath,
		strconv.FormatInt(doc.KnowledgeBaseID, 10),
		strconv.FormatInt(doc.ID, 10),
		doc.Name,
	)

	// 5. 下载
	c.FileAttachment(filePath, doc.Name)
}
