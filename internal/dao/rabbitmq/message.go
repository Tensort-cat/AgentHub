package rabbitmq

import "time"

const (
	TaskTypeParse = "document.parse"
	TaskTypeIndex = "document.index"
)

type DocumentMessage struct {
	TaskID string `json:"task_id"`
	DocID  int64  `json:"doc_id"`
	KbID   int64  `json:"kb_id"`

	// 文件已经在 HTTP 请求阶段保存到服务器
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`

	// 用于简单的幂等/重试控制
	RetryCount int `json:"retry_count"`

	CreatedAt time.Time `json:"created_at"`
}
