package model

type KnowledgeBase struct {
	BaseModel

	UserID      int64  `gorm:"column:user_id;index;not null"`
	Name        string `gorm:"column:name;type:varchar(100);not null"`
	Description string `gorm:"column:description;type:text"`
	EmbedderID  int64  `gorm:"column:embedder_id;index;not null"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_base"
}
