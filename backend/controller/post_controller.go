package controller

import (
	"blue-note/model"
	"blue-note/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostController struct {
	postService *service.PostService
}

func NewPostController(postService *service.PostService) *PostController {
	return &PostController{postService: postService}
}

func (c *PostController) CreatePost(ctx *gin.Context) {
	var req model.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 从上下文中获取用户信息
	userID := ctx.GetString("userId")
	username := ctx.GetString("username")
	nickname := ctx.GetString("nickname")
	avatar := ctx.GetString("avatar")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID格式错误",
		})
		return
	}

	user := &model.User{
		ID:       objectID,
		Username: username,
		Nickname: nickname,
		Avatar:   avatar,
	}

	post, err := c.postService.CreatePost(user, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建帖子失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    post,
	})
}

func (c *PostController) GetPostList(ctx *gin.Context) {
	// 判断是否使用游标分页
	if ctx.Query("cursor") != "" {
		c.GetPostsWithCursor(ctx)
		return
	}

	var query model.PostQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	result, err := c.postService.GetPostList(&query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取帖子列表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// GetPostsWithCursor 获取帖子列表（使用游标分页）
func (c *PostController) GetPostsWithCursor(ctx *gin.Context) {
	var query model.CursorQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	result, err := c.postService.GetPostListWithCursor(&query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取帖子列表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

func (c *PostController) GetPostDetail(ctx *gin.Context) {
	postID := ctx.Param("postId")
	post, err := c.postService.GetPostDetail(postID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "帖子不存在",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    post,
	})
}

func (c *PostController) UpdatePost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	var req model.UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	userID := ctx.GetString("userId")
	post, err := c.postService.UpdatePost(postID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    post,
	})
}

func (c *PostController) DeletePost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.GetString("userId")

	err := c.postService.DeletePost(postID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// 获取帖子评论列表
func (c *PostController) GetPostComments(ctx *gin.Context) {
	postID, err := primitive.ObjectIDFromHex(ctx.Param("postId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的帖子ID"})
		return
	}

	var query model.CommentQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数"})
		return
	}

	comments, total, err := c.postService.GetPostComments(postID, &query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"total":    total,
		"page":     query.Page,
		"pageSize": query.PageSize,
		"sortBy":   query.SortBy,
		"order":    query.Order,
	})
}

// 创建评论
func (c *PostController) CreateComment(ctx *gin.Context) {
	postID, err := primitive.ObjectIDFromHex(ctx.Param("postId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的帖子ID"})
		return
	}

	var req model.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	comment, err := c.postService.CreateComment(postID, userID.(primitive.ObjectID), req.Content)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	ctx.JSON(http.StatusCreated, comment)
}

// 点赞评论
func (c *PostController) LikeComment(ctx *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(ctx.Param("commentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论ID"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	err = c.postService.LikeComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		if err.Error() == "已经点赞过了" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

// 取消点赞评论
func (c *PostController) UnlikeComment(ctx *gin.Context) {
	commentID, err := primitive.ObjectIDFromHex(ctx.Param("commentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的评论ID"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	err = c.postService.UnlikeComment(commentID, userID.(primitive.ObjectID))
	if err != nil {
		if err.Error() == "未找到点赞记录" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "取消点赞成功"})
}

func (c *PostController) DeleteComment(ctx *gin.Context) {
	postID := ctx.Param("postId")
	commentID := ctx.Param("commentId")
	userID := ctx.GetString("userId")

	err := c.postService.DeleteComment(postID, commentID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

func (c *PostController) ReviewPost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	var req model.ReviewPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	err := c.postService.ReviewPost(postID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "审核帖子失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// 点赞帖子
func (c *PostController) LikePost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.GetString("userId")

	err := c.postService.LikePost(postID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "点赞失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// 取消点赞
func (c *PostController) UnlikePost(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.GetString("userId")

	err := c.postService.UnlikePost(postID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "取消点赞失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
	})
}

// 检查是否已点赞
func (c *PostController) CheckLikeStatus(ctx *gin.Context) {
	postID := ctx.Param("postId")
	userID := ctx.GetString("userId")

	hasLiked, err := c.postService.HasLiked(postID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "检查点赞状态失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"hasLiked": hasLiked,
		},
	})
}

// SaveDraft 保存草稿
func (c *PostController) SaveDraft(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	
	var req model.CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取草稿ID（如果有）
	draftID := ctx.Query("draftId")
	
	// 保存草稿
	draft, err := c.postService.SaveDraft(userID, &req, draftID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "保存草稿失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "草稿保存成功",
		"data":    draft,
	})
}

// GetUserDrafts 获取用户草稿列表
func (c *PostController) GetUserDrafts(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	
	// 获取草稿列表
	result, err := c.postService.GetUserDrafts(userID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取草稿列表失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// GetDraftByID 获取草稿详情
func (c *PostController) GetDraftByID(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	draftID := ctx.Param("draftId")
	
	// 获取草稿详情
	draft, err := c.postService.GetDraftByID(draftID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取草稿详情失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    draft,
	})
}

// DeleteDraft 删除草稿
func (c *PostController) DeleteDraft(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	draftID := ctx.Param("draftId")
	
	// 删除草稿
	err := c.postService.DeleteDraft(draftID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "删除草稿失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "草稿删除成功",
	})
}

// PublishDraft 发布草稿
func (c *PostController) PublishDraft(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	draftID := ctx.Param("draftId")
	
	var req model.UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 如果没有提供更新内容，使用原草稿内容发布
		if err.Error() == "EOF" {
			req = model.UpdatePostRequest{}
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40003,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			return
		}
	}
	
	// 发布草稿
	post, err := c.postService.PublishDraft(draftID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "发布草稿失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "发布成功",
		"data":    post,
	})
}
