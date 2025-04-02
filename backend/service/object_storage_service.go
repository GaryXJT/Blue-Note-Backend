package service

import (
	"blue-note/config"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorageService struct {
	client         *minio.Client
	bucketName     string
	internalEndpoint string
	externalEndpoint string
}

func NewObjectStorageService(cfg *config.Config) (*ObjectStorageService, error) {
	// 使用内部端点
	endpoint := cfg.ObjectStorage.InternalEndpoint
	
	// 创建自定义传输配置，增加超时时间
	transport := &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	// 使用自定义HTTP客户端创建MinIO客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.ObjectStorage.AccessKey, cfg.ObjectStorage.SecretKey, ""),
		Secure:    cfg.ObjectStorage.UseSSL,
		Transport: transport,
	})
	
	if err != nil {
		fmt.Printf("创建MinIO客户端失败: %v\n", err)
		return &ObjectStorageService{
			client:           nil,
			bucketName:       "h49hpg7e-blue-note",
			internalEndpoint: cfg.ObjectStorage.InternalEndpoint,
			externalEndpoint: cfg.ObjectStorage.ExternalEndpoint,
		}, nil
	}
	
	fmt.Printf("MinIO客户端创建成功\n")
	
	// 创建服务实例
	service := &ObjectStorageService{
		client:           minioClient,
		bucketName:       "h49hpg7e-blue-note",
		internalEndpoint: cfg.ObjectStorage.InternalEndpoint,
		externalEndpoint: cfg.ObjectStorage.ExternalEndpoint,
	}
	
	// 在后台完成初始化
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// 测试连接 - 列出所有存储桶
		buckets, err := minioClient.ListBuckets(ctx)
		if err != nil {
			fmt.Printf("列出存储桶失败: %v\n", err)
			return
		}
		
		fmt.Printf("成功列出存储桶，共 %d 个\n", len(buckets))
		
		// 检查存储桶是否存在
		exists, err := minioClient.BucketExists(ctx, "h49hpg7e-blue-note")
		if err != nil {
			fmt.Printf("检查存储桶是否存在失败: %v\n", err)
			return
		}
		
		if !exists {
			fmt.Printf("存储桶不存在，尝试创建...\n")
			// 创建存储桶和设置策略的代码...
		} else {
			fmt.Printf("存储桶已存在\n")
		}
	}()
	
	return service, nil
}

// UploadFile 上传文件，带有故障转移机制
func (s *ObjectStorageService) UploadFile(fileReader io.Reader, objectName, contentType string) (string, error) {
	// 尝试上传到对象存储
	for retry := 0; retry < 3; retry++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		_, err := s.client.PutObject(ctx, s.bucketName, objectName, fileReader, -1, minio.PutObjectOptions{
			ContentType: contentType,
		})
		
		if err == nil {
			// 上传成功
			return fmt.Sprintf("https://%s/%s/%s", s.externalEndpoint, s.bucketName, objectName), nil
		}
		
		log.Printf("上传到对象存储失败 (尝试 %d/3): %v", retry+1, err)
		
		// 如果不是最后一次重试，重置reader（如果可能）
		if retry < 2 {
			if seeker, ok := fileReader.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				// 如果reader不支持seek，无法重试
				break
			}
			time.Sleep(time.Duration(retry+1) * time.Second) // 指数退避
		}
	}
	
	// 所有重试都失败，使用本地存储作为备用
	return s.uploadToLocalStorage(fileReader, objectName, contentType)
}

// 本地存储备用方案
func (s *ObjectStorageService) uploadToLocalStorage(fileReader io.Reader, objectName, contentType string) (string, error) {
	// 确保本地存储目录存在
	uploadDir := "./uploads"
	os.MkdirAll(filepath.Dir(filepath.Join(uploadDir, objectName)), 0755)
	
	// 创建本地文件
	localPath := filepath.Join(uploadDir, objectName)
	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer out.Close()
	
	// 复制数据到本地文件
	_, err = io.Copy(out, fileReader)
	if err != nil {
		return "", fmt.Errorf("写入本地文件失败: %w", err)
	}
	
	// 返回本地文件URL
	// 注意：这需要配置一个静态文件服务器来提供这些文件
	return fmt.Sprintf("/uploads/%s", objectName), nil
}

