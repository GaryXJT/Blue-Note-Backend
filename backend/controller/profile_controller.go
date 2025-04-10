package controller

import (
	"blue-note/model"
	"blue-note/service"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileService *service.ProfileService
}

func NewProfileController(profileService *service.ProfileService) *ProfileController {
	return &ProfileController{profileService: profileService}
}

// GetProfile 获取用户资料
func (c *ProfileController) GetProfile(ctx *gin.Context) {
	userID := ctx.Param("userId")
	currentUserID := ctx.GetString("userId")

	profile, err := c.profileService.GetUserProfile(userID, currentUserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取用户资料失败",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    profile,
	})
}

// UpdateProfile 更新用户资料（包含头像上传）
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	// 使用 Gin 的日志
	ctx.Set("startTime", time.Now())
	
	// 打印请求信息
	ctx.Set("requestPath", ctx.Request.URL.Path)
	ctx.Set("requestMethod", ctx.Request.Method)
	
	// 获取并打印 userId
	userID := ctx.GetString("userId")
	ctx.Set("userId", userID)
	
	// 打印到控制台
	fmt.Println("==== UpdateProfile 被调用 ====")
	fmt.Printf("userID: %s\n", userID)
	
	// 打印所有请求参数
	fmt.Println("请求参数:")
	for key, values := range ctx.Request.URL.Query() {
		fmt.Printf("  %s: %v\n", key, values)
	}
	
	// 打印表单数据
	if err := ctx.Request.ParseMultipartForm(10 << 20); err == nil {
		fmt.Println("表单数据:")
		for key, values := range ctx.Request.Form {
			fmt.Printf("  %s: %v\n", key, values)
		}
	}
	
	// 处理表单数据
	var req model.UpdateProfileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}
	
	// 处理头像文件上传
	file, err := ctx.FormFile("avatar")
	fmt.Printf("头像文件: %v, 错误: %v\n", file, err)
	
	var avatarFile io.Reader
	if file != nil {
		f, err := file.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40006,
				"message": "上传文件处理失败",
				"error":   err.Error(),
			})
			return
		}
		defer f.Close()
		
		// 检查文件内容
		buf := make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40007,
				"message": "读取文件失败",
				"error":   err.Error(),
			})
			return
		}
		
		// 重置文件位置
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40008,
				"message": "处理文件失败",
				"error":   err.Error(),
			})
			return
		}
		
		fmt.Printf("文件内容前 %d 字节: %v\n", n, buf[:n])
		avatarFile = f
	}
	
	// 检查对象存储服务状态
	isStorageAvailable := c.profileService.IsObjectStorageAvailable()
	fmt.Printf("对象存储服务状态: %v\n", isStorageAvailable)
	
	// 调用服务更新资料和头像
	profile, err := c.profileService.UpdateProfileWithAvatar(userID, &req, avatarFile)
	if err != nil {
		// 记录错误但返回一个友好的错误信息
		fmt.Printf("更新用户资料失败: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "更新用户资料失败，请稍后再试",
			"error":   "服务器内部错误",
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "个人资料更新成功",
		"data":    profile,
	})
}

// FollowUser 关注用户
func (c *ProfileController) FollowUser(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	followingID := ctx.Param("userId")
	
	result, err := c.profileService.FollowUser(userID, followingID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "关注用户失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "关注成功",
		"data":    result,
	})
}

// UnfollowUser 取消关注用户
func (c *ProfileController) UnfollowUser(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	followingID := ctx.Param("userId")
	
	result, err := c.profileService.UnfollowUser(userID, followingID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "取消关注失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "取消关注成功",
		"data":    result,
	})
}

// CheckFollowStatus 检查关注状态
func (c *ProfileController) CheckFollowStatus(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	followingID := ctx.Param("userId")
	
	isFollowing, err := c.profileService.CheckFollowStatus(userID, followingID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "检查关注状态失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"isFollowing": isFollowing,
		},
	})
}

// GetFollowingList 获取关注列表
func (c *ProfileController) GetFollowingList(ctx *gin.Context) {
	userID := ctx.Param("userId")
	currentUserID := ctx.GetString("userId")
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	result, err := c.profileService.GetFollowingList(userID, currentUserID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取关注列表失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    result,
	})
}

// GetFansList 获取粉丝列表
func (c *ProfileController) GetFansList(ctx *gin.Context) {
	userID := ctx.Param("userId")
	currentUserID := ctx.GetString("userId")
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	result, err := c.profileService.GetFansList(userID, currentUserID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取粉丝列表失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    result,
	})
}

// GetUserLikedPosts 获取用户喜欢的笔记
func (c *ProfileController) GetUserLikedPosts(ctx *gin.Context) {
	userID := ctx.Param("userId")
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	result, err := c.profileService.GetUserLikedPosts(userID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取喜欢的笔记失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    result,
	})
}

// GetUserCollectedPosts 获取用户收藏的笔记
func (c *ProfileController) GetUserCollectedPosts(ctx *gin.Context) {
	userID := ctx.Param("userId")
	
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	
	result, err := c.profileService.GetUserCollectedPosts(userID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "获取收藏的笔记失败",
			"error":   err.Error(),
		})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    result,
	})
} 