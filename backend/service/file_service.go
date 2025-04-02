package service

import (
	"blue-note/model"
	"context"
	"fmt"
	"log"
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
func (s *FileService) MarkTemporary(userID string, filePath string) error {
	// 实现标记临时文件的逻辑
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}
	
	now := time.Now()
	fileRecord := model.FileRecord{
		FilePath:  filePath,
		URL:       s.objectStorageService.GetFileURL(filePath),
		UserID:    userObjID,
		Status:    model.FileStatusTemporary,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	_, err = s.db.Collection("file_records").InsertOne(context.Background(), fileRecord)
	if err != nil {
		return fmt.Errorf("创建文件记录失败: %w", err)
	}
	
	return nil
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