package service

import (
	"blue-note/middleware"
	"blue-note/model"
	"blue-note/util"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db             *mongo.Database
	profileService *ProfileService
}

func NewAuthService(db *mongo.Database, profileService *ProfileService) *AuthService {
	return &AuthService{
		db:             db,
		profileService: profileService,
	}
}

func (s *AuthService) GenerateCaptcha() (string, string, error) {
	captcha, err := util.GenerateCaptcha()
	if err != nil {
		return "", "", err
	}
	return captcha.ID, captcha.Base64, nil
}

// LoginOrRegister 处理登录或注册逻辑，返回用户信息、token和是否为新用户
func (s *AuthService) LoginOrRegister(req *model.LoginRequest) (*model.User, string, time.Time, bool, error) {
	// 验证验证码
	if req.CaptchaID != "" && req.CaptchaCode != "" {
		if !util.VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
			return nil, "", time.Time{}, false, errors.New("验证码错误")
		}
	}

	// 查找用户
	var user model.User
	err := s.db.Collection("users").FindOne(context.Background(), bson.M{"username": req.Username}).Decode(&user)
	
	// 如果用户不存在，则创建新用户（注册）
	isNewUser := false
	if err == mongo.ErrNoDocuments {
		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", time.Time{}, false, err
		}
		
		// 创建新用户
		newUser := &model.User{
			Username:  req.Username,
			Password:  string(hashedPassword),
			Nickname:  req.Username,
			Status:    "active",
			IsAdmin:   false,
			Avatar:    s.profileService.GetDefaultAvatarURL(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		result, err := s.db.Collection("users").InsertOne(context.Background(), newUser)
		if err != nil {
			return nil, "", time.Time{}, false, err
		}
		
		newUser.ID = result.InsertedID.(primitive.ObjectID)

		// 生成JWT token
		token, err := middleware.GenerateToken(newUser)
		if err != nil {
			return nil, "", time.Time{}, false, err
		}
		
		// 计算过期时间
		expireHours := viper.GetInt("jwt.expire")
		expiresAt := time.Now().Add(time.Hour * time.Duration(expireHours))
		
		isNewUser = true
		return newUser, token, expiresAt, isNewUser, nil
	} else if err != nil {
		return nil, "", time.Time{}, false, errors.New("用户查询失败")
	}
	
	// 用户存在，验证密码（登录）
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, "", time.Time{}, false, errors.New("用户名或密码错误")
	}
	
	// 验证用户ID格式
	_, err = primitive.ObjectIDFromHex(user.ID.Hex())
	if err != nil {
		return nil, "", time.Time{}, false, fmt.Errorf("用户ID格式无效: %w", err)
	}
	
	// 生成JWT token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		return nil, "", time.Time{}, false, err
	}
	
	// 计算过期时间
	expireHours := viper.GetInt("jwt.expire")
	expiresAt := time.Now().Add(time.Hour * time.Duration(expireHours))
	
	return &user, token, expiresAt, false, nil
}
