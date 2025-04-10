package controller

import (
	"blue-note/service"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type UploadController struct {
	objectStorageService *service.ObjectStorageService
	fileService          *service.FileService
}

func NewUploadController(objectStorageService *service.ObjectStorageService, fileService *service.FileService) *UploadController {
	return &UploadController{objectStorageService: objectStorageService, fileService: fileService}
}

// UploadFile 上传文件
func (c *UploadController) UploadFile(ctx *gin.Context) {
	// 获取请求ID
	requestID := uuid.New().String()
	
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40003,
			"message": "请求参数错误",
			"error":   "未找到文件",
		})
		return
	}
	
	// 记录开始上传
	log.Printf("[%s] 开始上传文件: %s, 大小: %d bytes", requestID, file.Filename, file.Size)
	
	// 获取文件类型
	fileType := ctx.DefaultPostForm("type", "image") // 默认为图片
	
	// 文件大小限制
	var maxSize int64
	if fileType == "video" {
		maxSize = 100 * 1024 * 1024 // 视频100MB
	} else {
		maxSize = 10 * 1024 * 1024 // 图片10MB
	}
	
	if file.Size > maxSize {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    40004,
			"message": fmt.Sprintf("文件大小超过限制(%dMB)", maxSize/1024/1024),
		})
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	
	// 允许的文件扩展名
	allowedImageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	allowedVideoExts := map[string]bool{".mp4": true, ".webm": true, ".mov": true, ".avi": true}
	
	if fileType == "video" {
		if !allowedVideoExts[ext] {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40005,
				"message": "不支持的视频格式，请上传MP4、WebM、MOV或AVI格式",
			})
			return
		}
	} else {
		if !allowedImageExts[ext] {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    40005,
				"message": "不支持的图片格式，请上传JPG、PNG、GIF或WebP格式",
			})
			return
		}
	}

	// 打开文件
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

	// 使用有限制的缓冲区读取文件
	limitedReader := io.LimitReader(f, file.Size)

	// 生成文件路径和确定内容类型
	filename := filepath.Base(file.Filename)
	objectName := fmt.Sprintf("%s/%s/%s", fileType, time.Now().Format("2006/01/02"), filename)
	contentType := file.Header.Get("Content-Type")

	// 调用上传方法
	result, err := c.objectStorageService.UploadFile(limitedReader, objectName, contentType)
	if err != nil {
		log.Printf("[%s] 上传失败: %v", requestID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    50001,
			"message": "上传文件失败",
			"error":   err.Error(),
		})
		return
	}

	// 直接在这里标记为临时文件，不再需要单独的API
	if c.fileService != nil {
		userID := ctx.GetString("userId")
		err := c.fileService.MarkTemporary(userID, objectName, file.Size, fileType)
		if err != nil {
			log.Printf("[%s] 标记临时文件失败: %v", requestID, err)
			// 即使标记失败也继续，不影响上传结果
		}
	}

	log.Printf("[%s] 上传成功: %s", requestID, result)

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "文件上传成功",
		"data": gin.H{
			"url": result,
			"type": fileType,
			"size": file.Size,
			"name": filename,
		},
	})
} 