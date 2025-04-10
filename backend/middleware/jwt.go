package middleware

import (
	"blue-note/config"
	"blue-note/model"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	jwt.RegisteredClaims
}

func GenerateToken(user *model.User) (string, error) {
	// 确保用户ID是有效的ObjectID
	userIDStr := user.ID.Hex()
	
	// 验证ID格式
	_, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return "", fmt.Errorf("无效的用户ID格式: %v", err)
	}
	
	expireHours := config.GetConfig().JWT.Expire
	
	claims := &Claims{
		UserID:   userIDStr,
		Username: user.Username,
		Role:     user.Role,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetConfig().JWT.Secret))
}

// AuthMiddleware 验证用户是否登录
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证信息",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证格式错误",
			})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetConfig().JWT.Secret), nil
		})

		// 检查错误类型，区分 token 过期和其他无效情况
		if err != nil {
			// 检查错误消息中是否包含"expired"关键词
			errMsg := err.Error()
			if strings.Contains(strings.ToLower(errMsg), "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "认证信息已过期，请重新登录",
				})
				c.Abort()
				return
			}
			
			// 其他类型的错误
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证信息",
			})
			c.Abort()
			return
		}

		// token 解析成功但可能无效
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证信息",
			})
			c.Abort()
			return
		}

		// 从 MapClaims 中提取字段
		userID, ok := claims["userId"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的用户ID格式",
			})
			c.Abort()
			return
		}
		
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)
		nickname, _ := claims["nickname"].(string)
		avatar, _ := claims["avatar"].(string)
		
		// 验证ObjectID格式
		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			fmt.Printf("警告: 用户ID格式无效 (%s): %v\n", userID, err)
		}
		
		// 将用户信息设置到上下文
		c.Set("userId", userID)
		c.Set("username", username)
		c.Set("role", role)
		c.Set("nickname", nickname)
		c.Set("avatar", avatar)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		for _, role := range roles {
			if role == userRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无权限访问",
		})
		c.Abort()
	}
}
