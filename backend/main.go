package main

import (
	"blue-note/config"
	"blue-note/controller"
	"blue-note/router"
	"blue-note/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 在连接MongoDB之前
	log.Printf("当前环境: %s", config.GetConfig().Environment)
	log.Printf("正在连接MongoDB: %s", config.GetConfig().MongoDB.URI)

	// 连接MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 临时使用本地MongoDB连接
	mongoURI := "mongodb://localhost:27017"
	log.Printf("正在连接MongoDB: %s", mongoURI)
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("连接MongoDB失败: %v", err)
	}

	// 测试连接
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB Ping失败: %v", err)
	}
	log.Printf("MongoDB连接成功，数据库: %s", config.GetConfig().MongoDB.Database)
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("断开MongoDB连接失败: %v", err)
		}
	}()

	// 初始化服务和控制器
	db := mongoClient.Database(config.GetConfig().MongoDB.Database)

	// 初始化对象存储服务
	fmt.Println("开始初始化对象存储服务...")
	objectStorageChan := make(chan *service.ObjectStorageService, 1)
	errorChan := make(chan error, 1)

	go func() {
		// 获取配置并传递给 NewObjectStorageService
		cfg := config.GetConfig()
		objectStorageService, err := service.NewObjectStorageService(cfg)
		if err != nil {
			errorChan <- err
			return
		}
		objectStorageChan <- objectStorageService
	}()

	var objectStorageService *service.ObjectStorageService

	// 等待对象存储服务初始化，最多等待5秒
	select {
	case err := <-errorChan:
		log.Printf("初始化对象存储服务失败: %v，将使用降级模式", err)
		objectStorageService = service.NewDegradedObjectStorageService()
	case objectStorageService = <-objectStorageChan:
		log.Println("对象存储服务初始化成功")
	case <-time.After(5 * time.Second):
		log.Println("初始化对象存储服务超时，将使用降级模式")
		objectStorageService = service.NewDegradedObjectStorageService()
	}

	// 初始化文件服务
	fileService := service.NewFileService(db, objectStorageService)

	// 创建 ProfileService
	profileService := service.NewProfileService(db, objectStorageService)

	// 创建 AuthService，传入 ProfileService
	authService := service.NewAuthService(db, profileService)

	// 其他服务
	postService := service.NewPostService(db, fileService)
	adminService := service.NewAdminService(db)

	authController := controller.NewAuthController(authService)
	profileController := controller.NewProfileController(profileService)
	postController := controller.NewPostController(postService)
	adminController := controller.NewAdminController(adminService, objectStorageService)
	uploadController := controller.NewUploadController(objectStorageService, fileService)
	fileController := controller.NewFileController(fileService)

	// 设置路由
	r := router.SetupRouter(
		authController, 
		profileController, 
		postController, 
		adminController, 
		uploadController, 
		fileController,
		mongoClient,
	)

	// 设置最大并发连接数
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.GetConfig().Server.Port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 设置全局HTTP客户端，限制最大连接数
	http.DefaultTransport.(*http.Transport).MaxIdleConns = 100
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	http.DefaultTransport.(*http.Transport).MaxConnsPerHost = 100

	// 启动服务器
	log.Printf("服务器启动在 %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
