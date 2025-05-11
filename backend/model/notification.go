package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem  = "system"   // 系统通知
	NotificationTypeLike    = "like"     // 点赞通知
	NotificationTypeComment = "comment"  // 评论通知
	NotificationTypeFollow  = "follow"   // 关注通知
)

// Notification 通知模型
type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`           // 接收通知的用户ID
	Type        NotificationType   `bson:"type" json:"type"`               // 通知类型
	Title       string            `bson:"title" json:"title"`             // 通知标题
	Content     string            `bson:"content" json:"content"`         // 通知内容
	IsRead      bool              `bson:"is_read" json:"isRead"`         // 是否已读
	CreatedAt   time.Time         `bson:"created_at" json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time         `bson:"updated_at" json:"updatedAt"`   // 更新时间
	
	// 关联数据
	RelatedID   primitive.ObjectID `bson:"related_id,omitempty" json:"relatedId,omitempty"`   // 关联ID（如帖子ID、评论ID等）
	RelatedType string            `bson:"related_type,omitempty" json:"relatedType,omitempty"` // 关联类型（post/comment等）
	
	// 发送者信息（用于点赞和关注通知）
	SenderID    primitive.ObjectID `bson:"sender_id,omitempty" json:"senderId,omitempty"`     // 发送者ID
	SenderName  string            `bson:"sender_name,omitempty" json:"senderName,omitempty"` // 发送者名称
	SenderAvatar string           `bson:"sender_avatar,omitempty" json:"senderAvatar,omitempty"` // 发送者头像
}

// NotificationQuery 通知查询参数
type NotificationQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=like follow system"` // 通知类型过滤
	IsRead   *bool  `form:"isRead" binding:"omitempty"`                       // 是否已读过滤
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	Total int            `json:"total"`
	List  []Notification `json:"list"`
}

// CreateNotificationRequest 创建通知请求
type CreateNotificationRequest struct {
	Type        string `json:"type" binding:"required,oneof=system like comment follow"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
	RelatedID   string `json:"relatedId" binding:"required"`
	RelatedType string `json:"relatedType" binding:"required,oneof=post comment user"`
	SenderID    string `json:"senderId,omitempty"`
	SenderName  string `json:"senderName,omitempty"`
	SenderAvatar string `json:"senderAvatar,omitempty"`
}

// UpdateNotificationRequest 更新通知请求
type UpdateNotificationRequest struct {
	IsRead bool `json:"isRead" binding:"required"`
} 