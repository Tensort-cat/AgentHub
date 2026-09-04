package model

import model_enum "AgentHub/pkg/enum/model"

type Model struct {
	BaseModel

	UserID   int64                    `gorm:"column:user_id;index;not null"`
	Name     string                   `gorm:"column:name;type:varchar(100);not null"`
	Provider model_enum.ModelProvider `gorm:"column:provider;type:tinyint;not null"`
	BaseURL  string                   `gorm:"column:base_url;type:varchar(255);not null"`
	APIKey   string                   `gorm:"column:api_key;type:varchar(255);not null"`
	Type     model_enum.ModelType     `gorm:"column:type;not null"`
}

func (Model) TableName() string {
	return "models"
}
