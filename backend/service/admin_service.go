package service

import (
	"blue-note/model"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdminService struct {
	db *mongo.Database
}

func NewAdminService(db *mongo.Database) *AdminService {
	return &AdminService{db: db}
}

type StatisticsResponse struct {
	TotalUsers    int            `json:"totalUsers"`
	TotalPosts    int            `json:"totalPosts"`
	PendingPosts  int            `json:"pendingPosts"`
	TotalComments int            `json:"totalComments"`
	DailyStats    []*DailyStat   `json:"dailyStats"`
	TagStats      []*TagStat     `json:"tagStats"`
}

type DailyStat struct {
	Date        string `json:"date"`
	NewUsers    int    `json:"newUsers"`
	NewPosts    int    `json:"newPosts"`
	NewComments int    `json:"newComments"`
}

type TagStat struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func (s *AdminService) GetStatistics() (*StatisticsResponse, error) {
	stats := &StatisticsResponse{}

	// 获取总用户数
	totalUsers, err := s.db.Collection("users").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = int(totalUsers)

	// 获取总帖子数
	totalPosts, err := s.db.Collection("posts").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalPosts = int(totalPosts)

	// 获取待审核帖子数
	pendingPosts, err := s.db.Collection("posts").CountDocuments(context.Background(), bson.M{"status": "pending"})
	if err != nil {
		return nil, err
	}
	stats.PendingPosts = int(pendingPosts)

	// 获取总评论数
	totalComments, err := s.db.Collection("comments").CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalComments = int(totalComments)

	// 获取最近7天的每日统计
	stats.DailyStats, err = s.getDailyStats()
	if err != nil {
		return nil, err
	}

	// 获取标签统计
	stats.TagStats, err = s.getTagStats()
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *AdminService) getDailyStats() ([]*DailyStat, error) {
	// 计算最近7天的日期
	now := time.Now()
	var results []*DailyStat

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

		// 当日新增用户
		newUsers, err := s.db.Collection("users").CountDocuments(context.Background(), bson.M{
			"created_at": bson.M{
				"$gte": startOfDay,
				"$lte": endOfDay,
			},
		})
		if err != nil {
			return nil, err
		}

		// 当日新增帖子
		newPosts, err := s.db.Collection("posts").CountDocuments(context.Background(), bson.M{
			"created_at": bson.M{
				"$gte": startOfDay,
				"$lte": endOfDay,
			},
		})
		if err != nil {
			return nil, err
		}

		// 当日新增评论
		newComments, err := s.db.Collection("comments").CountDocuments(context.Background(), bson.M{
			"created_at": bson.M{
				"$gte": startOfDay,
				"$lte": endOfDay,
			},
		})
		if err != nil {
			return nil, err
		}

		results = append(results, &DailyStat{
			Date:        startOfDay.Format("2006-01-02"),
			NewUsers:    int(newUsers),
			NewPosts:    int(newPosts),
			NewComments: int(newComments),
		})
	}

	return results, nil
}

func (s *AdminService) getTagStats() ([]*TagStat, error) {
	// 获取所有标签及其使用次数
	pipeline := []bson.M{
		{"$unwind": "$tags"},
		{"$group": bson.M{
			"_id":   "$tags",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	}

	cursor, err := s.db.Collection("posts").Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*TagStat
	for cursor.Next(context.Background()) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, &TagStat{
			Tag:   result.ID,
			Count: result.Count,
		})
	}

	return results, nil
}

func (s *AdminService) GetPendingPosts(page, limit int) (*model.PostListResponse, error) {
	skip := (page - 1) * limit

	// 查询条件
	filter := bson.M{"status": "pending"}

	// 获取总数
	total, err := s.db.Collection("posts").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := s.db.Collection("posts").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []model.Post
	if err = cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	// 转换为响应格式
	postItems := make([]model.PostListItem, 0, len(posts))
	for _, post := range posts {
		item := model.PostListItem{
			ID:         post.ID,
			PostID:     post.ID.Hex(),
			Title:      post.Title,
			Content:    post.Content,
			Type:       post.Type,
			Tags:       post.Tags,
			Files:      post.Files,
			CoverImage: post.CoverImage,
			UserID:     post.UserID,
			Username:   post.Username,
			Nickname:   post.Nickname,
			Avatar:     post.Avatar,
			CreatedAt:  post.CreatedAt,
			UpdatedAt:  post.UpdatedAt,
			Status:     post.Status,
		}

		// 如果 CoverImage 为空，则使用第一张图片作为封面
		if item.CoverImage == "" && len(post.Files) > 0 {
			item.CoverImage = post.Files[0]
		}

		// 设置用户信息
		item.User.UserID = post.UserID.Hex()
		item.User.Nickname = post.Nickname
		item.User.Avatar = post.Avatar

		postItems = append(postItems, item)
	}

	return &model.PostListResponse{
		Total: int(total),
		List:  postItems,
	}, nil
}

