# 小蓝书个人资料 API 文档

本文档描述了小蓝书应用中与用户个人资料相关的所有 API 端点。

## 基础 URL

所有 API 请求的基础 URL 为：`/api`

## 认证

除非特别说明，所有 API 请求都需要在请求头中包含有效的 JWT 令牌：

```
Authorization: Bearer <token>
```

## API 端点

### 2. 获取指定用户资料

根据用户 ID 获取指定用户的个人资料信息。

**请求**

```
GET /users/profile/:userId
```

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "userId": "12345",
    "username": "blue_note_user",
    "nickname": "蓝色笔记",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "认真吃饭",
    "gender": "female",
    "birthday": "1995-01-01",
    "location": "北京 海淀区",
    "status": "happy",
    "followCount": 120,
    "fansCount": 230,
    "likeCount": 1500,
    "collectCount": 345,
    "postCount": 56,
    "isFollowing": false // 当前登录用户是否关注了该用户
  }
}
```

### 3. 更新个人资料

更新当前登录用户的个人资料信息。

**请求**

```
PUT /users/profile
```

**请求参数**

```json
{
  "username": "new_username",
  "nickname": "新昵称",
  "avatar": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEA...", // Base64编码的图片数据或图片URL
  "bio": "新的个人简介",
  "gender": "male", // "male", "female", "other"
  "birthday": "1990-01-01", // YYYY-MM-DD格式
  "location": "上海 浦东新区",
  "status": "excited" // "happy", "relaxed", "bored", "excited", "sad", "anxious", "peaceful", "energetic", "tired", "thinking"
}
```

**响应**

```json
{
  "code": 0,
  "message": "个人资料更新成功",
  "data": {
    "userId": "12345",
    "username": "new_username",
    "nickname": "新昵称",
    "avatar": "https://example.com/new-avatar.jpg",
    "bio": "新的个人简介",
    "gender": "male",
    "birthday": "1990-01-01",
    "location": "上海 浦东新区",
    "status": "excited"
  }
}
```

### 4. 关注用户

关注指定的用户。

**请求**

```
POST /users/follow/:userId
```

**响应**

```json
{
  "code": 0,
  "message": "关注成功",
  "data": {
    "followId": "789",
    "followingUserId": "12345", // 被关注的用户ID
    "followCount": 121, // 当前用户的关注数
    "fansCount": 231 // 被关注用户的粉丝数
  }
}
```

### 5. 取消关注用户

取消关注指定的用户。

**请求**

```
DELETE /users/follow/:userId
```

**响应**

```json
{
  "code": 0,
  "message": "已取消关注",
  "data": {
    "followCount": 120, // 当前用户的关注数
    "fansCount": 230 // 被关注用户的粉丝数
  }
}
```

### 6. 检查关注状态

检查当前登录用户是否关注了指定用户。

**请求**

```
GET /users/follow/check/:userId
```

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "isFollowing": true
  }
}
```

### 7. 获取用户的关注列表

获取指定用户的关注列表。

**请求**

```
GET /users/:userId/following
```

**查询参数**

- `page`: 页码，默认为 1
- `limit`: 每页记录数，默认为 20

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 120,
    "list": [
      {
        "userId": "67890",
        "username": "user1",
        "nickname": "用户1",
        "avatar": "https://example.com/avatar1.jpg",
        "bio": "这是用户1的简介",
        "isFollowing": true // 当前登录用户是否关注了该用户
      }
      // 更多关注用户...
    ]
  }
}
```

### 8. 获取用户的粉丝列表

获取指定用户的粉丝列表。

**请求**

```
GET /users/:userId/followers
```

**查询参数**

- `page`: 页码，默认为 1
- `limit`: 每页记录数，默认为 20

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 230,
    "list": [
      {
        "userId": "13579",
        "username": "fan1",
        "nickname": "粉丝1",
        "avatar": "https://example.com/fan1.jpg",
        "bio": "这是粉丝1的简介",
        "isFollowing": false // 当前登录用户是否关注了该用户
      }
      // 更多粉丝用户...
    ]
  }
}
```

### 9. 上传头像

单独上传用户头像。

**请求**

```
POST /users/avatar
```

**请求内容**

使用`multipart/form-data`格式，包含以下字段：

- `avatar`: 图片文件

**响应**

```json
{
  "code": 0,
  "message": "头像上传成功",
  "data": {
    "avatarUrl": "https://example.com/new-avatar.jpg"
  }
}
```

### 10. 获取用户喜欢的笔记

获取用户喜欢的笔记列表。

**请求**

```
GET /users/:userId/likes
```

**查询参数**

- `page`: 页码，默认为 1
- `limit`: 每页记录数，默认为 20

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 100,
    "list": [
      {
        "postId": "post123",
        "title": "标题",
        "coverImage": "https://example.com/cover1.jpg",
        "likeCount": 120,
        "commentCount": 30,
        "collectCount": 45,
        "createdAt": "2023-01-01T08:00:00Z",
        "user": {
          "userId": "user123",
          "nickname": "发布者",
          "avatar": "https://example.com/avatar2.jpg"
        }
      }
      // 更多笔记...
    ]
  }
}
```

### 11. 获取用户收藏的笔记

获取用户收藏的笔记列表。

**请求**

```
GET /users/:userId/collections
```

**查询参数**

- `page`: 页码，默认为 1
- `limit`: 每页记录数，默认为 20

**响应**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 80,
    "list": [
      {
        "postId": "post456",
        "title": "收藏的笔记",
        "coverImage": "https://example.com/cover2.jpg",
        "likeCount": 150,
        "commentCount": 40,
        "collectCount": 60,
        "createdAt": "2023-02-01T10:30:00Z",
        "user": {
          "userId": "user456",
          "nickname": "笔记作者",
          "avatar": "https://example.com/avatar3.jpg"
        }
      }
      // 更多笔记...
    ]
  }
}
```

## 错误码说明

| 错误码 | 描述               |
| ------ | ------------------ |
| 0      | 成功               |
| 40001  | 未授权，请先登录   |
| 40002  | 无效的用户 ID      |
| 40003  | 无效的请求参数     |
| 40004  | 用户不存在         |
| 40005  | 用户名已被占用     |
| 40006  | 上传文件格式不支持 |
| 40007  | 上传文件过大       |
| 50001  | 服务器内部错误     |
