package service

import (
	"blue-note/model"
	"context"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProfileService struct {
	db                  *mongo.Database
	objectStorageService *ObjectStorageService
}

func NewProfileService(db *mongo.Database, objectStorageService *ObjectStorageService) *ProfileService {
	return &ProfileService{
		db:                  db,
		objectStorageService: objectStorageService,
	}
}

// GetDefaultAvatarURL 获取默认头像URL
func (s *ProfileService) GetDefaultAvatarURL() string {
	return fmt.Sprintf("https://%s/%s/static/default-avatar.jpg", 
					  s.objectStorageService.GetExternalEndpoint(), 
					  s.objectStorageService.GetBucketName())
}

// GetUserProfile 获取用户资料
func (s *ProfileService) GetUserProfile(userID string, currentUserID string) (*model.ProfileResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// 如果用户没有头像，设置默认头像
	if user.Avatar == "" {
		user.Avatar = s.GetDefaultAvatarURL()
	}

	// 检查当前用户是否关注了该用户
	isFollowing := false
	if currentUserID != "" && currentUserID != userID {
		currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			count, err := s.db.Collection("user_follows").CountDocuments(
				context.Background(),
				bson.M{
					"user_id":      currentUserObjectID,
					"following_id": objectID,
				},
			)
			if err == nil && count > 0 {
				isFollowing = true
			}
		}
	}

	return &model.ProfileResponse{
		UserID:       user.ID.Hex(),
		Username:     user.Username,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Bio:          user.Bio,
		Gender:       user.Gender,
		Birthday:     user.Birthday,
		Location:     user.Location,
		Status:       user.Status,
		FollowCount:  user.FollowCount,
		FansCount:    user.FansCount,
		LikeCount:    user.LikeCount,
		CollectCount: user.CollectCount,
		PostCount:    user.PostCount,
		IsFollowing:  isFollowing,
	}, nil
}

// UpdateProfileWithAvatar 更新用户资料和头像
func (s *ProfileService) UpdateProfileWithAvatar(userID string, req *model.UpdateProfileRequest, avatarFile io.Reader) (*model.ProfileResponse, error) {
	// 验证用户ID格式
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("无效的用户ID格式3: %w", err)
	}

	update := bson.M{
		"updated_at": time.Now(),
	}

	// 处理基本资料更新
	if req.Username != "" {
		// 检查用户名是否已被占用
		count, err := s.db.Collection("users").CountDocuments(
			context.Background(),
			bson.M{
				"username": req.Username,
				"_id":      bson.M{"$ne": objectID},
			},
		)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("用户名已被占用")
		}
		update["username"] = req.Username
	}

	if req.Nickname != "" {
		update["nickname"] = req.Nickname
	}
	
	if req.Bio != "" {
		update["bio"] = req.Bio
	}
	
	if req.Gender != "" {
		update["gender"] = req.Gender
	}
	
	if req.Birthday != "" {
		update["birthday"] = req.Birthday
	}
	
	if req.Location != "" {
		update["location"] = req.Location
	}
	
	if req.Status != "" {
		update["status"] = req.Status
	}
	
	// 处理头像上传
	if avatarFile != nil {
		// 尝试上传到对象存储
		avatarURL, err := s.objectStorageService.UploadAvatar(userID, avatarFile)
		if err != nil {
			// 如果上传失败，记录错误但不中断流程
			fmt.Printf("头像上传失败: %v，将使用默认头像\n", err)
			update["avatar"] = s.GetDefaultAvatarURL()
		} else {
			update["avatar"] = avatarURL
			fmt.Printf("头像上传成功，URL: %s\n", avatarURL)
		}
	} else if req.Avatar != "" {
		update["avatar"] = req.Avatar
	}

	// 更新数据库
	_, err = s.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, err
	}

	// 获取更新后的用户信息
	return s.GetUserProfile(userID, userID)
}

// FollowUser 关注用户
func (s *ProfileService) FollowUser(userID string, followingID string) (map[string]interface{}, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	followingObjectID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		return nil, err
	}

	// 检查是否已经关注
	count, err := s.db.Collection("user_follows").CountDocuments(
		context.Background(),
		bson.M{
			"user_id":      userObjectID,
			"following_id": followingObjectID,
		},
	)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("已经关注该用户")
	}

	// 创建关注关系
	follow := model.UserFollow{
		UserID:      userObjectID,
		FollowingID: followingObjectID,
		CreatedAt:   time.Now(),
	}

	result, err := s.db.Collection("user_follows").InsertOne(context.Background(), follow)
	if err != nil {
		return nil, err
	}

	// 更新关注数和粉丝数
	_, err = s.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$inc": bson.M{"follow_count": 1}},
	)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": followingObjectID},
		bson.M{"$inc": bson.M{"fans_count": 1}},
	)
	if err != nil {
		return nil, err
	}

	// 获取更新后的关注数和粉丝数
	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	var followingUser model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": followingObjectID}).Decode(&followingUser)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"followId":       result.InsertedID.(primitive.ObjectID).Hex(),
		"followingUserId": followingID,
		"followCount":    user.FollowCount,
		"fansCount":      followingUser.FansCount,
	}, nil
}

