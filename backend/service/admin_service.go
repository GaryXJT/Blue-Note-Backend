package service

import (
	"blue-note/model"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
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
	query := &model.PostQuery{
		Page:   page,
		Limit:  limit,
		Status: "pending",
	}

	// 使用PostService获取待审核帖子列表
	postService := NewPostService(s.db, nil)
	return postService.GetPostList(query)
} 