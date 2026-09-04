package model

import document_enum "AgentHub/pkg/enum/document"

type Document struct {
	BaseModel

	KnowledgeBaseID int64                   `gorm:"column:knowledge_base_id;index;not null"`
	Name            string                  `gorm:"column:name;type:varchar(255);not null"`
	Type            document_enum.DocType   `gorm:"column:status;default:0"`
	Status          document_enum.DocStatus `gorm:"column:status;default:0"`
	Size            int64                   `gorm:"column:size;not null"`
	ChunkNum        int                     `gorm:"column:chunk_num;default:0"`
}

func (Document) TableName() string {
	return "documents"
}