// GetUsers 获取用户列表
func (s *AdminService) GetUsers(page, limit int, search string) (*model.UserListResponse, error) {
	skip := (page - 1) * limit

	// 构建查询条件
	filter := bson.M{}
	if search != "" {
		// 使用正则表达式进行模糊搜索
		regex := regexp.MustCompile(search)
		filter = bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": regex}},
				{"nickname": bson.M{"$regex": regex}},
			},
		}
	}

	// 获取总数
	total, err := s.db.Collection("users").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// 查询数据
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := s.db.Collection("users").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []model.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	// 转换为响应格式
	userItems := make([]model.UserListItem, 0, len(users))
	for _, user := range users {
		item := model.UserListItem{
			UserID:    user.ID.Hex(),
			Username:  user.Username,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Bio:       user.Bio,
			Status:    user.Status,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		}
		userItems = append(userItems, item)
	}

	return &model.UserListResponse{
		Total: int(total),
		List:  userItems,
	}, nil
}

// UpdateUserRole 更新用户角色
func (s *AdminService) UpdateUserRole(userID string, role string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 获取用户信息
	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"role":       role,
			"updated_at": time.Now(),
		},
	}

	_, err = s.db.Collection("users").UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	// 发送角色变更通知
	var title, content string
	switch role {
	case "admin":
		title = "您已被设为管理员"
		content = "恭喜！您已被授予管理员权限。现在您可以管理所有用户、帖子和评论，并对提交的帖子进行审核。"
	case "user":
		title = "您的角色已变更为普通用户"
		content = "您的账户角色已被设置为普通用户。"
	case "vip":
		title = "您已成为VIP用户"
		content = "恭喜！您的账户已升级为VIP用户，可以享受更多功能和特权。"
	default:
		title = "您的账户角色已更新"
		content = fmt.Sprintf("您的账户角色已更新为：%s", role)
	}

	notification := &model.CreateNotificationRequest{
		Type:        model.NotificationTypeSystem,
		Title:       title,
		Content:     content,
		RelatedType: "user_role",
	}

	// 获取通知服务
	notificationService := NewNotificationService(s.db)
	_, err = notificationService.CreateNotification(userID, notification)
	if err != nil {
		log.Printf("创建角色变更通知失败: %v", err)
		// 不影响角色更新功能，继续执行
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *AdminService) DeleteUser(userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 更新用户状态为已注销
	userFilter := bson.M{"_id": objectID}
	userUpdate := bson.M{
		"$set": bson.M{
			"status":     "deleted",
			"username":   "已注销用户",
			"nickname":   "已注销用户",
			"avatar":     "", // 清空头像
			"bio":        "", // 清空简介
			"updated_at": time.Now(),
		},
	}
	_, err = s.db.Collection("users").UpdateOne(context.Background(), userFilter, userUpdate)
	if err != nil {
		return err
	}

	// 更新用户发布的帖子
	postFilter := bson.M{"user_id": objectID}
	postUpdate := bson.M{
		"$set": bson.M{
			"username": "已注销用户",
			"nickname": "已注销用户",
			"avatar":   "",
			"updated_at": time.Now(),
		},
	}
	_, err = s.db.Collection("posts").UpdateMany(context.Background(), postFilter, postUpdate)
	if err != nil {
		return err
	}

	// 更新用户的评论
	commentFilter := bson.M{"user_id": objectID}
	commentUpdate := bson.M{
		"$set": bson.M{
			"username": "已注销用户",
			"nickname": "已注销用户",
			"avatar":   "",
			"updated_at": time.Now(),
		},
	}
	_, err = s.db.Collection("comments").UpdateMany(context.Background(), commentFilter, commentUpdate)
	if err != nil {
		return err
	}

	// 删除用户的关注关系
	_, err = s.db.Collection("user_follows").DeleteMany(context.Background(), bson.M{
		"$or": []bson.M{
			{"user_id": objectID},
			{"following_id": objectID},
		},
	})

	return err
} 