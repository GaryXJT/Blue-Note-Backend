package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileStatus 表示文件的状态
type FileStatus string

const (
	FileStatusTemporary FileStatus = "temporary" // 临时文件
	FileStatusUsed      FileStatus = "used"      // 已使用的文件
	FileStatusDeleted   FileStatus = "deleted"   // 已删除的文件
)

// FileRecord 文件记录
type FileRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FilePath  string             `bson:"file_path" json:"filePath"`     // 文件路径
	URL       string             `bson:"url" json:"url"`                // 文件URL
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`         // 上传用户ID
	Status    FileStatus         `bson:"status" json:"status"`          // 文件状态
	Size      int64              `bson:"size" json:"size"`              // 文件大小(字节)
	Type      string             `bson:"type" json:"type"`              // 文件类型(image/video)
	Width     int                `bson:"width,omitempty" json:"width,omitempty"`           // 图片宽度(像素)
	Height    int                `bson:"height,omitempty" json:"height,omitempty"`         // 图片高度(像素)
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`   // 创建时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`   // 更新时间
	UsedAt    *time.Time         `bson:"used_at,omitempty" json:"usedAt,omitempty"` // 使用时间
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	FilePath string `json:"filePath" binding:"required"`
} 