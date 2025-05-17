package service

import (
	"blue-note/model"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CommentService 评论服务
type CommentService struct {
	db *mongo.Database
}

// NewCommentService 创建评论服务
func NewCommentService(db *mongo.Database) *CommentService {
	return &CommentService{db: db}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(req *model.CreateCommentRequest, currentUser *model.User) (*model.Comment, error) {
	// 解析帖子ID
	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		return nil, errors.New("无效的帖子ID")
	}

	// 检查帖子是否存在
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("帖子不存在")
		}
		return nil, err
	}

	// 检查帖子状态是否为已审核通过
	if post.Status != "approved" {
		return nil, errors.New("该帖子未通过审核或已被删除，无法评论")
	}

	now := time.Now()
	comment := &model.Comment{
		ID:           primitive.NewObjectID(),
		PostID:       postID,
		UserID:       currentUser.ID,
		Username:     currentUser.Username,
		Nickname:     currentUser.Nickname,
		Avatar:       currentUser.Avatar,
		Content:      req.Content,
		Likes:        0,
		ChildrenCount: 0,
		Level:        0, // 默认为顶级评论
		IsAuthor:     currentUser.ID == post.UserID,
		IsAdmin:      currentUser.Role == "admin", // 假设用户模型中有Role字段
		Status:       "normal",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 处理回复逻辑
	if req.ParentID != "" {
		parentID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err != nil {
			return nil, errors.New("无效的父评论ID")
		}

		// 查找父评论
		var parentComment model.Comment
		err = s.db.Collection("comments").FindOne(context.Background(), bson.M{
			"_id": parentID,
			"post_id": postID,
			"status": "normal",
		}).Decode(&parentComment)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.New("父评论不存在或已被删除")
			}
			return nil, err
		}

		comment.ParentID = &parentID
		comment.Level = parentComment.Level + 1

		// 如果是回复其他评论，并且该评论是子评论，那么设置根评论ID为父评论的根评论ID
		if parentComment.RootID != nil {
			comment.RootID = parentComment.RootID
		} else {
			// 否则设置根评论ID为父评论ID
			comment.RootID = &parentID
		}

		// 如果指定了回复目标用户ID
		if req.ReplyToID != "" {
			replyToID, err := primitive.ObjectIDFromHex(req.ReplyToID)
			if err != nil {
				return nil, errors.New("无效的回复目标用户ID")
			}

			// 验证回复目标用户存在
			var replyToUser model.User
			err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": replyToID}).Decode(&replyToUser)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					return nil, errors.New("回复目标用户不存在")
				}
				return nil, err
			}

			comment.ReplyToID = &replyToID
			comment.ReplyToName = replyToUser.Nickname
		}

		// 更新父评论的子评论数
		_, err = s.db.Collection("comments").UpdateOne(
			context.Background(),
			bson.M{"_id": parentID},
			bson.M{"$inc": bson.M{"children_count": 1}},
		)
		if err != nil {
			return nil, err
		}

		// 如果有根评论，也更新根评论的子评论数
		if comment.RootID != nil && *comment.RootID != parentID {
			_, err = s.db.Collection("comments").UpdateOne(
				context.Background(),
				bson.M{"_id": *comment.RootID},
				bson.M{"$inc": bson.M{"children_count": 1}},
			)
			if err != nil {
				return nil, err
			}
		}
	}

	// 插入评论
	_, err = s.db.Collection("comments").InsertOne(context.Background(), comment)
	if err != nil {
		return nil, err
	}

	// 更新帖子的评论数
	_, err = s.db.Collection("posts").UpdateOne(
		context.Background(),
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"comments": 1}},
	)
	if err != nil {
		return nil, err
	}

	// 创建评论通知
	notificationTitle := "收到新的评论"
	notificationContent := fmt.Sprintf("%s 评论了你的帖子《%s》：%s", currentUser.Nickname, post.Title, req.Content)

	// 如果是回复评论，通知被回复的用户
	if comment.ReplyToID != nil {
		notificationTitle = "收到回复"
		notificationContent = fmt.Sprintf("%s 回复了你的评论：%s", currentUser.Nickname, req.Content)

		// 创建通知
		notification := model.CreateNotificationRequest{
			Type:        model.NotificationTypeComment,
			Title:       notificationTitle,
			Content:     notificationContent,
			RelatedID:   comment.ID.Hex(),
			RelatedType: "comment",
			SenderID:    currentUser.ID.Hex(),
			SenderName:  currentUser.Nickname,
			SenderAvatar: currentUser.Avatar,
		}

		// 获取通知服务
		notificationService := NewNotificationService(s.db)
		_, err = notificationService.CreateNotification(comment.ReplyToID.Hex(), &notification)
		if err != nil {
			log.Printf("创建评论回复通知失败: %v", err)
			// 不影响评论功能，继续执行
		}
	} else if comment.ParentID != nil {
		// 如果是回复评论但没有明确指定回复对象，则通知父评论作者
		var parentComment model.Comment
		err = s.db.Collection("comments").FindOne(context.Background(), bson.M{"_id": *comment.ParentID}).Decode(&parentComment)
		if err == nil {
			notificationTitle = "收到回复"
			notificationContent = fmt.Sprintf("%s 回复了你的评论：%s", currentUser.Nickname, req.Content)

			// 创建通知
			notification := model.CreateNotificationRequest{
				Type:        model.NotificationTypeComment,
				Title:       notificationTitle,
				Content:     notificationContent,
				RelatedID:   comment.ID.Hex(),
				RelatedType: "comment",
				SenderID:    currentUser.ID.Hex(),
				SenderName:  currentUser.Nickname,
				SenderAvatar: currentUser.Avatar,
			}

			// 获取通知服务
			notificationService := NewNotificationService(s.db)
			_, err = notificationService.CreateNotification(parentComment.UserID.Hex(), &notification)
			if err != nil {
				log.Printf("创建评论回复通知失败: %v", err)
				// 不影响评论功能，继续执行
			}
		}
	} else {
		// 如果是对帖子的直接评论，通知帖子作者
		notification := model.CreateNotificationRequest{
			Type:        model.NotificationTypeComment,
			Title:       notificationTitle,
			Content:     notificationContent,
			RelatedID:   comment.ID.Hex(),
			RelatedType: "comment",
			SenderID:    currentUser.ID.Hex(),
			SenderName:  currentUser.Nickname,
			SenderAvatar: currentUser.Avatar,
		}

		// 获取通知服务
		notificationService := NewNotificationService(s.db)
		_, err = notificationService.CreateNotification(post.UserID.Hex(), &notification)
		if err != nil {
			log.Printf("创建评论通知失败: %v", err)
			// 不影响评论功能，继续执行
		}
	}

	return comment, nil
}

