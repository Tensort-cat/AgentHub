package response

import (
	workflow_enum "AgentHub/pkg/enum/workflow"
	"time"
)

type WorkflowPageItem struct {
	ID          int64                        `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Status      workflow_enum.WorkflowStatus `json:"status"`
	CreatedAt   time.Time                    `json:"created_at"`
}
