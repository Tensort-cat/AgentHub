package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	message_enum "AgentHub/pkg/enum/message"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"time"

	"go.uber.org/zap"
)

func Create(sessionID int64, content string, msgType message_enum.MessageType) (constant.Code, constant.Msg) {
	msg := model.Message{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		SessionID: sessionID,
		Content:   content,
		Type:      msgType,
	}

	res := dao.DB.Create(&msg)
	if err := res.Error; err != nil {
		zlog.Error("创建消息失败: ", zap.Error(err))
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}
