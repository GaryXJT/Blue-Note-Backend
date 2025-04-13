package service

import (
	"blue-note/model"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostService struct {
	db          *mongo.Database
	fileService *FileService
}

// 星火API的配置
var (
	hostUrl   = "wss://spark-api.cn-huabei-1.xf-yun.com/v2.1/image"
	appid     = "d27a2420"
	apiKey    = "bf31b89875fb809c98c545d7237e8fda"
	apiSecret = "ZmExZDE3YzYzNWE0MTc0YzRhNjIwZGY4"
)

// Message 定义与星火API通信的消息结构
type Message struct {
	Role        string `json:"role"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

func NewPostService(db *mongo.Database, fileService *FileService) *PostService {
	return &PostService{db: db, fileService: fileService}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// analyzeTags 使用星火API分析帖子标签
func (s *PostService) analyzeTags(title, content, imageFile string) ([]string, error) {
	// 打印调试信息
	log.Printf("开始分析标签，标题长度: %d, 内容长度: %d, 图片文件: %s", len(title), len(content), imageFile)

	// 创建WebSocket连接
	d := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second, // 增加超时时间
	}

	authUrl := s.assembleAuthUrl(hostUrl, apiKey, apiSecret)
	log.Printf("连接星火API，URL长度: %d", len(authUrl))

	conn, resp, err := d.Dial(authUrl, nil)
	if err != nil {
		respContent := s.readResp(resp)
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		log.Printf("连接星火API失败: 状态码: %d, 响应: %s, 错误: %v",
			statusCode, respContent, err)
		return nil, fmt.Errorf("连接星火API失败: %w", err)
	} else if resp.StatusCode != 101 {
		log.Printf("WebSocket握手失败: %s", s.readResp(resp))
		return nil, fmt.Errorf("websocket握手失败: %d", resp.StatusCode)
	}
	defer conn.Close()
	log.Printf("星火API连接成功")

	// 获取图片内容
	var imageBase64 string
	if imageFile != "" {
		log.Printf("尝试获取图片文件内容: %s", imageFile)
		// 使用fileService获取图片数据
		imageData, err := s.fileService.GetFileContent(imageFile)
		if err != nil {
			log.Printf("获取图片内容失败: %v", err)
			// 图片获取失败，将只使用文本分析
		} else {
			log.Printf("成功获取图片内容，大小: %d bytes", len(imageData))

			// 检查图片尺寸
			if s.fileService.objectStorageService != nil {
				fileURL := s.fileService.objectStorageService.GetFileURL(imageFile)
				log.Printf("图片URL: %s", fileURL)

				// 检查图片格式
				fileExt := strings.ToLower(filepath.Ext(imageFile))
				if fileExt != ".jpg" && fileExt != ".jpeg" && fileExt != ".png" {
					log.Printf("警告: 图片格式可能不受星火API支持，当前格式: %s", fileExt)
					// 如果文件格式不支持，不使用图片分析
					imageFile = ""
					imageData = nil
				} else {
					// 获取图片尺寸
					width, height, err := s.fileService.GetImageDimensions(imageFile)
					if err != nil {
						log.Printf("获取图片尺寸失败: %v", err)
					} else {
						totalPixels := width * height
						log.Printf("图片尺寸: %dx%d, 像素总数: %d", width, height, totalPixels)

						// 检查是否符合星火API的要求: 50×50 < 图片总像素值 < 6000×6000
						if totalPixels <= 50*50 || totalPixels >= 6000*6000 {
							log.Printf("警告: 图片分辨率不符合星火API要求 (50×50 < 总像素 < 6000×6000)")
							// 如果图片不符合要求，不使用图片分析
							imageFile = ""
							imageData = nil
						}

						// 检查图片大小
						imageSize := float64(len(imageData)) / 1024 / 1024 // 转换为MB
						log.Printf("图片大小: %.2f MB", imageSize)
						if imageSize > 4 {
							log.Printf("警告: 图片大小(%.2fMB)超过限制(4MB)", imageSize)
							imageFile = ""
							imageData = nil
						}
					}
				}
			} else {
				log.Printf("objectStorageService为空，无法获取图片URL和尺寸")
			}

			// 如果图片通过了检查，才进行Base64编码
			if imageFile != "" && imageData != nil {
				imageBase64 = base64.StdEncoding.EncodeToString(imageData)
				log.Printf("成功获取图片内容并转换为Base64，长度: %d", len(imageBase64))
			} else {
				log.Printf("跳过图片分析，仅使用文本分析")
			}
		}
	} else {
		log.Printf("没有提供图片文件，将只使用文本分析")
	}

	// 创建融合的帖子内容和提示词
	prompt := "请分析这篇帖子属于哪些类别，以JSON数组格式返回一到三个最合适的标签，例如[\"旅行\",\"风景\"]。标签范围限定在：穿搭、美食、彩妆、影视、职场、情感、家居、游戏、旅行、风景、健身"
	combinedContent := fmt.Sprintf("标题：%s\n\n正文：%s\n\n%s", title, content, prompt)
	log.Printf("组合后的内容长度: %d", len(combinedContent))

	// 创建消息序列
	var messages []Message

	// 如果有图片，添加图片消息
	if imageBase64 != "" {
		messages = append(messages, Message{
			Role:        "user",
			Content:     imageBase64,
			ContentType: "image",
		})
		log.Printf("添加图片消息")
	}

	// 添加文本消息
	messages = append(messages, Message{
		Role:        "user",
		Content:     combinedContent,
		ContentType: "text",
	})
	log.Printf("添加文本消息，共有 %d 条消息", len(messages))

	// 日志打印消息结构（不包含实际内容，避免日志过长）
	for i, msg := range messages {
		contentType := msg.ContentType
		contentLength := len(msg.Content)
		log.Printf("消息[%d]: 角色=%s, 类型=%s, 内容长度=%d",
			i, msg.Role, contentType, contentLength)
	}

	// 发送请求到星火API
	data := map[string]interface{}{
		"header": map[string]interface{}{
			"app_id": appid,
		},
		"parameter": map[string]interface{}{
			"chat": map[string]interface{}{
				"domain":      "imagev3",
				"temperature": 0.5,
				"top_k":       4,
				"max_tokens":  2048,
			},
		},
		"payload": map[string]interface{}{
			"message": map[string]interface{}{
				"text": messages,
			},
		},
	}

	// 打印请求数据（不包含图片内容，过长）
	logData := map[string]interface{}{
		"header":    data["header"],
		"parameter": data["parameter"],
		"payload": map[string]interface{}{
			"message": map[string]interface{}{
				"text": fmt.Sprintf("%d条消息", len(messages)),
			},
		},
	}
	logDataBytes, _ := json.Marshal(logData)
	log.Printf("发送请求到星火API: %s", string(logDataBytes))

	err = conn.WriteJSON(data)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	log.Printf("请求已发送，等待响应...")

	// 获取API响应
	var answer = ""
	messageCount := 0
	for {
		messageCount++
		log.Printf("等待第%d条响应消息...", messageCount)
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("读取消息错误: %v", err)
			break
		}
		log.Printf("收到第%d条响应消息，长度: %d", messageCount, len(msg))

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Printf("解析JSON错误: %v, 原始消息: %s", err, string(msg[:min(100, len(msg))]))
			return nil, fmt.Errorf("解析JSON错误: %w", err)
		}

		// 解析数据
		header := data["header"].(map[string]interface{})
		code := header["code"].(float64)

		if code != 0 {
			// 处理payload可能为nil的情况
			errorMsg := "未知错误"
			if data["payload"] != nil {
				errorMsg = fmt.Sprintf("%v", data["payload"])
			} else if header["message"] != nil {
				errorMsg = fmt.Sprintf("错误信息: %v", header["message"])
			}

			log.Printf("请求错误代码: %v, 详细信息: %s", int(code), errorMsg)
			return nil, fmt.Errorf("API请求错误(代码:%d): %s", int(code), errorMsg)
		}

		payload := data["payload"].(map[string]interface{})
		choices := payload["choices"].(map[string]interface{})
		status := choices["status"].(float64)

		text := choices["text"].([]interface{})
		content := text[0].(map[string]interface{})["content"].(string)

		if status != 2 {
			answer += content
		} else {
			answer += content
			log.Println("收到最终结果")
			break
		}
	}

	// 尝试解析返回的标签数组
	var tags []string
	log.Printf("解析结果: %s", answer)

	if strings.Contains(answer, "[") && strings.Contains(answer, "]") {
		// 提取JSON数组部分
		start := strings.Index(answer, "[")
		end := strings.LastIndex(answer, "]") + 1
		if start >= 0 && end > start {
			jsonStr := answer[start:end]
			log.Printf("提取到JSON数组: %s", jsonStr)
			// 尝试解析JSON数组
			err := json.Unmarshal([]byte(jsonStr), &tags)
			if err != nil {
				log.Printf("解析标签数组失败: %v", err)
				// 尝试分词提取标签
				for _, tag := range []string{"穿搭", "美食", "彩妆", "影视", "职场", "情感", "家居", "游戏", "旅行", "风景", "健身"} {
					if strings.Contains(answer, tag) {
						tags = append(tags, tag)
					}
				}
			}
		}
	} else {
		// 如果没有找到数组格式，尝试提取关键词
		log.Printf("未找到JSON数组格式，尝试提取关键词")
		for _, tag := range []string{"穿搭", "美食", "彩妆", "影视", "职场", "情感", "家居", "游戏", "旅行", "风景", "健身"} {
			if strings.Contains(answer, tag) {
				tags = append(tags, tag)
			}
		}
	}

	log.Printf("提取的标签: %v", tags)

	// 验证标签是否在允许的范围内
	var validTags []string
	allowedTags := []string{"穿搭", "美食", "彩妆", "影视", "职场", "情感", "家居", "游戏", "旅行", "风景", "健身"}

	for _, tag := range tags {
		for _, allowedTag := range allowedTags {
			if tag == allowedTag {
				validTags = append(validTags, tag)
				break
			}
		}
	}

	// 如果没有有效标签，添加默认标签
	if len(validTags) == 0 {
		log.Printf("没有找到有效标签，使用默认标签")
		validTags = append(validTags, "其他")
	}

	log.Printf("最终有效标签: %v", validTags)
	return validTags, nil
}

// 创建鉴权url
func (s *PostService) assembleAuthUrl(hosturl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
		fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	//签名结果
	sha := s.HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	return callurl
}

func (s *PostService) HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

func (s *PostService) readResp(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return ""
	}
	return fmt.Sprintf("code=%d,body=%s", resp.StatusCode, string(b))
}

func (s *PostService) CreatePost(user *model.User, req *model.CreatePostRequest) (*model.Post, error) {
	post := &model.Post{
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Tags:       req.Tags,
		Files:      req.Files,
		CoverImage: req.CoverImage,
		Status:     "pending",
		UserID:     user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Avatar:     user.Avatar,
		Likes:      0,
		Comments:   0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 只为图文帖子自动设置封面图片
	if post.Type == "image" && post.CoverImage == "" && len(post.Files) > 0 {
		post.CoverImage = post.Files[0]
	}

	// 自动分析标签（如果tags为空）
	if len(post.Tags) == 0 {
		// 确定要分析的图片
		var imageToAnalyze string
		if post.Type == "video" {
			imageToAnalyze = post.CoverImage
		} else if post.Type == "image" && len(post.Files) > 0 {
			imageToAnalyze = post.Files[0]
		}

		log.Printf("开始为帖子自动分析标签")
		// 调用API分析标签
		tags, err := s.analyzeTags(post.Title, post.Content, imageToAnalyze)
		log.Printf("11111111111111111111")
		log.Printf("11111111111111111111")
		log.Printf("11111111111111111111")
		if err != nil {
			log.Printf("分析标签失败: %v，将使用默认标签", err)
			// 失败时使用默认标签，不影响创建帖子
			post.Tags = []string{"其他"}
		} else if len(tags) > 0 {
			log.Printf("成功获取标签: %v", tags)
			post.Tags = tags
		} else {
			// API返回的标签数组为空时使用默认标签
			log.Printf("未获取到有效标签，使用默认标签")
			post.Tags = []string{"其他"}
		}
	}

	if req.IsDraft {
		post.Status = "draft"
	}

	result, err := s.db.Collection("posts").InsertOne(context.Background(), post)
	if err != nil {
		return nil, err
	}

	post.ID = result.InsertedID.(primitive.ObjectID)

	// 标记文件为已使用状态
	if s.fileService != nil {
		go func() {
			err := s.fileService.MarkUsed(req.Files)
			if err != nil {
				log.Printf("标记文件为已使用状态失败: %v", err)
			}
		}()
	}

	return post, nil
}

func (s *PostService) GetPostList(query *model.PostQuery) (*model.PostListResponse, error) {
	// 设置默认值
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}

	// 构建查询条件
	filter := bson.M{}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.Tag != "" {
		filter["tags"] = query.Tag
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.UserID != "" {
		userID, err := primitive.ObjectIDFromHex(query.UserID)
		if err != nil {
			return nil, err
		}
		filter["user_id"] = userID
	}

	// 获取总数
	total, err := s.db.Collection("posts").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// 查询数据
	opts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{"created_at", -1}})

	cursor, err := s.db.Collection("posts").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []*model.Post
	if err = cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	var postItems []model.PostListItem
	for _, post := range posts {
		item := model.PostListItem{
			ID:         post.ID,
			PostID:     post.ID.Hex(),
			Title:      post.Title,
			Content:    post.Content,
			Type:       post.Type,
			Tags:       post.Tags,
			Files:      post.Files,
			CoverImage: post.CoverImage,
			UserID:     post.UserID,
			Username:   post.Username,
			Nickname:   post.Nickname,
			Avatar:     post.Avatar,
			Likes:      post.Likes,
			Comments:   post.Comments,
			CreatedAt:  post.CreatedAt,
			UpdatedAt:  post.UpdatedAt,
		}

		// 如果 CoverImage 为空，则使用第一张图片作为封面
		if item.CoverImage == "" && len(post.Files) > 0 {
			item.CoverImage = post.Files[0]
		}

		// 设置用户信息
		item.User.UserID = post.UserID.Hex()
		item.User.Nickname = post.Nickname
		item.User.Avatar = post.Avatar

		postItems = append(postItems, item)
	}

	return &model.PostListResponse{
		Total: int(total),
		List:  postItems,
	}, nil
}

func (s *PostService) GetPostDetail(postID string) (*model.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostService) UpdatePost(postID string, userID string, req *model.UpdatePostRequest) (*model.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	// 检查帖子是否存在且属于当前用户
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		return nil, err
	}

	if post.UserID.Hex() != userID {
		return nil, errors.New("无权限修改此帖子")
	}

	// 构建更新内容
	update := bson.M{
		"updated_at": time.Now(),
	}

	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Content != "" {
		update["content"] = req.Content
	}
	if req.Tags != nil {
		update["tags"] = req.Tags
	}
	if req.Files != nil {
		update["files"] = req.Files
	}
	if req.CoverImage != "" {
		update["cover_image"] = req.CoverImage
	}
	if req.IsDraft {
		update["status"] = "draft"
	} else {
		update["status"] = "pending"
	}

	// 标记文件为已使用状态
	if s.fileService != nil && len(req.Files) > 0 {
		go func() {
			err := s.fileService.MarkUsed(req.Files)
			if err != nil {
				log.Printf("标记文件为已使用状态失败: %v", err)
			}
		}()
	}

	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, err
	}

	return s.GetPostDetail(postID)
}

func (s *PostService) DeletePost(postID string, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	// 检查帖子是否存在且属于当前用户
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&post)
	if err != nil {
		return err
	}

	if post.UserID.Hex() != userID {
		return errors.New("无权限删除此帖子")
	}

	_, err = s.db.Collection("posts").DeleteOne(context.Background(), bson.M{"_id": objectID})
	return err
}

func (s *PostService) DeleteComment(postID string, commentID string, userID string) error {
	commentObjectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return err
	}

	// 检查评论是否存在且属于当前用户
	var comment model.Comment
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{"_id": commentObjectID}).Decode(&comment)
	if err != nil {
		return err
	}

	if comment.UserID.Hex() != userID {
		return errors.New("无权限删除此评论")
	}

	// 删除评论
	_, err = s.db.Collection("comments").DeleteOne(context.Background(), bson.M{"_id": commentObjectID})
	if err != nil {
		return err
	}

	// 更新帖子评论数
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": postObjectID},
		bson.M{"$inc": bson.M{"comments": -1}},
	)
	return err
}

func (s *PostService) ReviewPost(postID string, req *model.ReviewPostRequest) error {
	objectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	update := bson.M{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	if req.Reason != "" {
		update["reject_reason"] = req.Reason
	}

	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	return err
}

// 点赞帖子
func (s *PostService) LikePost(postID string, userID string) error {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 检查帖子是否存在
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": postObjectID}).Decode(&post)
	if err != nil {
		return err
	}

	// 检查是否已经点赞
	count, err := s.db.Collection("post_likes").CountDocuments(
		context.Background(),
		bson.M{
			"post_id": postObjectID,
			"user_id": userObjectID,
		},
	)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("已经点赞过此帖子")
	}

	// 创建点赞记录
	like := &model.PostLike{
		PostID:    postObjectID,
		UserID:    userObjectID,
		CreatedAt: time.Now(),
	}

	_, err = s.db.Collection("post_likes").InsertOne(context.Background(), like)
	if err != nil {
		return err
	}

	// 更新帖子点赞数
	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": postObjectID},
		bson.M{"$inc": bson.M{"likes": 1}},
	)
	return err
}

// 取消点赞
func (s *PostService) UnlikePost(postID string, userID string) error {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 检查是否已经点赞
	count, err := s.db.Collection("post_likes").CountDocuments(
		context.Background(),
		bson.M{
			"post_id": postObjectID,
			"user_id": userObjectID,
		},
	)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("尚未点赞此帖子")
	}

	// 删除点赞记录
	_, err = s.db.Collection("post_likes").DeleteOne(
		context.Background(),
		bson.M{
			"post_id": postObjectID,
			"user_id": userObjectID,
		},
	)
	if err != nil {
		return err
	}

	// 更新帖子点赞数
	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": postObjectID},
		bson.M{"$inc": bson.M{"likes": -1}},
	)
	return err
}

// 检查用户是否已点赞
func (s *PostService) HasLiked(postID string, userID string) (bool, error) {
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, err
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	count, err := s.db.Collection("post_likes").CountDocuments(
		context.Background(),
		bson.M{
			"post_id": postObjectID,
			"user_id": userObjectID,
		},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 计算评论的综合评分
func (s *PostService) calculateCommentScore(comment *model.Comment) float64 {
	// 时间衰减因子 (使用改进的对数衰减)
	// 使用 24 小时作为基准时间，让新评论在 24 小时内保持较高权重
	hours := time.Since(comment.CreatedAt).Hours()
	timeDecay := 1.0 / (1.0 + math.Log1p(hours/24.0))

	// 点赞权重 (使用对数函数处理点赞数，避免点赞数过多时权重过大)
	likeWeight := 0.4
	likeScore := math.Log1p(float64(comment.Likes)) * likeWeight

	// 作者/管理员权重 (使用指数衰减，让权重随时间逐渐降低)
	authorWeight := 0.3
	authorScore := 0.0
	if comment.IsAuthor {
		authorScore += authorWeight * math.Exp(-hours/48.0) // 48小时后权重减半
	}
	if comment.IsAdmin {
		authorScore += authorWeight * 0.5 * math.Exp(-hours/48.0)
	}

	// 时间权重 (使用 sigmoid 函数，让新评论获得更好的初始展示机会)
	timeWeight := 0.3
	timeScore := timeDecay * timeWeight * (1.0 / (1.0 + math.Exp(-hours/12.0)))

	// 计算总分
	totalScore := likeScore + authorScore + timeScore

	// 归一化处理，确保分数在合理范围内
	maxScore := 10.0
	if totalScore > maxScore {
		totalScore = maxScore
	}

	return totalScore
}

// 更新评论评分
func (s *PostService) updateCommentScore(comment *model.Comment) error {
	comment.Score = s.calculateCommentScore(comment)
	comment.UpdatedAt = time.Now()

	_, err := s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": comment.ID},
		bson.M{
			"$set": bson.M{
				"score":      comment.Score,
				"updated_at": comment.UpdatedAt,
			},
		},
	)
	return err
}

// 获取帖子评论列表（带排序）
func (s *PostService) GetPostComments(postID primitive.ObjectID, query *model.CommentQuery) ([]model.Comment, int64, error) {
	// 设置默认值
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}
	if query.SortBy == "" {
		query.SortBy = "score"
	}
	if query.Order == "" {
		query.Order = "desc"
	}

	skip := int64((query.Page - 1) * query.PageSize)

	// 构建查询条件
	filter := bson.M{"post_id": postID}

	// 获取总数
	total, err := s.db.Collection("comments").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, 0, err
	}

	// 构建排序条件
	var sort bson.D
	switch query.SortBy {
	case "time":
		sort = bson.D{{"created_at", getSortOrder(query.Order)}}
	case "likes":
		sort = bson.D{{"likes", getSortOrder(query.Order)}}
	default: // score
		sort = bson.D{{"score", getSortOrder(query.Order)}}
	}

	// 查询数据
	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(query.PageSize))

	cursor, err := s.db.Collection("comments").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())

	var comments []model.Comment
	if err = cursor.All(context.Background(), &comments); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// 获取排序顺序
func getSortOrder(order string) int {
	if order == "asc" {
		return 1
	}
	return -1
}

// 创建评论
func (s *PostService) CreateComment(postID primitive.ObjectID, userID primitive.ObjectID, content string) (*model.Comment, error) {
	// 获取帖子信息
	post, err := s.GetPostDetail(postID.Hex())
	if err != nil {
		return nil, err
	}

	// 创建评论
	comment := &model.Comment{
		PostID:    postID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Likes:     0,
		IsAuthor:  userID == post.UserID,
		IsAdmin:   false, // TODO: 从用户服务获取管理员状态
	}

	// 计算初始评分
	comment.Score = s.calculateCommentScore(comment)

	// 插入评论
	result, err := s.db.Collection("comments").InsertOne(context.Background(), comment)
	if err != nil {
		return nil, err
	}

	comment.ID = result.InsertedID.(primitive.ObjectID)

	// 更新帖子评论数
	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"comments": 1}},
	)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// 点赞评论
func (s *PostService) LikeComment(commentID primitive.ObjectID, userID primitive.ObjectID) error {
	// 检查是否已经点赞
	exists, err := s.db.Collection("comment_likes").CountDocuments(
		context.Background(),
		bson.M{
			"comment_id": commentID,
			"user_id":    userID,
		},
	)
	if err != nil {
		return err
	}

	if exists > 0 {
		return fmt.Errorf("已经点赞过了")
	}

	// 创建点赞记录
	like := &model.CommentLike{
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	_, err = s.db.Collection("comment_likes").InsertOne(context.Background(), like)
	if err != nil {
		return err
	}

	// 更新评论点赞数
	_, err = s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": commentID},
		bson.M{"$inc": bson.M{"likes": 1}},
	)
	if err != nil {
		return err
	}

	// 更新评论评分
	comment := &model.Comment{}
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{"_id": commentID}).Decode(comment)
	if err != nil {
		return err
	}

	return s.updateCommentScore(comment)
}

// 取消点赞评论
func (s *PostService) UnlikeComment(commentID primitive.ObjectID, userID primitive.ObjectID) error {
	// 删除点赞记录
	result, err := s.db.Collection("comment_likes").DeleteOne(
		context.Background(),
		bson.M{
			"comment_id": commentID,
			"user_id":    userID,
		},
	)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到点赞记录")
	}

	// 更新评论点赞数
	_, err = s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": commentID},
		bson.M{"$inc": bson.M{"likes": -1}},
	)
	if err != nil {
		return err
	}

	// 更新评论评分
	comment := &model.Comment{}
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{"_id": commentID}).Decode(comment)
	if err != nil {
		return err
	}

	return s.updateCommentScore(comment)
}

// SaveDraft 保存草稿
func (s *PostService) SaveDraft(userID string, req *model.CreatePostRequest, draftID string) (*model.Post, error) {
	// 验证用户ID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	// 获取用户信息
	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	now := time.Now()

	// 如果提供了draftID，则更新现有草稿
	if draftID != "" {
		draftObjID, err := primitive.ObjectIDFromHex(draftID)
		if err != nil {
			return nil, fmt.Errorf("无效的草稿ID: %w", err)
		}

		// 确保草稿属于当前用户
		var existingDraft model.Post
		err = s.db.Collection("posts").FindOne(
			context.Background(),
			bson.M{
				"_id":     draftObjID,
				"user_id": userObjID,
				"status":  "draft",
			},
		).Decode(&existingDraft)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("草稿不存在或不属于当前用户")
			}
			return nil, fmt.Errorf("查询草稿失败: %w", err)
		}

		// 构建更新内容
		update := bson.M{
			"title":       req.Title,
			"content":     req.Content,
			"type":        req.Type,
			"tags":        req.Tags,
			"files":       req.Files,
			"cover_image": req.CoverImage,
			"updated_at":  now,
		}

		// 只为图文帖子自动设置封面图片
		if req.Type == "image" && req.CoverImage == "" && len(req.Files) > 0 {
			update["cover_image"] = req.Files[0]
		}

		_, err = s.db.Collection("posts").UpdateOne(
			context.Background(),
			bson.M{"_id": draftObjID},
			update,
		)

		if err != nil {
			return nil, fmt.Errorf("更新草稿失败: %w", err)
		}

		// 获取更新后的草稿
		var updatedDraft model.Post
		err = s.db.Collection("posts").FindOne(
			context.Background(),
			bson.M{"_id": draftObjID},
		).Decode(&updatedDraft)

		if err != nil {
			return nil, fmt.Errorf("获取更新后的草稿失败: %w", err)
		}

		return &updatedDraft, nil
	}

	// 创建新草稿
	draft := model.Post{
		ID:         primitive.NewObjectID(),
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Tags:       req.Tags,
		Files:      req.Files,
		CoverImage: req.CoverImage,
		Status:     "draft",
		UserID:     userObjID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Avatar:     user.Avatar,
		Likes:      0,
		Comments:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 只为图文帖子自动设置封面图片
	if draft.Type == "image" && draft.CoverImage == "" && len(draft.Files) > 0 {
		draft.CoverImage = draft.Files[0]
	}

	_, err = s.db.Collection("posts").InsertOne(context.Background(), draft)
	if err != nil {
		return nil, fmt.Errorf("创建草稿失败: %w", err)
	}

	return &draft, nil
}

// GetUserDrafts 获取用户草稿列表
func (s *PostService) GetUserDrafts(userID string, page, limit int) (*model.PostListResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	skip := (page - 1) * limit

	// 查询条件
	filter := bson.M{
		"user_id": userObjID,
		"status":  "draft",
	}

	// 获取总数
	total, err := s.db.Collection("posts").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("获取草稿总数失败: %w", err)
	}

	// 获取草稿列表
	cursor, err := s.db.Collection("posts").Find(
		context.Background(),
		filter,
		options.Find().
			SetSort(bson.M{"updated_at": -1}).
			SetSkip(int64(skip)).
			SetLimit(int64(limit)),
	)

	if err != nil {
		return nil, fmt.Errorf("查询草稿列表失败: %w", err)
	}

	var drafts []model.Post
	if err = cursor.All(context.Background(), &drafts); err != nil {
		return nil, fmt.Errorf("解析草稿列表失败: %w", err)
	}

	// 转换为响应格式
	draftItems := make([]model.PostListItem, 0, len(drafts))
	for _, draft := range drafts {
		item := model.PostListItem{
			ID:         draft.ID,
			PostID:     draft.ID.Hex(),
			Title:      draft.Title,
			Content:    draft.Content,
			Type:       draft.Type,
			Tags:       draft.Tags,
			Files:      draft.Files,
			CoverImage: draft.CoverImage,
			UserID:     draft.UserID,
			Username:   draft.Username,
			Nickname:   draft.Nickname,
			Avatar:     draft.Avatar,
			CreatedAt:  draft.CreatedAt,
			UpdatedAt:  draft.UpdatedAt,
		}

		// 如果 CoverImage 为空，则使用第一张图片作为封面
		if item.CoverImage == "" && len(draft.Files) > 0 {
			item.CoverImage = draft.Files[0]
		}

		// 设置用户信息
		item.User.UserID = draft.UserID.Hex()
		item.User.Nickname = draft.Nickname
		item.User.Avatar = draft.Avatar

		draftItems = append(draftItems, item)
	}

	return &model.PostListResponse{
		Total: int(total),
		List:  draftItems,
	}, nil
}

// GetDraftByID 获取草稿详情
func (s *PostService) GetDraftByID(draftID, userID string) (*model.Post, error) {
	draftObjID, err := primitive.ObjectIDFromHex(draftID)
	if err != nil {
		return nil, fmt.Errorf("无效的草稿ID: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	// 查询草稿
	var draft model.Post
	err = s.db.Collection("posts").FindOne(
		context.Background(),
		bson.M{
			"_id":     draftObjID,
			"user_id": userObjID,
			"status":  "draft",
		},
	).Decode(&draft)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("草稿不存在或不属于当前用户")
		}
		return nil, fmt.Errorf("查询草稿失败: %w", err)
	}

	return &draft, nil
}

// DeleteDraft 删除草稿
func (s *PostService) DeleteDraft(draftID, userID string) error {
	draftObjID, err := primitive.ObjectIDFromHex(draftID)
	if err != nil {
		return fmt.Errorf("无效的草稿ID: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("无效的用户ID: %w", err)
	}

	// 删除草稿
	result, err := s.db.Collection("posts").DeleteOne(
		context.Background(),
		bson.M{
			"_id":     draftObjID,
			"user_id": userObjID,
			"status":  "draft",
		},
	)

	if err != nil {
		return fmt.Errorf("删除草稿失败: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("草稿不存在或不属于当前用户")
	}

	return nil
}

// PublishDraft 发布草稿
func (s *PostService) PublishDraft(draftID, userID string, updateReq *model.UpdatePostRequest) (*model.Post, error) {
	draftObjID, err := primitive.ObjectIDFromHex(draftID)
	if err != nil {
		return nil, fmt.Errorf("无效的草稿ID: %w", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID: %w", err)
	}

	// 查询草稿
	var draft model.Post
	err = s.db.Collection("posts").FindOne(
		context.Background(),
		bson.M{
			"_id":     draftObjID,
			"user_id": userObjID,
			"status":  "draft",
		},
	).Decode(&draft)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("草稿不存在或不属于当前用户")
		}
		return nil, fmt.Errorf("查询草稿失败: %w", err)
	}

	// 更新字段
	update := bson.M{
		"$set": bson.M{
			"status":     "pending",
			"updated_at": time.Now(),
		},
	}

	// 如果提供了更新字段，则更新
	if updateReq != nil {
		if updateReq.Title != "" {
			update["$set"].(bson.M)["title"] = updateReq.Title
		}
		if updateReq.Content != "" {
			update["$set"].(bson.M)["content"] = updateReq.Content
		}
		if len(updateReq.Tags) > 0 {
			update["$set"].(bson.M)["tags"] = updateReq.Tags
		}
		if len(updateReq.Files) > 0 {
			update["$set"].(bson.M)["files"] = updateReq.Files
		}
		if updateReq.CoverImage != "" {
			update["$set"].(bson.M)["cover_image"] = updateReq.CoverImage
		}
	}

	// 更新草稿状态为pending
	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": draftObjID},
		update,
	)

	if err != nil {
		return nil, fmt.Errorf("发布草稿失败: %w", err)
	}

	// 获取更新后的帖子
	var publishedPost model.Post
	err = s.db.Collection("posts").FindOne(
		context.Background(),
		bson.M{"_id": draftObjID},
	).Decode(&publishedPost)

	if err != nil {
		return nil, fmt.Errorf("获取发布后的帖子失败: %w", err)
	}

	// 标记文件为已使用状态
	if s.fileService != nil {
		files := publishedPost.Files
		go func() {
			err := s.fileService.MarkUsed(files)
			if err != nil {
				log.Printf("标记文件为已使用状态失败: %v", err)
			}
		}()
	}

	return &publishedPost, nil
}

// GetPostListWithCursor 获取帖子列表（基于游标的分页）
func (s *PostService) GetPostListWithCursor(query *model.CursorQuery) (*model.CursorBasedPostResponse, error) {
	// 设置默认值
	if query.Limit < 1 {
		query.Limit = 10
	}

	// 构建查询条件
	filter := bson.M{"status": "approved"} // 默认只查询已审核通过的帖子

	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.Tag != "" {
		filter["tags"] = query.Tag
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.UserID != "" {
		userID, err := primitive.ObjectIDFromHex(query.UserID)
		if err != nil {
			return nil, err
		}
		filter["user_id"] = userID
	}

	// 使用游标进行分页
	if query.Cursor != "" {
		cursorID, err := primitive.ObjectIDFromHex(query.Cursor)
		if err != nil {
			return nil, fmt.Errorf("无效的游标值: %w", err)
		}
		// 查询比当前游标ID更早的数据（按创建时间降序排序）
		filter["_id"] = bson.M{"$lt": cursorID}
	}

	// 查询数据
	opts := options.Find().
		SetLimit(int64(query.Limit + 1)). // 多查询一条数据，用于判断是否还有更多
		SetSort(bson.D{{"_id", -1}})      // 按ID降序排序，等同于按创建时间降序

	cursor, err := s.db.Collection("posts").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []*model.Post
	if err = cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	// 判断是否还有更多数据
	hasMore := false
	nextCursor := ""
	if len(posts) > query.Limit {
		hasMore = true
		posts = posts[:query.Limit] // 去掉多查询的那一条
	}

	// 准备响应数据
	var postItems []model.PostItem
	for _, post := range posts {
		// 获取封面图片的宽高信息
		width, height := 800, 600 // 默认宽高
		if post.CoverImage != "" {
			// 这里可以添加获取图片宽高的逻辑
			// 可以从文件元数据服务获取，或者用其他方式计算
			if s.fileService != nil {
				w, h, err := s.fileService.GetImageDimensions(post.CoverImage)
				if err == nil {
					width, height = w, h
				}
			}
		} else if len(post.Files) > 0 {
			// 如果没有设置封面图，使用第一张图片
			if s.fileService != nil {
				w, h, err := s.fileService.GetImageDimensions(post.Files[0])
				if err == nil {
					width, height = w, h
				}
			}
		}

		item := model.PostItem{
			ID:         post.ID.Hex(),
			Title:      post.Title,
			Content:    post.Content,
			Type:       post.Type,
			Tags:       post.Tags,
			Files:      post.Files,
			CoverImage: post.CoverImage,
			Width:      width,
			Height:     height,
			UserID:     post.UserID.Hex(),
			Username:   post.Username,
			Nickname:   post.Nickname,
			Avatar:     post.Avatar,
			Likes:      post.Likes,
			Comments:   post.Comments,
			CreatedAt:  post.CreatedAt,
		}

		// 如果封面图为空，使用第一张图片
		if item.CoverImage == "" && len(post.Files) > 0 {
			item.CoverImage = post.Files[0]
		}

		postItems = append(postItems, item)
	}

	// 设置下一页游标
	if hasMore && len(posts) > 0 {
		nextCursor = posts[len(posts)-1].ID.Hex()
	}

	return &model.CursorBasedPostResponse{
		Posts:      postItems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
