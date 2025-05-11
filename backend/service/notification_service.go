package service

import (
	"blue-note/model"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationService struct {
	db *mongo.Database
}

func NewNotificationService(db *mongo.Database) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification 创建通知
func (s *NotificationService) CreateNotification(userID string, req *model.CreateNotificationRequest) (*model.Notification, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	notification := &model.Notification{
		UserID:    userObjID,
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		IsRead:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置关联数据
	if req.RelatedID != "" {
		relatedObjID, err := primitive.ObjectIDFromHex(req.RelatedID)
		if err != nil {
			return nil, fmt.Errorf("无效的关联ID: %w", err)
		}
		notification.RelatedID = relatedObjID
		notification.RelatedType = req.RelatedType
	}

	// 设置发送者信息
	if req.SenderID != "" {
		senderObjID, err := primitive.ObjectIDFromHex(req.SenderID)
		if err != nil {
			return nil, fmt.Errorf("无效的发送者ID: %w", err)
		}
		notification.SenderID = senderObjID
		notification.SenderName = req.SenderName
		notification.SenderAvatar = req.SenderAvatar
	}

	result, err := s.db.Collection("notifications").InsertOne(context.Background(), notification)
	if err != nil {
		return nil, fmt.Errorf("创建通知失败: %w", err)
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)
	return notification, nil
}

// GetNotificationList 获取通知列表
func (s *NotificationService) GetNotificationList(userID string, query *model.NotificationQuery) (*model.NotificationListResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	// 设置默认值
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}
	skip := (query.Page - 1) * query.PageSize

	// 构建查询条件
	filter := bson.M{"user_id": userObjID}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.IsRead != nil {
		filter["is_read"] = *query.IsRead
	}

	// 获取总数
	total, err := s.db.Collection("notifications").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("获取通知总数失败: %w", err)
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(query.PageSize))

	cursor, err := s.db.Collection("notifications").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询通知列表失败: %w", err)
	}
	defer cursor.Close(context.Background())

	var notifications []model.Notification
	if err = cursor.All(context.Background(), &notifications); err != nil {
		return nil, fmt.Errorf("解析通知列表失败: %w", err)
	}

	return &model.NotificationListResponse{
		Total: int(total),
		List:  notifications,
	}, nil
}

// UpdateNotification 更新通知
func (s *NotificationService) UpdateNotification(notificationID string, userID string, req *model.UpdateNotificationRequest) error {
	notificationObjID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("无效的通知ID: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}

	// 更新通知
	result, err := s.db.Collection("notifications").UpdateOne(
		context.Background(),
		bson.M{
			"_id":     notificationObjID,
			"user_id": userObjID,
		},
		bson.M{
			"$set": bson.M{
				"is_read":    req.IsRead,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("更新通知失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("通知不存在或不属于当前用户")
	}

	return nil
}

// DeleteNotification 删除通知
func (s *NotificationService) DeleteNotification(notificationID string, userID string) error {
	notificationObjID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("无效的通知ID: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}

	result, err := s.db.Collection("notifications").DeleteOne(
		context.Background(),
		bson.M{
			"_id":     notificationObjID,
			"user_id": userObjID,
		},
	)
	if err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("通知不存在或不属于当前用户")
	}

	return nil
}

// MarkAllAsRead 将所有通知标记为已读
func (s *NotificationService) MarkAllAsRead(userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}

	_, err = s.db.Collection("notifications").UpdateMany(
		context.Background(),
		bson.M{
			"user_id": userObjID,
			"is_read": false,
		},
		bson.M{
			"$set": bson.M{
				"is_read":    true,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("标记所有通知为已读失败: %w", err)
	}

	return nil
}

// GetUnreadCount 获取未读通知数量
func (s *NotificationService) GetUnreadCount(userID string) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, fmt.Errorf("无效的用户ID: %w", err)
	}

	count, err := s.db.Collection("notifications").CountDocuments(
		context.Background(),
		bson.M{
			"user_id": userObjID,
			"is_read": false,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("获取未读通知数量失败: %w", err)
	}

	return count, nil
} 