// GetComments 获取评论列表
func (s *CommentService) GetComments(query *model.GetCommentsQuery, currentUserID string) (*model.CommentListResponse, error) {
	// 解析帖子ID
	postID, err := primitive.ObjectIDFromHex(query.PostID)
	if err != nil {
		return nil, errors.New("无效的帖子ID")
	}

	// 设置默认值
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "time"
	}
	if query.Order == "" {
		query.Order = "desc"
	}

	skip := (query.Page - 1) * query.PageSize

	// 构建基础查询条件
	filter := bson.M{
		"post_id": postID,
		"status":  "normal",
	}

	// 处理父评论ID查询
	if query.ParentID != "" {
		// 如果指定了父评论ID，查询该评论的子评论
		parentID, err := primitive.ObjectIDFromHex(query.ParentID)
		if err != nil {
			return nil, errors.New("无效的父评论ID")
		}
		filter["parent_id"] = parentID
	} else {
		// 否则查询顶级评论
		filter["level"] = 0
	}

	// 获取总数
	total, err := s.db.Collection("comments").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// 计算是否有更多数据
	hasMore := total > int64(skip+query.PageSize)

	// 构建排序条件
	var sort bson.D
	if query.SortBy == "likes" {
		sort = bson.D{{Key: "likes", Value: getSortOrder(query.Order)}}
	} else {
		// 默认按时间排序
		sort = bson.D{{Key: "created_at", Value: getSortOrder(query.Order)}}
	}

	// 查询评论
	opts := options.Find().
		SetSort(sort).
		SetSkip(int64(skip)).
		SetLimit(int64(query.PageSize))

	cursor, err := s.db.Collection("comments").Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var comments []model.Comment
	if err = cursor.All(context.Background(), &comments); err != nil {
		return nil, err
	}

	// 转换为带子评论的结构
	commentWithChildren := make([]*model.CommentWithChildren, 0, len(comments))
	for _, comment := range comments {
		// 检查当前用户是否点赞了该评论
		likedByUser := false
		if currentUserID != "" {
			currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
			if err == nil {
				count, err := s.db.Collection("comment_likes").CountDocuments(
					context.Background(),
					bson.M{
						"comment_id": comment.ID,
						"user_id":    currentUserObjectID,
					},
				)
				if err == nil {
					likedByUser = count > 0
				}
			}
		}

		// 转换为API响应格式
		cwc := &model.CommentWithChildren{
			ID:           comment.ID.Hex(),
			PostID:       comment.PostID.Hex(),
			UserID:       comment.UserID.Hex(),
			Username:     comment.Username,
			Nickname:     comment.Nickname,
			Avatar:       comment.Avatar,
			Content:      comment.Content,
			Likes:        comment.Likes,
			ChildrenCount: comment.ChildrenCount,
			Level:        comment.Level,
			IsAuthor:     comment.IsAuthor,
			IsAdmin:      comment.IsAdmin,
			Status:       comment.Status,
			CreatedAt:    comment.CreatedAt,
			UpdatedAt:    comment.UpdatedAt,
			LikedByUser:  likedByUser,
			Children:     []*model.CommentWithChildren{},
		}

		// 设置ParentID（如果有）
		if comment.ParentID != nil {
			cwc.ParentID = comment.ParentID.Hex()
		}

		// 设置RootID（如果有）
		if comment.RootID != nil {
			cwc.RootID = comment.RootID.Hex()
		}

		// 设置ReplyToID和ReplyToName（如果有）
		if comment.ReplyToID != nil {
			cwc.ReplyToID = comment.ReplyToID.Hex()
			cwc.ReplyToName = comment.ReplyToName
		}

		// 如果是顶级评论且查询的不是父评论下的子评论，则加载部分子评论
		if comment.Level == 0 && query.ParentID == "" && comment.ChildrenCount > 0 {
			// 加载前3条子评论
			childrenLimit := 3
			childrenFilter := bson.M{
				"post_id":   postID,
				"parent_id": comment.ID,
				"status":    "normal",
			}

			childrenOpts := options.Find().
				SetSort(bson.D{{Key: "created_at", Value: -1}}). // 按创建时间降序
				SetLimit(int64(childrenLimit))

			childrenCursor, err := s.db.Collection("comments").Find(context.Background(), childrenFilter, childrenOpts)
			if err == nil {
				var childComments []model.Comment
				if err = childrenCursor.All(context.Background(), &childComments); err == nil {
					for _, childComment := range childComments {
						// 检查当前用户是否点赞了该子评论
						childLikedByUser := false
						if currentUserID != "" {
							currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
							if err == nil {
								count, err := s.db.Collection("comment_likes").CountDocuments(
									context.Background(),
									bson.M{
										"comment_id": childComment.ID,
										"user_id":    currentUserObjectID,
									},
								)
								if err == nil {
									childLikedByUser = count > 0
								}
							}
						}

						// 转换子评论
						childCwc := &model.CommentWithChildren{
							ID:           childComment.ID.Hex(),
							PostID:       childComment.PostID.Hex(),
							UserID:       childComment.UserID.Hex(),
							Username:     childComment.Username,
							Nickname:     childComment.Nickname,
							Avatar:       childComment.Avatar,
							Content:      childComment.Content,
							ParentID:     childComment.ParentID.Hex(),
							Likes:        childComment.Likes,
							ChildrenCount: childComment.ChildrenCount,
							Level:        childComment.Level,
							IsAuthor:     childComment.IsAuthor,
							IsAdmin:      childComment.IsAdmin,
							Status:       childComment.Status,
							CreatedAt:    childComment.CreatedAt,
							UpdatedAt:    childComment.UpdatedAt,
							LikedByUser:  childLikedByUser,
						}

						// 设置RootID（如果有）
						if childComment.RootID != nil {
							childCwc.RootID = childComment.RootID.Hex()
						}

						// 设置ReplyToID和ReplyToName（如果有）
						if childComment.ReplyToID != nil {
							childCwc.ReplyToID = childComment.ReplyToID.Hex()
							childCwc.ReplyToName = childComment.ReplyToName
						}

						cwc.Children = append(cwc.Children, childCwc)
					}
				}
				childrenCursor.Close(context.Background())
			}
		}

		commentWithChildren = append(commentWithChildren, cwc)
	}

	return &model.CommentListResponse{
		Comments: commentWithChildren,
		Total:    int(total),
		Page:     query.Page,
		PageSize: query.PageSize,
		HasMore:  hasMore,
	}, nil
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(commentID string, userID string) error {
	// 解析评论ID
	commentObjectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return errors.New("无效的评论ID")
	}

	// 解析用户ID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 查询评论信息
	var comment model.Comment
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{"_id": commentObjectID}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("评论不存在")
		}
		return err
	}

	// 检查是否是评论作者或帖子作者或管理员
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": comment.PostID}).Decode(&post)
	if err != nil {
		return err
	}

	// 获取用户信息以检查角色
	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return err
	}

	isAdmin := user.Role == "admin" // 假设用户模型中有Role字段
	isCommentAuthor := comment.UserID == userObjectID
	isPostAuthor := post.UserID == userObjectID

	if !isCommentAuthor && !isPostAuthor && !isAdmin {
		return errors.New("无权限删除此评论")
	}

	// 采用软删除，更新评论状态为deleted
	_, err = s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": commentObjectID},
		bson.M{"$set": bson.M{
			"status":     "deleted",
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		return err
	}

	// 如果是顶级评论，更新帖子评论数
	if comment.Level == 0 {
		_, err = s.db.Collection("posts").UpdateOne(
			context.Background(),
			bson.M{"_id": comment.PostID},
			bson.M{"$inc": bson.M{"comments": -1}},
		)
		if err != nil {
			return err
		}
	}

	// 如果是子评论，更新父评论的子评论数
	if comment.ParentID != nil {
		_, err = s.db.Collection("comments").UpdateOne(
			context.Background(),
			bson.M{"_id": *comment.ParentID},
			bson.M{"$inc": bson.M{"children_count": -1}},
		)
		if err != nil {
			return err
		}
	}

	// 如果有根评论且不是父评论，更新根评论的子评论数
	if comment.RootID != nil && (comment.ParentID == nil || *comment.RootID != *comment.ParentID) {
		_, err = s.db.Collection("comments").UpdateOne(
			context.Background(),
			bson.M{"_id": *comment.RootID},
			bson.M{"$inc": bson.M{"children_count": -1}},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// LikeComment 点赞评论
func (s *CommentService) LikeComment(commentID string, userID string) error {
	// 解析评论ID
	commentObjectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return errors.New("无效的评论ID")
	}

	// 解析用户ID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 检查评论是否存在
	var comment model.Comment
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{
		"_id":    commentObjectID,
		"status": "normal",
	}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("评论不存在或已被删除")
		}
		return err
	}

	// 检查是否已经点赞
	count, err := s.db.Collection("comment_likes").CountDocuments(
		context.Background(),
		bson.M{
			"comment_id": commentObjectID,
			"user_id":    userObjectID,
		},
	)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("已经点赞过该评论")
	}

	// 创建点赞记录
	like := model.CommentLike{
		CommentID: commentObjectID,
		UserID:    userObjectID,
		CreatedAt: time.Now(),
	}

	_, err = s.db.Collection("comment_likes").InsertOne(context.Background(), like)
	if err != nil {
		return err
	}

	// 更新评论点赞数
	_, err = s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": commentObjectID},
		bson.M{"$inc": bson.M{"likes": 1}},
	)
	if err != nil {
		return err
	}

	// 创建点赞通知
	// 根据评论查找帖子信息
	var post model.Post
	err = s.db.Collection("posts").FindOne(context.Background(), bson.M{"_id": comment.PostID}).Decode(&post)
	if err == nil {
		// 查找点赞用户信息
		var user model.User
		err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
		if err == nil {
			// 创建通知
			notification := model.CreateNotificationRequest{
				Type:        model.NotificationTypeLike,
				Title:       "收到评论点赞",
				Content:     fmt.Sprintf("%s 赞了你在《%s》中的评论", user.Nickname, post.Title),
				RelatedID:   commentID,
				RelatedType: "comment",
				SenderID:    userID,
				SenderName:  user.Nickname,
				SenderAvatar: user.Avatar,
			}

			// 获取通知服务
			notificationService := NewNotificationService(s.db)
			_, err = notificationService.CreateNotification(comment.UserID.Hex(), &notification)
			if err != nil {
				log.Printf("创建评论点赞通知失败: %v", err)
				// 不影响点赞功能，继续执行
			}
		}
	}

	return nil
}

