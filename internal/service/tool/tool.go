package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	tool_enum "AgentHub/pkg/enum/tool"
	"AgentHub/pkg/zlog"
	"errors"
	"time"

	"gorm.io/gorm"
)

func List() (constant.Code, constant.Msg, []response.ToolListResp) {
	var tools []model.Tool
	// 只传没有被禁用的工具
	res := dao.DB.Where("status = ?", tool_enum.Enable).Order("created_at desc").Find(&tools)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("当前没有工具可用")
			return constant.Success, "当前没有工具可用", nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	resp := make([]response.ToolListResp, len(tools))
	for i, tool := range tools {
		resp[i] = response.ToolListResp{
			ID:          tool.ID,
			Name:        tool.Name,
			Description: tool.Description,
		}
	}

	return constant.Success, constant.Ok, resp
}

func Detail(toolID string) (constant.Code, constant.Msg, response.ToolDetailResp) {
	var tool model.Tool
	res := dao.DB.First(&tool, "id = ?", toolID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("工具不存在")
			return constant.BadRequest, "工具不存在", response.ToolDetailResp{}
		}
		return constant.InternalServerError, constant.Error, response.ToolDetailResp{}
	}

	return constant.Success, constant.Ok, response.ToolDetailResp{
		ID:          tool.ID,
		Name:        tool.Name,
		Description: tool.Description,
		Status:      tool.Status,
		Avatar:      tool.Avatar,
		CreateAt:    tool.CreatedAt,
	}
}

func Subscribe(userID, toolID int64, status tool_enum.UserToolStatus) (constant.Code, constant.Msg) {
	// 如果用户是第一次订阅，那就创建新的行，否则只更改状态即可
	var ut model.UserTools
	res := dao.DB.First(&ut, "user_id = ? and tool_id = ?", userID, toolID)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 第一次订阅
			ut = model.UserTools{
				UserID:    userID,
				ToolID:    toolID,
				Status:    status,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := dao.DB.Create(&ut).Error; err != nil {
				zlog.Error(err.Error())
				return constant.InternalServerError, constant.Error
			}
			return constant.Success, constant.Ok
		}

		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	// 曾经订阅过
	ut.Status = status
	dao.DB.Save(&ut)

	return constant.Success, constant.Ok
}

func EnableOrDisable(toolID int64, status tool_enum.ToolStatus) (constant.Code, constant.Msg) {
	res := dao.DB.Model(&model.Tool{}).
		Where("id = ?", toolID).
		Update("status", status)
	if err := res.Error; err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}
