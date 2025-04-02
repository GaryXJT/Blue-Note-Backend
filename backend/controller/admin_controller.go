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