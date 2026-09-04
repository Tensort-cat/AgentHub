package model

type Session struct {
	BaseModel

	WorkflowID int64  `gorm:"column:workflow_id;index;not null"`
	Title      string `gorm:"column:title;type:varchar(100)"`
}

func (Session) TableName() string {
	return "sessions"
}
