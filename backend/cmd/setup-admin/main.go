package main

import (
	"blue-note/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("正在设置第一个用户为管理员...")

	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 获取配置
	cfg := config.GetConfig()
	mongoURI := cfg.MongoDB.URI
	database := cfg.MongoDB.Database

	fmt.Printf("连接到MongoDB: %s\n", mongoURI)
	fmt.Printf("数据库: %s\n", database)

	// 连接MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(ctx)

	// 测试连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB连接测试失败: %v", err)
	}

	fmt.Println("成功连接到MongoDB")

	db := client.Database(database)

	// 查找第一个用户（按创建时间排序）
	var firstUser struct {
		ID       primitive.ObjectID `bson:"_id"`
		Username string             `bson:"username"`
		Role     string             `bson:"role"`
	}

	err = db.Collection("users").FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"created_at": 1})).Decode(&firstUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("未找到任何用户，请先注册一个用户")
			return
		}
		log.Fatalf("查询用户失败: %v", err)
	}

	fmt.Printf("找到第一个用户: %s (ID: %s)\n", firstUser.Username, firstUser.ID.Hex())

	// 检查是否已经是管理员
	if firstUser.Role == "admin" {
		fmt.Println("该用户已经是管理员，无需修改")
		return
	}

	// 更新用户角色为管理员
	update := bson.M{
		"$set": bson.M{
			"role":       "admin",
			"is_admin":   true,
			"updated_at": time.Now(),
		},
	}

	result, err := db.Collection("users").UpdateOne(ctx, bson.M{"_id": firstUser.ID}, update)
	if err != nil {
		log.Fatalf("更新用户角色失败: %v", err)
	}

	if result.ModifiedCount > 0 {
		fmt.Printf("成功将用户 %s 设置为管理员\n", firstUser.Username)

		// 创建系统通知
		notification := bson.M{
			"user_id":      firstUser.ID,
			"type":         "system",
			"title":        "您已被设为管理员",
			"content":      "恭喜！您已被授予管理员权限。现在您可以管理所有用户、帖子和评论，并对提交的帖子进行审核。",
			"related_type": "user_role",
			"is_read":      false,
			"created_at":   time.Now(),
			"updated_at":   time.Now(),
		}

		_, err = db.Collection("notifications").InsertOne(ctx, notification)
		if err != nil {
			fmt.Printf("创建通知失败: %v\n", err)
		} else {
			fmt.Println("已创建管理员权限通知")
		}
	} else {
		fmt.Println("更新失败")
	}

	fmt.Println("管理员设置完成！")
} 