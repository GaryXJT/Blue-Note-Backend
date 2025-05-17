package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment 评论模型（树状结构）
type Comment struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	PostID       primitive.ObjectID   `bson:"post_id" json:"postId"`
	UserID       primitive.ObjectID   `bson:"user_id" json:"userId"`
	Username     string               `bson:"username" json:"username"`
	Nickname     string               `bson:"nickname" json:"nickname"`
	Avatar       string               `bson:"avatar" json:"avatar"`
	Content      string               `bson:"content" json:"content"`
	ParentID     *primitive.ObjectID  `bson:"parent_id,omitempty" json:"parentId,omitempty"` // 父评论ID，顶级评论为nil
	RootID       *primitive.ObjectID  `bson:"root_id,omitempty" json:"rootId,omitempty"`     // 根评论ID，顶级评论为nil
	ReplyToID    *primitive.ObjectID  `bson:"reply_to_id,omitempty" json:"replyToId,omitempty"` // 回复目标用户ID
	ReplyToName  string               `bson:"reply_to_name,omitempty" json:"replyToName,omitempty"` // 回复目标用户名
	Likes        int                  `bson:"likes" json:"likes"`         // 点赞数
	ChildrenCount int                 `bson:"children_count" json:"childrenCount"` // 子评论数
	Level        int                  `bson:"level" json:"level"`         // 评论层级，0为顶级评论
	IsAuthor     bool                 `bson:"is_author" json:"isAuthor"`  // 是否是帖子作者
	IsAdmin      bool                 `bson:"is_admin" json:"isAdmin"`    // 是否是管理员
	Status       string               `bson:"status" json:"status"`      // 状态：normal-正常，deleted-已删除，hidden-已隐藏
	Score        float64              `bson:"score" json:"score"`        // 评论综合评分，用于排序
	CreatedAt    time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updatedAt"`
	LikedByUser  bool                 `bson:"liked_by_user,omitempty" json:"likedByUser,omitempty"` // 当前用户是否点赞（非数据库字段）
	Children     []*CommentWithChildren `bson:"-" json:"children,omitempty"` // 子评论列表（非数据库字段）
}

// CommentWithChildren 带有子评论的评论结构（用于API响应）
type CommentWithChildren struct {
	ID           string               `json:"id"`
	PostID       string               `json:"postId"`
	UserID       string               `json:"userId"`
	Username     string               `json:"username"`
	Nickname     string               `json:"nickname"`
	Avatar       string               `json:"avatar"`
	Content      string               `json:"content"`
	ParentID     string               `json:"parentId,omitempty"`
	RootID       string               `json:"rootId,omitempty"`
	ReplyToID    string               `json:"replyToId,omitempty"`
	ReplyToName  string               `json:"replyToName,omitempty"`
	Likes        int                  `json:"likes"`
	ChildrenCount int                 `json:"childrenCount"`
	Level        int                  `json:"level"`
	IsAuthor     bool                 `json:"isAuthor"`
	IsAdmin      bool                 `json:"isAdmin"`
	Status       string               `json:"status"`
	Score        float64              `json:"score"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
	LikedByUser  bool                 `json:"likedByUser"`
	Children     []*CommentWithChildren `json:"children,omitempty"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	PostID       string `json:"postId" binding:"required"`
	Content      string `json:"content" binding:"required,min=1,max=1000"`
	ParentID     string `json:"parentId,omitempty"` // 父评论ID，如果为空则为顶级评论
	ReplyToID    string `json:"replyToId,omitempty"` // 回复目标用户ID
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content      string `json:"content" binding:"required,min=1,max=1000"`
}

// GetCommentsQuery 获取评论列表查询参数
type GetCommentsQuery struct {
	PostID       string `form:"postId" binding:"required"`
	ParentID     string `form:"parentId,omitempty"` // 父评论ID，为空则查询顶级评论
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	SortBy       string `form:"sortBy" binding:"omitempty,oneof=time likes"` // 排序方式
	Order        string `form:"order" binding:"omitempty,oneof=asc desc"` // 排序顺序
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments     []*CommentWithChildren `json:"comments"`
	Total        int                    `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"pageSize"`
	HasMore      bool                   `json:"hasMore"`
}

// CommentLike 评论点赞记录
type CommentLike struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty"`
	CommentID    primitive.ObjectID   `bson:"comment_id"`
	UserID       primitive.ObjectID   `bson:"user_id"`
	CreatedAt    time.Time            `bson:"created_at"`
} 