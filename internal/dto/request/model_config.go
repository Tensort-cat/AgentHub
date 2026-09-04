package request

import model_enum "AgentHub/pkg/enum/model"

type ModelConfigReq struct {
	Name    string               `json:"name"`
	BaseUrl string               `json:"base_url"`
	ApiKey  string               `json:"api_key"`
	Type    model_enum.ModelType `json:"type"`
}
