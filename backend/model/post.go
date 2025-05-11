package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"postId"`
	Title     string            `bson:"title" json:"title"`
	Content   string            `bson:"content" json:"content"`
	Type      string            `bson:"type" json:"type"` // image/video
	Tags       []string `json:"tags" binding:"required,dive,oneof=穿搭 美食 彩妆 影视 职场 情感 家居 游戏 旅行 风景 健身 其他"`
	Files     []string          `bson:"files" json:"files"`
	CoverImage string            `bson:"cover_image" json:"coverImage"` // 封面图片
	Status    string            `bson:"status" json:"status"` // draft(草稿)/pending(审核中)/approved(已通过)/rejected(已拒绝)
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	Username  string            `bson:"username" json:"username"`
	Nickname  string            `bson:"nickname" json:"nickname"`
	Avatar    string            `bson:"avatar" json:"avatar"`
	Likes     int               `bson:"likes" json:"likes"`
	Comments  int               `bson:"comments" json:"comments"`
	CreatedAt time.Time         `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updatedAt"`
}

type CreatePostRequest struct {
	Title      string   `json:"title" binding:"required,max=100"`
	Content    string   `json:"content" binding:"required"`
	Type       string   `json:"type" binding:"required,oneof=image video"`
	Tags       []string `json:"tags" binding:"required,dive,oneof=穿搭 美食 彩妆 影视 职场 情感 家居 游戏 旅行 风景 健身 其他"`
	Files      []string `json:"files" binding:"required"`
	CoverImage string   `json:"coverImage"` // 封面图片
	IsDraft    bool     `json:"isDraft"`
}

type UpdatePostRequest struct {
	Title      string   `json:"title" binding:"omitempty,max=100"`
	Content    string   `json:"content" binding:"omitempty"`
	Tags       []string `json:"tags" binding:"omitempty,dive,oneof=穿搭 美食 彩妆 影视 职场 情感 家居 游戏 旅行 风景 健身 其他"`
	Files      []string `json:"files" binding:"omitempty"`
	CoverImage string   `json:"coverImage"` // 封面图片
	IsDraft    bool     `json:"isDraft"`
}

// PostQuery 帖子查询参数
type PostQuery struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Type       string `form:"type" binding:"omitempty"`
	Tag        string `form:"tag" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty"`
	UserID     string `form:"userId" binding:"omitempty"`
	Search     string `form:"search" binding:"omitempty"`
	SearchType string `form:"searchType" binding:"omitempty,oneof=all author content"`
}

// PostListResponse 笔记列表响应
type PostListResponse struct {
	Total int           `json:"total"`
	List  []PostListItem `json:"list"`
}

type CommentQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	SortBy   string `form:"sortBy" binding:"omitempty,oneof=time score likes"` // 排序方式：time-按时间，score-按评分，likes-按点赞数
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`         // 排序顺序：asc-升序，desc-降序
}

// 点赞记录
type PostLike struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	PostID    primitive.ObjectID `bson:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
}

// 帖子关注记录
type PostFollower struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	PostID    primitive.ObjectID `bson:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
}

// 包含点赞和关注状态的帖子详情返回
type PostDetailResponse struct {
	Post     *Post `json:"post"`
	HasLiked bool  `json:"hasLiked"`
	HasFollowed bool `json:"hasFollowed"`
}

type ReviewPostRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Reason string `json:"reason" binding:"omitempty,max=200"`
}

// 修改 PostListItem 结构体
type PostListItem struct {
	ID           primitive.ObjectID `json:"id"`
	PostID       string             `json:"postId"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	Type         string             `json:"type"`
	Tags         []string           `json:"tags"`
	Files        []string           `json:"files"`
	CoverImage   string             `json:"coverImage"`
	UserID       primitive.ObjectID `json:"userId"`
	Username     string             `json:"username"`
	Nickname     string             `json:"nickname"`
	Avatar       string             `json:"avatar"`
	Likes        int                `json:"likes"`
	Comments     int                `json:"comments"`
	LikeCount    int                `json:"likeCount"`
	CommentCount int                `json:"commentCount"`
	CollectCount int                `json:"collectCount"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	User         struct {
		UserID   string `json:"userId"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"user"`
	Status string `json:"status"`
}

// CursorQuery 基于游标的查询参数
type CursorQuery struct {
	Cursor        string `form:"cursor" binding:"omitempty"`
	Limit         int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Type          string `form:"type" binding:"omitempty,oneof=image video"`
	Tag           string `form:"tag" binding:"omitempty"`
	Status        string `form:"status" binding:"omitempty,oneof=draft pending approved rejected"`
	UserID        string `form:"userId" binding:"omitempty"`
	Search        string `form:"search" binding:"omitempty"`
	SearchType    string `form:"searchType" binding:"omitempty,oneof=all author content"`
	CurrentUserID string `form:"currentUserId" binding:"omitempty"` // 当前登录用户ID
	FilterUser    string `form:"filterUser" binding:"omitempty"`    // 筛选目标用户ID
	FilterType    string `form:"filterType" binding:"omitempty,oneof= onlyCurrentUser like follow"` // 筛选类型：空、onlyCurrentUser、like、follow
}

// CursorBasedPostResponse 基于游标的帖子列表响应
type CursorBasedPostResponse struct {
	Posts      []PostItem `json:"posts"`
	NextCursor string     `json:"nextCursor"`
	HasMore    bool       `json:"hasMore"`
}

// PostItem 帖子项，含有更多的字段用于瀑布流布局
type PostItem struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Type         string    `json:"type"`
	Tags         []string  `json:"tags"`
	Files        []string  `json:"files"`
	CoverImage   string    `json:"coverImage"`
	UserID       string    `json:"userId"`
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar"`
	Likes        int       `json:"likes"`
	Comments     int       `json:"comments"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	CollectCount int       `json:"collectCount"`
	CreatedAt    time.Time `json:"createdAt"`
	Status       string    `json:"status"`
	LikedByUser     bool   `json:"likedByUser"`
	FollowedByUser  bool   `json:"followedByUser"`
}