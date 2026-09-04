package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	"errors"

	"gorm.io/gorm"
)

func Me(id int64) (constant.Code, constant.Msg, *response.MeResp) {
	var user model.User
	if err := dao.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constant.BadRequest, "id不存在", nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	resp := &response.MeResp{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Avatar: user.Avatar,
	}
	return constant.Success, constant.Ok, resp
}
