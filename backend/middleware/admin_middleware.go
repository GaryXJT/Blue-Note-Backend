package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AdminMiddleware 检查用户是否具有管理员权限
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限访问",
			})
			c.Abort()
			return
		}
		c.Next()
	}
} 