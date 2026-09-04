package model

import (
	"time"

	"gorm.io/gorm"
)

// 所有实体直接嵌入该结构体

type BaseModel struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}
