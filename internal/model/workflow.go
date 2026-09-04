package model

import (
	workflow_enum "AgentHub/pkg/enum/workflow"

	"gorm.io/datatypes"
)

type Workflow struct {
	BaseModel

	UserID      int64                        `gorm:"column:user_id;index;not null"`
	Name        string                       `gorm:"column:name;type:varchar(100);not null"`
	Description string                       `gorm:"column:description;type:text"`
	Status      workflow_enum.WorkflowStatus `gorm:"column:status"`
}

func (Workflow) TableName() string {
	return "workflows"
}

type WorkflowNode struct {
	BaseModel

	WorkflowID int64                          `gorm:"column:workflow_id;index;not null"`
	Name       string                         `gorm:"column:name;type:varchar(100);not null"`
	Type       workflow_enum.WorkflowNodeType `gorm:"column:type;not null"`
	PositionX  int                            `gorm:"column:position_x;not null"`
	PositionY  int                            `gorm:"column:position_y;not null"`
	Config     datatypes.JSON                 `gorm:"column:config;type:json"`
}

func (WorkflowNode) TableName() string {
	return "workflow_nodes"
}

type WorkflowEdge struct {
	BaseModel

	WorkflowID   int64 `gorm:"column:workflow_id;index;not null"`
	SourceNodeID int64 `gorm:"column:source_node_id;index;not null"`
	TargetNodeID int64 `gorm:"column:target_node_id;index;not null"`
}

func (WorkflowEdge) TableName() string {
	return "workflow_edges"
}
