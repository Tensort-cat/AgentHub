package service

import (
	"AgentHub/internal/dao"
	"AgentHub/internal/dto/response"
	"AgentHub/internal/model"
	"AgentHub/pkg/constant"
	model_enum "AgentHub/pkg/enum/model"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 列出当前用户已配置的模型
func List(userID int64) (constant.Code, constant.Msg, []response.ModelListResp) {
	var models []model.Model
	res := dao.DB.Where("user_id = ?", userID).Order("created_at asc").Find(&models)
	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 没有数据
			return constant.Success, constant.Ok, nil
		}
		return constant.InternalServerError, constant.Error, nil
	}

	resp := make([]response.ModelListResp, len(models))
	for i, model := range models {
		resp[i] = response.ModelListResp{
			ID:        model.ID,
			Name:      model.Name,
			Provider:  model.Provider,
			BaseUrl:   model.BaseURL,
			Type:      model.Type,
			CreatedAt: model.CreatedAt,
		}
	}
	return constant.Success, constant.Ok, resp
}

func isValidModelType(t model_enum.ModelType) bool {
	switch t {
	case model_enum.CHAT, model_enum.EMBEDDING:
		return true
	default:
		return false
	}
}

// 创建新模型
func Create(name, baseUrl, apiKey string, userID int64, modelType model_enum.ModelType) (constant.Code, constant.Msg) {
	// 不允许有参数为空或不合法
	if strings.TrimSpace(name) == "" ||
		strings.TrimSpace(baseUrl) == "" ||
		strings.TrimSpace(apiKey) == "" ||
		!isValidModelType(modelType) {
		zlog.Info("存在空参数或参数不合法")
		return constant.BadRequest, "存在空参数或参数不合法"
	}

	model := model.Model{
		BaseModel: model.BaseModel{
			ID:        util.GenID(),
			CreatedAt: time.Now(),
		},
		UserID:  userID,
		Name:    name,
		BaseURL: baseUrl,
		APIKey:  apiKey,
		Type:    modelType,
	}

	res := dao.DB.Create(&model)
	if err := res.Error; err != nil {
		zlog.Error(err.Error())
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

// 更新模型配置 (不允许修改模型类型)
func Update(modelID string, name, baseUrl, apiKey string) (constant.Code, constant.Msg) {
	updateData := make(map[string]any)
	fields := map[string]string{
		"name":     name,
		"base_url": baseUrl,
		"api_key":  apiKey,
	}
	for k, v := range fields {
		if v != "" {
			updateData[k] = v
		}
	}

	res := dao.DB.Model(&model.Model{}).Where("id = ?", modelID).Updates(updateData)
	if err := res.Error; err != nil {
		zlog.Error("数据库更新出错")
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}

// 删除模型配置
func Delete(modelID string) (constant.Code, constant.Msg) {
	if strings.TrimSpace(modelID) == "" {
		zlog.Info("参数不合法: modelID 为空")
		return constant.BadRequest, "modelID 不能为空"
	}

	deletedAt := gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.Model{}).
			Where("id = ?", modelID).
			Where("deleted_at IS NULL").
			Update("deleted_at", deletedAt)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zlog.Info("模型不存在")
			return constant.BadRequest, "模型不存在"
		}
		zlog.Error("删除模型出错")
		return constant.InternalServerError, constant.Error
	}

	return constant.Success, constant.Ok
}
