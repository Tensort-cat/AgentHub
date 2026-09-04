package response

type LoginResp struct {
	Token  string `json:"token,omitempty"`
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}
