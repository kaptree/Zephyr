package middleware

import (
	"strings"

	"labelpro-server/internal/config"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.Features.DemoMode {
			c.Set("user_id", "demo-user-id")
			c.Set("username", "demo")
			c.Set("role", "super_admin")
			c.Set("department_id", "")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("department_id", claims.DepartmentID)
		c.Set("access_token", token)
		c.Set("token_id", claims.ID)

		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		if userRole == "" {
			utils.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		utils.Forbidden(c, "权限不足")
		c.Abort()
	}
}

func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

func GetUserRole(c *gin.Context) string {
	return c.GetString("role")
}

func GetUserDeptID(c *gin.Context) string {
	return c.GetString("department_id")
}
