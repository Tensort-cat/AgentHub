package request

type KbCreateReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	EmbedderID  int64  `json:"embedder_id" binding:"required"`
}
