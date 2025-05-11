package controller

import (
	"blue-note/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService        *service.AdminService
	objectStorageService *service.ObjectStorageService
}

func NewAdminController(adminService *service.AdminService, objectStorageService *service.ObjectStorageService) *AdminController {
	return &AdminController{
		adminService:        adminService,
		objectStorageService: objectStorageService,
	}
}

func (c *AdminController) GetStatistics(ctx *gin.Context) {
	stats, err := c.adminService.GetStatistics()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取统计数据失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    stats,
	})
}

func (c *AdminController) GetPendingPosts(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := c.adminService.GetPendingPosts(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取待审核帖子失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
} 

// GetUsers 获取用户列表
func (c *AdminController) GetUsers(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")
	search := ctx.DefaultQuery("search", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := c.adminService.GetUsers(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取用户列表失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

// UpdateUserRole 更新用户角色
func (c *AdminController) UpdateUserRole(ctx *gin.Context) {
	userID := ctx.Param("userId")
	role := ctx.Param("role")

	if role != "admin" && role != "user" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色类型",
		})
		return
	}

	err := c.adminService.UpdateUserRole(userID, role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新用户角色失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新用户角色成功",
	})
}

// DeleteUser 删除用户
func (c *AdminController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("userId")

	err := c.adminService.DeleteUser(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除用户失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除用户成功",
	})
} 