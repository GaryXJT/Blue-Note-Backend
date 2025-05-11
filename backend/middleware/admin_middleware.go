package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware 检查用户是否具有管理员权限
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 检查token格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证格式",
			})
			c.Abort()
			return
		}

		// 从上下文中获取用户信息
		userID := c.GetString("userId")
		username := c.GetString("username")
		role := c.GetString("role")
		status := c.GetString("status")

		// 打印用户信息
		fmt.Printf("\n=== 管理员中间件信息 ===\n")
		fmt.Printf("用户ID: %s\n", userID)
		fmt.Printf("用户名: %s\n", username)
		fmt.Printf("用户角色: %s\n", role)
		fmt.Printf("用户状态: %s\n", status)
		fmt.Printf("请求路径: %s\n", c.Request.URL.Path)
		fmt.Printf("请求方法: %s\n", c.Request.Method)
		fmt.Printf("=====================\n\n")

		// 检查是否是管理员
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
} 