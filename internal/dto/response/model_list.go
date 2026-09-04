package response

import (
	model_enum "AgentHub/pkg/enum/model"
	"time"
)

type ModelListResp struct {
	ID        int64                    `json:"id"`
	Name      string                   `json:"name"`
	Provider  model_enum.ModelProvider `json:"provider"`
	BaseUrl   string                   `json:"base_url"`
	Type      model_enum.ModelType     `json:"type"`
	CreatedAt time.Time                `json:"created_at"`
}
