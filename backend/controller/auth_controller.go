package controller

import (
	"blue-note/model"
	"blue-note/service"
	"blue-note/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) GetCaptcha(ctx *gin.Context) {
	captchaID, captchaImage, err := c.authService.GenerateCaptcha()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成验证码失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"captcha_id":    captchaID,
		"captcha_image": captchaImage,
	})
}

// Login 用户登录
func (c *AuthController) Login(ctx *gin.Context) {
	var req model.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	user, token, expiresAt, isNewUser, err := c.authService.LoginOrRegister(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    40101,
			"message": "用户名或密码错误",
			"error":   err.Error(),
		})
		return
	}

	// 根据是否为新用户确定状态码和消息
	statusCode := http.StatusOK
	responseCode := 200
	message := "登录成功"

	if isNewUser {
		statusCode = http.StatusCreated
		responseCode = 201
		message = "注册成功"
	}

	// 返回用户信息和token，添加头像、昵称和过期时间
	ctx.JSON(statusCode, gin.H{
		"code":    responseCode,
		"message": message,
		"data": gin.H{
			"token":     token,
			"expiresAt": expiresAt,
			"user_id":   user.ID.Hex(),
			"username":  user.Username,
			"role":      user.Role,
			"avatar":    user.Avatar,
			"nickname":  user.Nickname,
		},
	})
}

// getRole 根据用户信息确定角色
func getRole(user *model.User) string {
	if user.IsAdmin {
		return "admin"
	}
	return "user"
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
		CaptchaID   string `json:"captchaId" binding:"required"`
		CaptchaCode string `json:"captchaCode" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证验证码
	if !util.VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "验证码错误",
		})
		return
	}

	// 调用服务层修改密码
	err := c.authService.ChangePassword(req.Username, req.OldPassword, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功",
		"data":    nil,
	})
}
