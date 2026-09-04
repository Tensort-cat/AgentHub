package request

type WorkflowCreateReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
