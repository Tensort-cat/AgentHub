package response

import (
	document_enum "AgentHub/pkg/enum/document"
	"time"
)

type KbDetailResp struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	EmbedderID  int64     `json:"embedder_id"`
	CreatedAt   time.Time `json:"created_at"`
	Docs        []DocMeta `json:"docs"`
}

type DocMeta struct {
	ID        int64                   `json:"id"`
	Name      string                  `json:"name"`
	Size      int64                   `json:"size"`
	Status    document_enum.DocStatus `json:"status"`
	Type      document_enum.DocType   `json:"type"`
	CreatedAt time.Time               `json:"created_at"`
}