// UnlikeComment 取消点赞评论
func (s *CommentService) UnlikeComment(commentID string, userID string) error {
	// 解析评论ID
	commentObjectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return errors.New("无效的评论ID")
	}

	// 解析用户ID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("无效的用户ID")
	}

	// 检查评论是否存在
	var comment model.Comment
	err = s.db.Collection("comments").FindOne(context.Background(), bson.M{
		"_id":    commentObjectID,
		"status": "normal",
	}).Decode(&comment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("评论不存在或已被删除")
		}
		return err
	}

	// 删除点赞记录
	result, err := s.db.Collection("comment_likes").DeleteOne(
		context.Background(),
		bson.M{
			"comment_id": commentObjectID,
			"user_id":    userObjectID,
		},
	)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("未点赞过该评论")
	}

	// 更新评论点赞数
	_, err = s.db.Collection("comments").UpdateOne(
		context.Background(),
		bson.M{"_id": commentObjectID},
		bson.M{"$inc": bson.M{"likes": -1}},
	)
	if err != nil {
		return err
	}

	return nil
}

// HasLikedComment 检查用户是否已点赞评论
func (s *CommentService) HasLikedComment(commentID string, userID string) (bool, error) {
	// 解析评论ID
	commentObjectID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return false, errors.New("无效的评论ID")
	}

	// 解析用户ID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, errors.New("无效的用户ID")
	}

	// 检查是否已经点赞
	count, err := s.db.Collection("comment_likes").CountDocuments(
		context.Background(),
		bson.M{
			"comment_id": commentObjectID,
			"user_id":    userObjectID,
		},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetAllComments 获取所有评论（不分页）
func (s *CommentService) GetAllComments(postID string, currentUserID string) (*model.CommentListResponse, error) {
	// 解析帖子ID
	postObjectID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, errors.New("无效的帖子ID")
	}

	// 构建基础查询条件
	filter := bson.M{
		"post_id": postObjectID,
		"status":  "normal",
	}

	// 获取总数
	total, err := s.db.Collection("comments").CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// 查询所有评论，按时间排序
	sort := bson.D{{Key: "created_at", Value: 1}} // 按创建时间升序

	cursor, err := s.db.Collection("comments").Find(
		context.Background(), 
		filter,
		options.Find().SetSort(sort),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var comments []model.Comment
	if err = cursor.All(context.Background(), &comments); err != nil {
		return nil, err
	}

	// 构建评论树结构
	commentMap := make(map[string]*model.CommentWithChildren)
	rootComments := make([]*model.CommentWithChildren, 0)

	// 第一步：转换所有评论为API格式并创建映射
	for _, comment := range comments {
		// 检查当前用户是否点赞了该评论
		likedByUser := false
		if currentUserID != "" {
			currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
			if err == nil {
				count, err := s.db.Collection("comment_likes").CountDocuments(
					context.Background(),
					bson.M{
						"comment_id": comment.ID,
						"user_id":    currentUserObjectID,
					},
				)
				if err == nil {
					likedByUser = count > 0
				}
			}
		}

		// 转换为API响应格式
		cwc := &model.CommentWithChildren{
			ID:           comment.ID.Hex(),
			PostID:       comment.PostID.Hex(),
			UserID:       comment.UserID.Hex(),
			Username:     comment.Username,
			Nickname:     comment.Nickname,
			Avatar:       comment.Avatar,
			Content:      comment.Content,
			Likes:        comment.Likes,
			ChildrenCount: comment.ChildrenCount,
			Level:        comment.Level,
			IsAuthor:     comment.IsAuthor,
			IsAdmin:      comment.IsAdmin,
			Status:       comment.Status,
			Score:        comment.Score,
			CreatedAt:    comment.CreatedAt,
			UpdatedAt:    comment.UpdatedAt,
			LikedByUser:  likedByUser,
			Children:     []*model.CommentWithChildren{},
		}

		// 设置ParentID（如果有）
		if comment.ParentID != nil {
			cwc.ParentID = comment.ParentID.Hex()
		}

		// 设置RootID（如果有）
		if comment.RootID != nil {
			cwc.RootID = comment.RootID.Hex()
		}

		// 设置ReplyToID和ReplyToName（如果有）
		if comment.ReplyToID != nil {
			cwc.ReplyToID = comment.ReplyToID.Hex()
			cwc.ReplyToName = comment.ReplyToName
		}

		// 将评论添加到映射中
		commentMap[cwc.ID] = cwc

		// 如果是顶级评论，添加到根评论列表
		if comment.Level == 0 {
			rootComments = append(rootComments, cwc)
		}
	}

	// 第二步：构建评论树
	for _, comment := range commentMap {
		if comment.ParentID != "" {
			// 如果有父评论，将当前评论添加到父评论的子评论列表中
			if parent, exists := commentMap[comment.ParentID]; exists {
				parent.Children = append(parent.Children, comment)
			}
		}
	}

	return &model.CommentListResponse{
		Comments: rootComments,
		Total:    int(total),
		Page:     1,
		PageSize: int(total),
		HasMore:  false,
	}, nil
} 