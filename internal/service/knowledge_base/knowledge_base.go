package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"time"

	"go.uber.org/zap"
)

func List(userID int64) (constant.Code, constant.Msg, []response.KbListItem) {
	var kbs []model.KnowledgeBase
	res := dao.DB.Where("user_id = ?", userID).Find(&kbs)
	if res.Error != nil {
		return constant.InternalServerError, constant.Error, nil
	}

	resp := make([]response.KbListItem, len(kbs))
	for i, kb := range kbs {
		resp[i] = response.KbListItem{
			ID:          kb.ID,
			Name:        kb.Name,
			Description: kb.Description,
			CreatedAt:   kb.CreatedAt,
		}
	}

	return constant.Success, constant.Ok, resp
}

func Create(name, description string, userID, embedderID int64) (constant.Code, constant.Msg) {
	kb := model.KnowledgeBase{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		UserID:      userID,
		Name:        name,
		Description: description,
		EmbedderID:  embedderID,
	}

	res := dao.DB.Create(&kb)
	if res.Error != nil {
		zlog.Error("创建知识库失败: ", zap.Error(res.Error))
		return constant.InternalServerError, constant.Error
	}
	return constant.Success, constant.Ok
}

func Detail(kbID string) (constant.Code, constant.Msg, response.KbDetailResp) {
	var kb model.KnowledgeBase
	res := dao.DB.First(&kb, "id = ?", kbID)
	if res.Error != nil {
		return constant.InternalServerError, constant.Error, response.KbDetailResp{}
	}

	// Fetch documents associated with the knowledge base
	var docs []model.Document
	res = dao.DB.Where("knowledge_base_id = ?", kbID).Find(&docs)
	if res.Error != nil {
		return constant.InternalServerError, constant.Error, response.KbDetailResp{}
	}

	// Convert documents to response format
	docMetas := make([]response.DocMeta, len(docs))
	for i, doc := range docs {
		docMetas[i] = response.DocMeta{
			ID:        doc.ID,
			Name:      doc.Name,
			Size:      doc.Size,
			Status:    doc.Status,
			Type:      doc.Type,
			CreatedAt: doc.CreatedAt,
		}
	}

	return constant.Success, constant.Ok, response.KbDetailResp{
		ID:          kb.ID,
		Name:        kb.Name,
		Description: kb.Description,
		EmbedderID:  kb.EmbedderID,
		CreatedAt:   kb.CreatedAt,
		Docs:        docMetas,
	}
}
