package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadService struct {
	uploadDir string
	baseURL   string
}

func NewUploadService(uploadDir, baseURL string) *UploadService {
	// 确保上传目录存在
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &UploadService{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (s *UploadService) UploadFile(file *multipart.FileHeader) (string, error) {
	// 获取文件扩展名
	ext := filepath.Ext(file.Filename)
	ext = strings.ToLower(ext)

	// 验证文件类型
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".mp4": true, ".mov": true, ".avi": true, ".webm": true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 生成文件名
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	filename := fmt.Sprintf("%d%s", timestamp, ext)
	filePath := filepath.Join(s.uploadDir, filename)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// 返回文件URL
	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

// 上传用户头像
func (s *UploadService) UploadAvatar(userID string, file io.Reader) (string, error) {
	return "", fmt.Errorf("未实现")
}

// 上传帖子图片
func (s *UploadService) UploadPostImage(postID string, imageIndex int, file io.Reader) (string, error) {
	// path := GeneratePostImagePath(postID, imageIndex)
	return "", fmt.Errorf("未实现")
}

// 上传广告图片
func (s *UploadService) UploadAdImage(adType string, adID string, file io.Reader) (string, error) {
	// path := GenerateAdImagePath(adType, adID)
	return "", fmt.Errorf("未实现")
} 