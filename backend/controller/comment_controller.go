package controller

import (
	"blue-note/model"
	"blue-note/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CommentController 评论控制器
type CommentController struct {
	commentService *service.CommentService
}

// NewCommentController 创建评论控制器
func NewCommentController(commentService *service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

// CreateComment 创建评论
func (c *CommentController) CreateComment(ctx *gin.Context) {
	var req model.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户信息
	userID := ctx.GetString("userId")
	username := ctx.GetString("username")
	nickname := ctx.GetString("nickname")
	avatar := ctx.GetString("avatar")
	role := ctx.GetString("role") // 假设中间件设置了用户角色

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	// 构建当前用户对象
	currentUser := &model.User{
		ID:       userObjectID,
		Username: username,
		Nickname: nickname,
		Avatar:   avatar,
		Role:     role,
	}

	// 创建评论
	comment, err := c.commentService.CreateComment(&req, currentUser)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建评论失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "评论成功",
		"data":    comment,
	})
}

// GetComments 获取评论列表
func (c *CommentController) GetComments(ctx *gin.Context) {
	var query model.GetCommentsQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID（用于检查点赞状态）
	currentUserID := ctx.GetString("userId")

	// 获取评论列表
	commentList, err := c.commentService.GetComments(&query, currentUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取评论失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取评论成功",
		"data":    commentList,
	})
}

// DeleteComment 删除评论
func (c *CommentController) DeleteComment(ctx *gin.Context) {
	commentID := ctx.Param("commentId")
	userID := ctx.GetString("userId")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 删除评论
	err := c.commentService.DeleteComment(commentID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除评论失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除评论成功",
	})
}

// LikeComment 点赞评论
func (c *CommentController) LikeComment(ctx *gin.Context) {
	commentID := ctx.Param("commentId")
	userID := ctx.GetString("userId")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 点赞评论
	err := c.commentService.LikeComment(commentID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "点赞失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "点赞成功",
	})
}

// UnlikeComment 取消点赞评论
func (c *CommentController) UnlikeComment(ctx *gin.Context) {
	commentID := ctx.Param("commentId")
	userID := ctx.GetString("userId")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 取消点赞评论
	err := c.commentService.UnlikeComment(commentID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "取消点赞失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "取消点赞成功",
	})
}

// CheckLikeStatus 检查评论点赞状态
func (c *CommentController) CheckLikeStatus(ctx *gin.Context) {
	commentID := ctx.Param("commentId")
	userID := ctx.GetString("userId")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 检查点赞状态
	hasLiked, err := c.commentService.HasLikedComment(commentID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查点赞状态失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取点赞状态成功",
		"data": gin.H{
			"hasLiked": hasLiked,
		},
	})
} 