// UnfollowUser 取消关注用户
func (s *ProfileService) UnfollowUser(userID string, followingID string) (map[string]interface{}, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	followingObjectID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		return nil, err
	}

	// 删除关注关系
	result, err := s.db.Collection("user_follows").DeleteOne(
		context.Background(),
		bson.M{
			"user_id":      userObjectID,
			"following_id": followingObjectID,
		},
	)
	if err != nil {
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("未关注该用户")
	}

	// 更新关注数和粉丝数
	_, err = s.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$inc": bson.M{"follow_count": -1}},
	)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": followingObjectID},
		bson.M{"$inc": bson.M{"fans_count": -1}},
	)
	if err != nil {
		return nil, err
	}

	// 获取更新后的关注数和粉丝数
	var user model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	var followingUser model.User
	err = s.db.Collection("users").FindOne(context.Background(), bson.M{"_id": followingObjectID}).Decode(&followingUser)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"followCount": user.FollowCount,
		"fansCount":   followingUser.FansCount,
	}, nil
}

// CheckFollowStatus 检查关注状态
func (s *ProfileService) CheckFollowStatus(userID string, followingID string) (bool, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	followingObjectID, err := primitive.ObjectIDFromHex(followingID)
	if err != nil {
		return false, err
	}

	count, err := s.db.Collection("user_follows").CountDocuments(
		context.Background(),
		bson.M{
			"user_id":      userObjectID,
			"following_id": followingObjectID,
		},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFollowingList 获取关注列表
func (s *ProfileService) GetFollowingList(userID string, currentUserID string, page, limit int) (*model.UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 查询用户关注的用户ID列表
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.Collection("user_follows").Find(
		context.Background(),
		bson.M{"user_id": userObjectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var follows []model.UserFollow
	if err = cursor.All(context.Background(), &follows); err != nil {
		return nil, err
	}

	// 获取关注用户总数
	total, err := s.db.Collection("user_follows").CountDocuments(
		context.Background(),
		bson.M{"user_id": userObjectID},
	)
	if err != nil {
		return nil, err
	}

	// 如果没有关注用户，返回空列表
	if len(follows) == 0 {
		return &model.UserListResponse{
			Total: int(total),
			List:  []model.UserListItem{},
		}, nil
	}

	// 提取被关注用户ID
	var followingIDs []primitive.ObjectID
	for _, follow := range follows {
		followingIDs = append(followingIDs, follow.FollowingID)
	}

	// 查询被关注用户信息
	userCursor, err := s.db.Collection("users").Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": followingIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer userCursor.Close(context.Background())

	var users []model.User
	if err = userCursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	// 检查当前用户是否关注了这些用户
	var currentUserFollowingIDs []primitive.ObjectID
	if currentUserID != "" && currentUserID != userID {
		currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			followCursor, err := s.db.Collection("user_follows").Find(
				context.Background(),
				bson.M{"user_id": currentUserObjectID},
			)
			if err == nil {
				defer followCursor.Close(context.Background())
				var currentUserFollows []model.UserFollow
				if err = followCursor.All(context.Background(), &currentUserFollows); err == nil {
					for _, follow := range currentUserFollows {
						currentUserFollowingIDs = append(currentUserFollowingIDs, follow.FollowingID)
					}
				}
			}
		}
	}

	// 构建响应
	var list []model.UserListItem
	for _, user := range users {
		isFollowing := false
		if currentUserID == userID {
			// 如果是查询自己的关注列表，则所有用户都是已关注的
			isFollowing = true
		} else if len(currentUserFollowingIDs) > 0 {
			// 检查当前用户是否关注了该用户
			for _, id := range currentUserFollowingIDs {
				if id == user.ID {
					isFollowing = true
					break
				}
			}
		}

		list = append(list, model.UserListItem{
			UserID:      user.ID.Hex(),
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Bio:         user.Bio,
			IsFollowing: isFollowing,
		})
	}

	return &model.UserListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetFollowersList 获取粉丝列表
func (s *ProfileService) GetFollowersList(userID string, currentUserID string, page, limit int) (*model.UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 查询关注该用户的用户ID列表
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.Collection("user_follows").Find(
		context.Background(),
		bson.M{"following_id": userObjectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var follows []model.UserFollow
	if err = cursor.All(context.Background(), &follows); err != nil {
		return nil, err
	}

	// 获取粉丝总数
	total, err := s.db.Collection("user_follows").CountDocuments(
		context.Background(),
		bson.M{"following_id": userObjectID},
	)
	if err != nil {
		return nil, err
	}

	// 如果没有粉丝，返回空列表
	if len(follows) == 0 {
		return &model.UserListResponse{
			Total: int(total),
			List:  []model.UserListItem{},
		}, nil
	}

	// 提取粉丝用户ID
	var followerIDs []primitive.ObjectID
	for _, follow := range follows {
		followerIDs = append(followerIDs, follow.UserID)
	}

	// 查询粉丝用户信息
	userCursor, err := s.db.Collection("users").Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": followerIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer userCursor.Close(context.Background())

	var users []model.User
	if err = userCursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	// 检查当前用户是否关注了这些粉丝
	var currentUserFollowingIDs []primitive.ObjectID
	if currentUserID != "" {
		currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			followCursor, err := s.db.Collection("user_follows").Find(
				context.Background(),
				bson.M{"user_id": currentUserObjectID},
			)
			if err == nil {
				defer followCursor.Close(context.Background())
				var currentUserFollows []model.UserFollow
				if err = followCursor.All(context.Background(), &currentUserFollows); err == nil {
					for _, follow := range currentUserFollows {
						currentUserFollowingIDs = append(currentUserFollowingIDs, follow.FollowingID)
					}
				}
			}
		}
	}

	// 构建响应
	var list []model.UserListItem
	for _, user := range users {
		isFollowing := false
		if len(currentUserFollowingIDs) > 0 {
			// 检查当前用户是否关注了该粉丝
			for _, id := range currentUserFollowingIDs {
				if id == user.ID {
					isFollowing = true
					break
				}
			}
		}

		list = append(list, model.UserListItem{
			UserID:      user.ID.Hex(),
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Bio:         user.Bio,
			IsFollowing: isFollowing,
		})
	}

	return &model.UserListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetUserLikedPosts 获取用户喜欢的笔记
func (s *ProfileService) GetUserLikedPosts(userID string, page, limit int) (*model.PostListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 查询用户喜欢的笔记ID
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.Collection("post_likes").Find(
		context.Background(),
		bson.M{"user_id": userObjectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var likes []struct {
		PostID    primitive.ObjectID `bson:"post_id"`
		CreatedAt time.Time          `bson:"created_at"`
	}
	if err = cursor.All(context.Background(), &likes); err != nil {
		return nil, err
	}

	// 获取喜欢的笔记总数
	total, err := s.db.Collection("post_likes").CountDocuments(
		context.Background(),
		bson.M{"user_id": userObjectID},
	)
	if err != nil {
		return nil, err
	}

	// 如果没有喜欢的笔记，返回空列表
	if len(likes) == 0 {
		return &model.PostListResponse{
			Total: int(total),
			List:  []model.PostListItem{},
		}, nil
	}

	// 提取笔记ID
	var postIDs []primitive.ObjectID
	for _, like := range likes {
		postIDs = append(postIDs, like.PostID)
	}

	// 查询笔记信息
	postCursor, err := s.db.Collection("posts").Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": postIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer postCursor.Close(context.Background())

	var posts []struct {
		ID           primitive.ObjectID `bson:"_id"`
		Title        string             `bson:"title"`
		Files        []string           `bson:"files"`
		UserID       primitive.ObjectID `bson:"user_id"`
		Username     string             `bson:"username"`
		Nickname     string             `bson:"nickname"`
		Avatar       string             `bson:"avatar"`
		Likes        int                `bson:"likes"`
		Comments     int                `bson:"comments"`
		Collections  int                `bson:"collections"`
		CreatedAt    time.Time          `bson:"created_at"`
	}
	if err = postCursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	// 构建响应
	var list []model.PostListItem
	for _, post := range posts {
		coverImage := ""
		if len(post.Files) > 0 {
			coverImage = post.Files[0]
		}

		list = append(list, model.PostListItem{
			PostID:       post.ID.Hex(),
			Title:        post.Title,
			CoverImage:   coverImage,
			LikeCount:    post.Likes,
			CommentCount: post.Comments,
			CollectCount: post.Collections,
			CreatedAt:    post.CreatedAt,
			User: struct {
				UserID   string `json:"userId"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
			}{
				UserID:   post.UserID.Hex(),
				Nickname: post.Nickname,
				Avatar:   post.Avatar,
			},
		})
	}

	return &model.PostListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetUserCollectedPosts 获取用户收藏的笔记
func (s *ProfileService) GetUserCollectedPosts(userID string, page, limit int) (*model.PostListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 查询用户收藏的笔记ID
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.Collection("post_collections").Find(
		context.Background(),
		bson.M{"user_id": userObjectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var collections []struct {
		PostID    primitive.ObjectID `bson:"post_id"`
		CreatedAt time.Time          `bson:"created_at"`
	}
	if err = cursor.All(context.Background(), &collections); err != nil {
		return nil, err
	}

	// 获取收藏的笔记总数
	total, err := s.db.Collection("post_collections").CountDocuments(
		context.Background(),
		bson.M{"user_id": userObjectID},
	)
	if err != nil {
		return nil, err
	}

	// 如果没有收藏的笔记，返回空列表
	if len(collections) == 0 {
		return &model.PostListResponse{
			Total: int(total),
			List:  []model.PostListItem{},
		}, nil
	}

	// 提取笔记ID
	var postIDs []primitive.ObjectID
	for _, collection := range collections {
		postIDs = append(postIDs, collection.PostID)
	}

	// 查询笔记信息
	postCursor, err := s.db.Collection("posts").Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": postIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer postCursor.Close(context.Background())

	var posts []struct {
		ID           primitive.ObjectID `bson:"_id"`
		Title        string             `bson:"title"`
		Files        []string           `bson:"files"`
		UserID       primitive.ObjectID `bson:"user_id"`
		Username     string             `bson:"username"`
		Nickname     string             `bson:"nickname"`
		Avatar       string             `bson:"avatar"`
		Likes        int                `bson:"likes"`
		Comments     int                `bson:"comments"`
		Collections  int                `bson:"collections"`
		CreatedAt    time.Time          `bson:"created_at"`
	}
	if err = postCursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	// 构建响应
	var list []model.PostListItem
	for _, post := range posts {
		coverImage := ""
		if len(post.Files) > 0 {
			coverImage = post.Files[0]
		}

		list = append(list, model.PostListItem{
			PostID:       post.ID.Hex(),
			Title:        post.Title,
			CoverImage:   coverImage,
			LikeCount:    post.Likes,
			CommentCount: post.Comments,
			CollectCount: post.Collections,
			CreatedAt:    post.CreatedAt,
			User: struct {
				UserID   string `json:"userId"`
				Nickname string `json:"nickname"`
				Avatar   string `json:"avatar"`
			}{
				UserID:   post.UserID.Hex(),
				Nickname: post.Nickname,
				Avatar:   post.Avatar,
			},
		})
	}

	return &model.PostListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// GetFansList 获取粉丝列表
func (s *ProfileService) GetFansList(userID string, currentUserID string, page, limit int) (*model.UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 查询关注该用户的用户ID列表
	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := s.db.Collection("user_follows").Find(
		context.Background(),
		bson.M{"following_id": userObjectID},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var follows []model.UserFollow
	if err = cursor.All(context.Background(), &follows); err != nil {
		return nil, err
	}

	// 获取粉丝总数
	total, err := s.db.Collection("user_follows").CountDocuments(
		context.Background(),
		bson.M{"following_id": userObjectID},
	)
	if err != nil {
		return nil, err
	}

	// 如果没有粉丝，返回空列表
	if len(follows) == 0 {
		return &model.UserListResponse{
			Total: int(total),
			List:  []model.UserListItem{},
		}, nil
	}

	// 提取粉丝用户ID
	var fanIDs []primitive.ObjectID
	for _, follow := range follows {
		fanIDs = append(fanIDs, follow.UserID)
	}

	// 查询粉丝用户信息
	userCursor, err := s.db.Collection("users").Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": fanIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer userCursor.Close(context.Background())

	var users []model.User
	if err = userCursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	// 检查当前用户是否关注了这些粉丝
	var followingMap map[string]bool
	if currentUserID != "" {
		followingMap = make(map[string]bool)
		currentUserObjectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			followCursor, err := s.db.Collection("user_follows").Find(
				context.Background(),
				bson.M{"user_id": currentUserObjectID, "following_id": bson.M{"$in": fanIDs}},
			)
			if err == nil {
				var currentUserFollows []model.UserFollow
				if err = followCursor.All(context.Background(), &currentUserFollows); err == nil {
					for _, follow := range currentUserFollows {
						followingMap[follow.FollowingID.Hex()] = true
					}
				}
				followCursor.Close(context.Background())
			}
		}
	}

	// 构建响应
	var list []model.UserListItem
	for _, user := range users {
		isFollowing := false
		if followingMap != nil {
			isFollowing = followingMap[user.ID.Hex()]
		}

		list = append(list, model.UserListItem{
			UserID:      user.ID.Hex(),
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Bio:         user.Bio,
			IsFollowing: isFollowing,
		})
	}

	return &model.UserListResponse{
		Total: int(total),
		List:  list,
	}, nil
}

// IsObjectStorageAvailable 检查对象存储服务是否可用
func (s *ProfileService) IsObjectStorageAvailable() bool {
	return s.objectStorageService != nil && s.objectStorageService.IsAvailable()
} 