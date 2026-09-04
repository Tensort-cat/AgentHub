package request

type RunReq struct {
	WfID  int64  `json:"id"`
	Input string `json:"input"`
}
