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
	Tags      []string          `bson:"tags" json:"tags"`
	Files     []string          `bson:"files" json:"files"`
	CoverImage string            `bson:"cover_image" json:"coverImage"` // 封面图片
	Status    string            `bson:"status" json:"status"` // draft/pending/approved/rejected
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
	Tags       []string `json:"tags" binding:"required"`
	Files      []string `json:"files" binding:"required"`
	CoverImage string   `json:"coverImage"` // 封面图片
	IsDraft    bool     `json:"isDraft"`
}

type UpdatePostRequest struct {
	Title      string   `json:"title" binding:"omitempty,max=100"`
	Content    string   `json:"content" binding:"omitempty"`
	Tags       []string `json:"tags" binding:"omitempty"`
	Files      []string `json:"files" binding:"omitempty"`
	CoverImage string   `json:"coverImage"` // 封面图片
	IsDraft    bool     `json:"isDraft"`
}

type PostQuery struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Type    string `form:"type" binding:"omitempty,oneof=image video"`
	Tag     string `form:"tag" binding:"omitempty"`
	Status  string `form:"status" binding:"omitempty,oneof=draft pending approved rejected"`
	UserID  string `form:"userId" binding:"omitempty"`
}

// PostListResponse 笔记列表响应
type PostListResponse struct {
	Total int           `json:"total"`
	List  []PostListItem `json:"list"`
}

type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string            `bson:"username" json:"username"`
	Nickname  string            `bson:"nickname" json:"nickname"`
	Avatar    string            `bson:"avatar" json:"avatar"`
	Content   string            `bson:"content" json:"content"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
	Likes     int               `bson:"likes" json:"likes"`
	Score     float64           `bson:"score" json:"score"` // 综合评分
	IsAuthor  bool              `bson:"is_author" json:"is_author"` // 是否是作者评论
	IsAdmin   bool              `bson:"is_admin" json:"is_admin"` // 是否是管理员评论
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
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
}

// 评论点赞记录
type CommentLike struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CommentID primitive.ObjectID `bson:"comment_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt time.Time         `bson:"created_at"`
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
}