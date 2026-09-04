package response

type WorkflowResult struct {
	SessionID int64  `json:"session_id"`
	Content   string `json:"content"`
}
