package route

import (
	"AgentHub/internal/config"

	"github.com/gin-gonic/gin"
)

func InitStatic(api *gin.RouterGroup) {
	cfg := config.Cfg.StaticSrcConfig
	api.Static("/avatars", cfg.StaticAvatarPath)
}
