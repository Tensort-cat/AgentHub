package middleware

import (
	"AgentHub/internal/controller"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/util"
	"AgentHub/pkg/zlog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 读取jwt
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader
		} else {
			// 兼容 URL 参数传 token
			token = c.Query("token")
		}

		if token == "" {
			controller.JsonBack(c, constant.Unauthorized, "无权限", nil)
			c.Abort()
			return
		}

		zlog.Info("token", zap.String("token", token))
		userID, err := util.ParseJWT(token)
		if err != nil {
			zlog.Info("JWT 解析出错", zap.Error(err))
			controller.JsonBack(c, constant.Unauthorized, "无权限", nil)
			c.Abort()
			return
		}

		// 将用户 id 放入上下文
		// 把字符串转为 int64
		numID, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			zlog.Info("类型转换失败 (字符串转int)", zap.Error(err))
			c.Abort()
			return
		}
		c.Set("user_id", numID)
		c.Next()
	}
}
