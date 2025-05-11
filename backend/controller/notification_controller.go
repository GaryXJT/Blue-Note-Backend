package controller

import (
	"blue-note/model"
	"blue-note/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	notificationService *service.NotificationService
}

func NewNotificationController(notificationService *service.NotificationService) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
	}
}

// GetNotificationList 获取通知列表
func (c *NotificationController) GetNotificationList(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var query model.NotificationQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 处理isRead参数
	if isReadStr := ctx.Query("isRead"); isReadStr != "" {
		isRead, err := strconv.ParseBool(isReadStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的isRead参数"})
			return
		}
		query.IsRead = &isRead
	}

	response, err := c.notificationService.GetNotificationList(userID, &query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateNotification 更新通知
func (c *NotificationController) UpdateNotification(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	notificationID := ctx.Param("id")
	if notificationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "通知ID不能为空"})
		return
	}

	var req model.UpdateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.notificationService.UpdateNotification(notificationID, userID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteNotification 删除通知
func (c *NotificationController) DeleteNotification(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	notificationID := ctx.Param("id")
	if notificationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "通知ID不能为空"})
		return
	}

	err := c.notificationService.DeleteNotification(notificationID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// MarkAllAsRead 将所有通知标记为已读
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	err := c.notificationService.MarkAllAsRead(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "标记成功"})
}

// GetUnreadCount 获取未读通知数量
func (c *NotificationController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.GetString("userId")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	count, err := c.notificationService.GetUnreadCount(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
} 