package request

import document_enum "AgentHub/pkg/enum/document"

type DocUploadReq struct {
	KnowledgeBaseID string                `json:"knowledge_base_id"`
	Name            string                `json:"name"`
	Type            document_enum.DocType `json:"type"`
	Size            int64                 `json:"size"`
}
