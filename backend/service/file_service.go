package service

import (
	"blue-note/model"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"  // 支持GIF格式
	_ "image/jpeg" // 支持JPEG格式
	_ "image/png"  // 支持PNG格式
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FileService 文件服务
type FileService struct {
	db                  *mongo.Database
	objectStorageService *ObjectStorageService
}

// NewFileService 创建文件服务实例
func NewFileService(db *mongo.Database, objectStorageService *ObjectStorageService) *FileService {
	return &FileService{
		db:                  db,
		objectStorageService: objectStorageService,
	}
}

// MarkTemporary 标记文件为临时状态
func (s *FileService) MarkTemporary(userID string, filePath string, fileSize int64, fileTypeHint string) error {
	// 实现标记临时文件的逻辑
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}
	
	now := time.Now()
	
	// 获取文件URL
	fileURL := s.objectStorageService.GetFileURL(filePath)
	
	// 获取文件类型和大小
	var fileType string
	
	// 如果传入了类型提示，优先使用
	if fileTypeHint != "" {
		fileType = fileTypeHint
	} else {
		// 根据文件路径判断是图片还是视频
		fileExt := strings.ToLower(filepath.Ext(filePath))
		if fileExt == ".jpg" || fileExt == ".jpeg" || fileExt == ".png" || fileExt == ".gif" || fileExt == ".webp" {
			fileType = "image"
		} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" || fileExt == ".avi" {
			fileType = "video"
		} else {
			fileType = "unknown"
		}
	}
	
	// 获取文件宽高
	width, height := 0, 0
	if fileType == "image" || fileType == "video" {
		w, h, err := s.GetFileDimensions(fileURL)
		if err == nil {
			width, height = w, h
		} else {
			log.Printf("获取文件尺寸失败: %v", err)
		}
	}
	
	fileRecord := model.FileRecord{
		FilePath:  filePath,
		URL:       fileURL,
		UserID:    userObjID,
		Status:    model.FileStatusTemporary,
		Size:      fileSize,
		Type:      fileType,
		Width:     width,
		Height:    height,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	_, err = s.db.Collection("file_records").InsertOne(context.Background(), fileRecord)
	if err != nil {
		return fmt.Errorf("创建文件记录失败: %w", err)
	}
	
	return nil
}

// GetFileDimensions 获取文件尺寸（图片或视频）
func (s *FileService) GetFileDimensions(fileURL string) (int, int, error) {
	// 检查文件是否在数据库中有记录
	var fileRecord model.FileRecord
	err := s.db.Collection("file_records").FindOne(
		context.Background(),
		bson.M{"url": fileURL},
	).Decode(&fileRecord)
	
	if err == nil && fileRecord.Width > 0 && fileRecord.Height > 0 {
		// 如果数据库中有记录并且包含尺寸信息，直接返回
		return fileRecord.Width, fileRecord.Height, nil
	}
	
	// 尝试从本地文件获取尺寸
	if strings.HasPrefix(fileURL, "/uploads/") {
		localPath := "." + fileURL
		return s.getLocalFileDimensions(localPath)
	}
	
	// 尝试从远程URL获取尺寸
	return s.getRemoteFileDimensions(fileURL)
}

// 获取本地文件尺寸
func (s *FileService) getLocalFileDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()
	
	// 检查文件类型
	fileExt := strings.ToLower(filepath.Ext(filePath))
	
	// 图片处理
	if fileExt == ".jpg" || fileExt == ".jpeg" || fileExt == ".png" || fileExt == ".gif" || fileExt == ".webp" {
		img, _, err := image.DecodeConfig(file)
		if err != nil {
			return 0, 0, fmt.Errorf("解析图片配置失败: %w", err)
		}
		return img.Width, img.Height, nil
	}
	
	// 视频处理 - 简化方式，实际项目应该使用ffprobe等工具
	// 这里仅返回默认值，实际项目中需要实现视频尺寸的提取
	if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" || fileExt == ".avi" {
		return 1280, 720, nil
	}
	
	return 0, 0, fmt.Errorf("不支持的文件类型: %s", fileExt)
}

// 获取远程文件尺寸
func (s *FileService) getRemoteFileDimensions(fileURL string) (int, int, error) {
	// 下载文件头部
	resp, err := http.Get(fileURL)
	if err != nil {
		return 0, 0, fmt.Errorf("请求文件失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取头部数据以检测类型
	buffer := make([]byte, 512)
	n, err := resp.Body.Read(buffer)
	if err != nil && err != io.EOF {
		return 0, 0, fmt.Errorf("读取文件头失败: %w", err)
	}
	
	contentType := http.DetectContentType(buffer[:n])
	
	// 如果是图片，尝试解析尺寸
	if strings.HasPrefix(contentType, "image/") {
		// 读取剩余数据
		restData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, 0, fmt.Errorf("读取图片数据失败: %w", err)
		}
		
		// 合并数据
		fullData := append(buffer[:n], restData...)
		
		// 使用image.DecodeConfig从图片数据中提取宽高信息
		// 这只会解析图片的头部元数据，不会加载整个图片
		img, _, err := image.DecodeConfig(bytes.NewReader(fullData))
		if err != nil {
			return 0, 0, fmt.Errorf("解析图片配置失败: %w", err)
		}
		
		return img.Width, img.Height, nil
	}
	
	// 视频处理 - 简化方式
	if strings.HasPrefix(contentType, "video/") {
		return 1280, 720, nil
	}
	
	return 0, 0, fmt.Errorf("不支持的文件类型: %s", contentType)
}

// MarkUsed 标记文件为已使用状态
func (s *FileService) MarkUsed(filePaths []string) error {
	// 实现标记已使用文件的逻辑
	if len(filePaths) == 0 {
		return nil
	}
	
	now := time.Now()
	_, err := s.db.Collection("file_records").UpdateMany(
		context.Background(),
		bson.M{"file_path": bson.M{"$in": filePaths}},
		bson.M{
			"$set": bson.M{
				"status":     model.FileStatusUsed,
				"updated_at": now,
				"used_at":    now,
			},
		},
	)
	
	if err != nil {
		return fmt.Errorf("更新文件状态失败: %w", err)
	}
	
	return nil
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(userID string, filePath string) error {
	// 实现删除文件的逻辑
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}
	
	// 检查文件所有权
	var fileRecord model.FileRecord
	err = s.db.Collection("file_records").FindOne(
		context.Background(),
		bson.M{
			"file_path": filePath,
			"user_id":   userObjID,
		},
	).Decode(&fileRecord)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("文件不存在或无权限删除")
		}
		return fmt.Errorf("查询文件记录失败: %w", err)
	}
	
	// 从对象存储中删除文件
	err = s.objectStorageService.DeleteFile(filePath)
	if err != nil {
		log.Printf("从对象存储删除文件失败: %v", err)
	}
	
	// 更新文件状态为已删除
	_, err = s.db.Collection("file_records").UpdateOne(
		context.Background(),
		bson.M{"file_path": filePath},
		bson.M{
			"$set": bson.M{
				"status":     model.FileStatusDeleted,
				"updated_at": time.Now(),
			},
		},
	)
	
	if err != nil {
		return fmt.Errorf("更新文件状态失败: %w", err)
	}
	
	return nil
}

// GetImageDimensions 获取图片尺寸（为兼容旧代码保留的方法）
func (s *FileService) GetImageDimensions(filePath string) (int, int, error) {
	// 调用新的通用方法
	return s.GetFileDimensions(s.objectStorageService.GetFileURL(filePath))
} 