package model

import (
	"time"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=1,max=20"`
	Password        string `json:"password" binding:"required,min=1,max=20"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	Nickname        string `json:"nickname" binding:"required,min=1,max=20"`
	Email           string `json:"email" binding:"required,email"`
	CaptchaID       string `json:"captchaId" binding:"required"`
	CaptchaValue    string `json:"captchaCode" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	Nickname  string    `json:"nickname"`
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID    string `json:"captcha_id"`
	CaptchaImage string `json:"captcha_image"`
}
