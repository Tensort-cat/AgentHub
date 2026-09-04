package response

import document_enum "AgentHub/pkg/enum/document"

type DocUploadResp struct {
	ID     int64                   `json:"id"`
	Name   string                  `json:"name"`
	Type   document_enum.DocType   `json:"type"`
	Size   int64                   `json:"size"`
	Status document_enum.DocStatus `json:"status"`
}