func (s *ObjectStorageService) DeleteFile(fileURL string) error {
	// 在开发环境中，如果客户端为空，直接返回
	if s.client == nil {
		return nil
	}

	// 从URL中提取对象名称
	// URL格式: https://externalEndpoint/bucketName/objectName
	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 {
		return fmt.Errorf("无效的文件URL: %s", fileURL)
	}

	objectName := strings.Join(parts[4:], "/")
	
	// 删除对象
	err := s.client.RemoveObject(
		context.Background(),
		s.bucketName,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("删除对象失败: %w", err)
	}

	return nil
}

// GenerateAvatarPath 生成头像文件路径
func GenerateAvatarPath(userID string) string {
	// 确保这里只返回相对路径，不包含存储桶名称
	return fmt.Sprintf("avatars/%s.jpg", userID)
}

// 生成帖子图片的存储路径
func GeneratePostImagePath(postID string, imageIndex int) string {
	return fmt.Sprintf("posts/%s/%d.jpg", postID, imageIndex)
}

// 生成广告图片的存储路径
func GenerateAdImagePath(adType string, adID string) string {
	return fmt.Sprintf("ads/%s/%s.jpg", adType, adID)
}

// 上传用户头像
func (s *ObjectStorageService) UploadAvatar(userID string, file io.Reader) (string, error) {
	if file == nil {
		return "", fmt.Errorf("文件为空")
	}
	
	path := fmt.Sprintf("avatars/%s.jpg", userID)
	fmt.Printf("上传头像: userID=%s, path=%s\n", userID, path)
	
	return s.UploadFile(file, path, "image/jpeg")
}

// 上传帖子图片
func (s *ObjectStorageService) UploadPostImage(postID string, imageIndex int, file io.Reader) (string, error) {
	path := GeneratePostImagePath(postID, imageIndex)
	return s.UploadFile(file, path, "image/jpeg")
}

// 上传广告图片
func (s *ObjectStorageService) UploadAdImage(adType string, adID string, file io.Reader) (string, error) {
	path := GenerateAdImagePath(adType, adID)
	return s.UploadFile(file, path, "image/jpeg")
}

// NewDegradedObjectStorageService 创建一个降级模式的对象存储服务
func NewDegradedObjectStorageService() *ObjectStorageService {
	cfg := config.GetConfig().ObjectStorage
	return &ObjectStorageService{
		client:           nil,
		bucketName:       "h49hpg7e-blue-note",
		internalEndpoint: cfg.InternalEndpoint,
		externalEndpoint: cfg.ExternalEndpoint,
	}
}

// IsAvailable 检查对象存储服务是否可用
func (s *ObjectStorageService) IsAvailable() bool {
	return s.client != nil
}

// GetExternalEndpoint 获取外部端点
func (s *ObjectStorageService) GetExternalEndpoint() string {
	return s.externalEndpoint
}

// GetBucketName 获取存储桶名称
func (s *ObjectStorageService) GetBucketName() string {
	return s.bucketName
}

// 启动健康检查
func (s *ObjectStorageService) StartHealthCheck() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, err := s.client.ListBuckets(ctx)
			cancel()
			
			if err != nil {
				log.Printf("对象存储健康检查失败: %v", err)
				// 可以在这里添加告警逻辑
			}
		}
	}()
}

// GetFileURL 获取文件URL
func (s *ObjectStorageService) GetFileURL(filePath string) string {
	return fmt.Sprintf("https://%s/%s/%s", 
		s.externalEndpoint,
		s.bucketName,
		filePath)
} 