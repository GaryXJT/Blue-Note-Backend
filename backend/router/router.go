package router

import (
	"blue-note/controller"
	"blue-note/middleware"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authController *controller.AuthController,
	profileController *controller.ProfileController,
	postController *controller.PostController,
	adminController *controller.AdminController,
	uploadController *controller.UploadController,
	fileController *controller.FileController,
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

	// 公开路由
	public := r.Group("/api/v1")
	{
		// 认证相关
		auth := public.Group("/auth")
		{
			auth.GET("/captcha", authController.GetCaptcha)
			auth.POST("/login", authController.Login)
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
			
			// 需要登录的路由
			authUserGroup := userGroup.Use(middleware.JWTAuth())
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
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	return r
}
