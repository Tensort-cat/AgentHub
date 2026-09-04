package request

import storage_enum "AgentHub/pkg/enum/storage"

type SessionCreateReq struct {
	WorkflowID  int64                    `json:"workflow_id" binding:"required"`
	Title       string                   `json:"title" binding:"required"`
	StorageMode storage_enum.StorageMode `json:"storage_mode" binding:"required"`
}
