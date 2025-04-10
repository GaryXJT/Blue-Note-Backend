package router

import (
	"blue-note/controller"
	"blue-note/middleware"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(
	authController *controller.AuthController,
	profileController *controller.ProfileController,
	postController *controller.PostController,
	adminController *controller.AdminController,
	uploadController *controller.UploadController,
	fileController *controller.FileController,
	mongoClient *mongo.Client,
) *gin.Engine {
	r := gin.Default()

	// 配置 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000", // 本地开发环境
			"https://blue-note-v443.vercel.app",
			"https://blue-note-v443-git-master-gary-xiongs-projects.vercel.app",
			"https://blue-note-v443-rmmfv3wi4-gary-xiongs-projects.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 配置静态文件服务
	// 确保上传目录存在
	os.MkdirAll("./uploads", 0755)
	r.Static("/uploads", "./uploads")
	
	// 添加日志，帮助调试静态文件服务
	fmt.Println("已配置静态文件服务: /uploads -> ./uploads")

	// 公开路由
	public := r.Group("/api/v1")
	{
		// 认证相关
		auth := public.Group("/auth")
		{
			auth.GET("/captcha", authController.GetCaptcha)
			auth.POST("/login", authController.Login)
			auth.POST("/change-password", authController.ChangePassword)
		}

		// 帖子相关（公开）
		posts := public.Group("/posts")
		{
			posts.GET("", postController.GetPostList)
			posts.GET("/:postId", postController.GetPostDetail)
		}
	}

	// 需要认证的路由
	authorized := r.Group("/api/v1")
	authorized.Use(middleware.AuthMiddleware())
	{
		// 用户相关
		userGroup := authorized.Group("/users")
		{
			// 获取用户资料
			userGroup.GET("/profile/:userId", profileController.GetProfile)
			
			// 单独的路由组，用于需要登录的操作
			authUserGroup := userGroup.Group("")
			{
				// 更新个人资料（包含头像上传）
				authUserGroup.PUT("/profile", profileController.UpdateProfile)
				
				// 关注用户
				authUserGroup.POST("/follow/:userId", profileController.FollowUser)
				
				// 取消关注
				authUserGroup.DELETE("/follow/:userId", profileController.UnfollowUser)
				
				// 检查关注状态
				authUserGroup.GET("/follow/check/:userId", profileController.CheckFollowStatus)
			}
			
			// 获取用户关注列表
			userGroup.GET("/:userId/following", profileController.GetFollowingList)
			
			// 获取用户粉丝列表
			userGroup.GET("/:userId/fans", profileController.GetFansList)
			
			// 获取用户喜欢的笔记
			userGroup.GET("/:userId/likes", profileController.GetUserLikedPosts)
			
			// 获取用户收藏的笔记
			userGroup.GET("/:userId/collections", profileController.GetUserCollectedPosts)
		}

		// 帖子相关
		posts := authorized.Group("/posts")
		{
			posts.POST("", postController.CreatePost)
			posts.PUT("/:postId", postController.UpdatePost)
			posts.DELETE("/:postId", postController.DeletePost)

			// 评论相关
			posts.GET("/:postId/comments", postController.GetPostComments)
			posts.POST("/:postId/comments", postController.CreateComment)
			posts.DELETE("/:postId/comments/:commentId", postController.DeleteComment)
			posts.POST("/:postId/comments/:commentId/like", postController.LikeComment)
			posts.DELETE("/:postId/comments/:commentId/like", postController.UnlikeComment)

			// 点赞相关
			posts.POST("/:postId/like", postController.LikePost)
			posts.DELETE("/:postId/like", postController.UnlikePost)
			posts.GET("/:postId/like", postController.CheckLikeStatus)

			// 草稿相关
			posts.POST("/draft", postController.SaveDraft)
			posts.GET("/drafts", postController.GetUserDrafts)
			posts.GET("/draft/:draftId", postController.GetDraftByID)
			posts.DELETE("/draft/:draftId", postController.DeleteDraft)
			posts.POST("/draft/:draftId/publish", postController.PublishDraft)
		}

		// 文件上传
		authorized.POST("/upload", uploadController.UploadFile)

		// 管理员相关
		admin := authorized.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/stats", adminController.GetStatistics)
			admin.GET("/posts/pending", adminController.GetPendingPosts)
			admin.PUT("/posts/:postId/review", postController.ReviewPost)
		}

		// 创建速率限制器
		rateLimiter := middleware.NewRateLimiter()

		// 文件管理相关路由
		files := authorized.Group("/file")
		{
			// 保留删除文件的路由
			files.POST("/delete", rateLimiter.RateLimit(10, time.Minute), fileController.DeleteFile)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		// 检查MongoDB连接
		mongoStatus := "ok"
		if err := mongoClient.Ping(context.Background(), nil); err != nil {
			mongoStatus = "error: " + err.Error()
		}
		
		// 检查Redis连接（如果您使用了Redis客户端）
		redisStatus := "ok"
		// 假设您有一个redisClient
		// if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		//     redisStatus = "error: " + err.Error()
		// }
		
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
			"database": gin.H{
				"mongodb": mongoStatus,
				"redis":   redisStatus,
			},
		})
	})

	return r
}
