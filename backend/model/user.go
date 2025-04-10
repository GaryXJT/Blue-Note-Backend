package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户模型
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"-"` // 不返回密码
	Role         string             `bson:"role" json:"role"` // "user", "admin"
	Status       string             `bson:"status" json:"status"` // "active", "banned", "happy", "relaxed" 等
	Nickname     string             `bson:"nickname" json:"nickname"`
	Avatar       string             `bson:"avatar" json:"avatar"`
	Bio          string             `bson:"bio" json:"bio"`
	Gender       string             `bson:"gender" json:"gender"` // "male", "female", "other"
	Birthday     string             `bson:"birthday" json:"birthday"`
	Location     string             `bson:"location" json:"location"`
	IsAdmin      bool               `bson:"is_admin" json:"is_admin"`
	FollowCount  int                `bson:"follow_count" json:"follow_count"`
	FansCount    int                `bson:"fans_count" json:"fans_count"`
	LikeCount    int                `bson:"like_count" json:"like_count"`
	CollectCount int                `bson:"collect_count" json:"collect_count"`
	PostCount    int                `bson:"post_count" json:"post_count"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updatedAt"`
}

// ProfileResponse 用户资料响应
type ProfileResponse struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Avatar       string `json:"avatar"`
	Bio          string `json:"bio"`
	Gender       string `json:"gender"`
	Birthday     string `json:"birthday"`
	Location     string `json:"location"`
	Status       string `json:"status"`
	FollowCount  int    `json:"followCount"`
	FansCount    int    `json:"fansCount"`
	LikeCount    int    `json:"likeCount"`
	CollectCount int    `json:"collectCount"`
	PostCount    int    `json:"postCount"`
	IsFollowing  bool   `json:"isFollowing"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=1,max=20"`
	Nickname string `json:"nickname" binding:"omitempty,min=1,max=20"`
	Avatar   string `json:"avatar" binding:"omitempty"`
	Bio      string `json:"bio" binding:"omitempty,max=200"`
	Gender   string `json:"gender" binding:"omitempty,oneof=male female other"`
	Birthday string `json:"birthday" binding:"omitempty"`
	Location string `json:"location" binding:"omitempty,max=50"`
	Status   string `json:"status" binding:"omitempty,max=20"`
}

// UserFollow 用户关注关系
type UserFollow struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`           // 关注者ID
	FollowingID    primitive.ObjectID `bson:"following_id" json:"following_id"` // 被关注者ID
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// UserListItem 用户列表项
type UserListItem struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	IsFollowing bool   `json:"isFollowing"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int            `json:"total"`
	List  []UserListItem `json:"list"`
}
