package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Page(wfID string, page, size int) (constant.Code, constant.Msg, []response.SessionPageItem) {
	if page < 1 || size < 1 {
		return constant.BadRequest, "page and size must be greater than 0", nil
	}

	var sessions []model.Session
	res := dao.DB.Where("workflow_id = ?", wfID).
		Order("created_at desc").
		Limit(size).
		Offset((page - 1) * size).
		Find(&sessions)

	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constant.Success, constant.Ok, nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	if len(sessions) == 0 {
		zlog.Debug("没有数据")
		return constant.Success, constant.Ok, nil
	}

	resp := make([]response.SessionPageItem, len(sessions))
	for i, session := range sessions {
		resp[i] = response.SessionPageItem{
			ID:        session.ID,
			Title:     session.Title,
			CreatedAt: session.CreatedAt,
		}
	}

	return constant.Success, constant.Ok, resp
}

func Create(wfID int64, title string) error {
	session := model.Session{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		WorkflowID: wfID,
		Title:      title,
	}

	res := dao.DB.Create(&session)
	if err := res.Error; err != nil {
		zlog.Error("创建会话失败: ", zap.Error(err))
		return err
	}
	return nil
}

func MsgPage(sessionID string, page, size int) (constant.Code, constant.Msg, []response.SessionMsgPageItem) {
	if page < 1 || size < 1 {
		return constant.BadRequest, "page and size must be greater than 0", nil
	}

	var messages []model.Message
	res := dao.DB.Where("session_id = ?", sessionID).
		Order("created_at desc").
		Limit(size).
		Offset((page - 1) * size).
		Find(&messages)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constant.Success, constant.Ok, nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	if len(messages) == 0 {
		zlog.Debug("没有数据")
		return constant.Success, constant.Ok, nil
	}

	resp := make([]response.SessionMsgPageItem, len(messages))
	for i, msg := range messages {
		resp[i] = response.SessionMsgPageItem{
			ID:        msg.ID,
			Content:   msg.Content,
			Type:      msg.Type,
			CreatedAt: msg.CreatedAt,
		}
	}

	return constant.Success, constant.Ok, resp
}
