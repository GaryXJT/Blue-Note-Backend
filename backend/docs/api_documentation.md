# Blue Note API 文档

## 基础信息

- 基础 URL: `/api/v1`
- 所有需要认证的接口都需要在请求头中携带 `Authorization: Bearer <token>`
- 响应格式统一为：

```json
{
  "code": 200, // 状态码：200成功，400请求错误，401未认证，403无权限，404未找到，500服务器错误
  "message": "success", // 响应消息
  "data": {} // 响应数据
}
```

## 用户认证 API

### 获取验证码

- 请求方法：GET
- 路径：`/auth/captcha`
- 权限：公开
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "captchaId": "string", // 验证码ID
    "captchaImage": "string" // Base64格式的验证码图片
  }
}
```

### 用户注册

- 请求方法：POST
- 路径：`/auth/register`
- 权限：公开
- 请求体：

```json
{
  "username": "string", // 用户名（必填，3-32个字符）
  "password": "string", // 密码（必填，6-32个字符）
  "captchaId": "string", // 验证码ID（必填）
  "captchaCode": "string" // 验证码（必填）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "string", // JWT token
    "userId": "string", // 用户ID
    "username": "string" // 用户名
  }
}
```

### 用户登录

- 请求方法：POST
- 路径：`/auth/login`
- 权限：公开
- 请求体：

```json
{
  "username": "string", // 用户名（必填）
  "password": "string", // 密码（必填）
  "captchaId": "string", // 验证码ID（必填）
  "captchaCode": "string" // 验证码（必填）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "string", // JWT token
    "userId": "string", // 用户ID
    "username": "string", // 用户名
    "role": "string" // 用户角色：user/admin
  }
}
```

## 用户信息 API

### 获取用户信息

- 请求方法：GET
- 路径：`/user/profile`
- 权限：需要认证
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "userId": "string", // 用户ID
    "username": "string", // 用户名
    "nickname": "string", // 昵称
    "avatar": "string", // 头像URL
    "bio": "string", // 个人简介
    "role": "string", // 用户角色
    "createdAt": "string", // 注册时间
    "updatedAt": "string" // 更新时间
  }
}
```

### 更新用户信息

- 请求方法：PUT
- 路径：`/user/profile`
- 权限：需要认证
- 请求体：

```json
{
  "nickname": "string", // 昵称（选填，最多32个字符）
  "avatar": "string", // 头像URL（选填，必须是有效URL）
  "bio": "string" // 个人简介（选填，最多200个字符）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "userId": "string",
    "username": "string",
    "nickname": "string",
    "avatar": "string",
    "bio": "string",
    "updatedAt": "string"
  }
}
```

## 帖子管理 API

### 创建帖子

- 请求方法：POST
- 路径：`/posts`
- 权限：需要认证
- 请求体：

```json
{
  "title": "string",        // 标题（必填，最多100个字符）
  "content": "string",      // 内容（必填）
  "type": "string",         // 类型：image/video（必填）
  "tags": ["string"],       // 标签数组（必填）
  "files": ["string"],      // 文件URL数组（必填）
  "isDraft": boolean        // 是否为草稿（选填，默认false）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "postId": "string",
    "title": "string",
    "content": "string",
    "type": "string",
    "tags": ["string"],
    "files": ["string"],
    "status": "string", // 状态：draft/pending/approved/rejected
    "createdAt": "string",
    "updatedAt": "string"
  }
}
```

### 获取帖子列表

- 请求方法：GET
- 路径：`/posts`
- 权限：公开
- 查询参数：
  - page: 页码（默认 1）
  - limit: 每页数量（默认 10，最大 100）
  - type: 类型筛选（image/video）
  - tag: 标签筛选
  - status: 状态筛选（draft/pending/approved/rejected）
  - userId: 用户 ID 筛选
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "limit": 10,
    "posts": [
      {
        "postId": "string",
        "title": "string",
        "content": "string",
        "type": "string",
        "tags": ["string"],
        "files": ["string"],
        "status": "string",
        "userId": "string",
        "username": "string",
        "nickname": "string",
        "avatar": "string",
        "likes": 0,
        "comments": 0,
        "createdAt": "string",
        "updatedAt": "string"
      }
    ]
  }
}
```

### 获取帖子详情

- 请求方法：GET
- 路径：`/posts/:postId`
- 权限：公开
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "postId": "string",
    "title": "string",
    "content": "string",
    "type": "string",
    "tags": ["string"],
    "files": ["string"],
    "status": "string",
    "userId": "string",
    "username": "string",
    "nickname": "string",
    "avatar": "string",
    "likes": 0,
    "comments": [
      {
        "commentId": "string",
        "content": "string",
        "userId": "string",
        "username": "string",
        "nickname": "string",
        "avatar": "string",
        "createdAt": "string"
      }
    ],
    "createdAt": "string",
    "updatedAt": "string"
  }
}
```

### 更新帖子

- 请求方法：PUT
- 路径：`/posts/:postId`
- 权限：需要认证（作者或管理员）
- 请求体：

```json
{
  "title": "string",      // 标题（选填，最多100个字符）
  "content": "string",    // 内容（选填）
  "tags": ["string"],     // 标签数组（选填）
  "files": ["string"],    // 文件URL数组（选填）
  "isDraft": boolean      // 是否为草稿（选填）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "postId": "string",
    "title": "string",
    "content": "string",
    "type": "string",
    "tags": ["string"],
    "files": ["string"],
    "status": "string",
    "updatedAt": "string"
  }
}
```

### 删除帖子

- 请求方法：DELETE
- 路径：`/posts/:postId`
- 权限：需要认证（作者或管理员）
- 响应：

```json
{
  "code": 200,
  "message": "success"
}
```

## 帖子点赞 API

### 点赞帖子

- 请求方法：POST
- 路径：`/posts/:postId/like`
- 权限：需要认证
- 响应：

```json
{
  "code": 200,
  "message": "success"
}
```

### 取消点赞

- 请求方法：DELETE
- 路径：`/posts/:postId/like`
- 权限：需要认证
- 响应：

```json
{
  "code": 200,
  "message": "success"
}
```

### 检查点赞状态

- 请求方法：GET
- 路径：`/posts/:postId/like`
- 权限：需要认证
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "hasLiked": true
  }
}
```

## 评论 API

### 发表评论

- 请求方法：POST
- 路径：`/posts/:postId/comments`
- 权限：需要认证
- 请求体：

```json
{
  "content": "string" // 评论内容（必填，最多500个字符）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "commentId": "string",
    "content": "string",
    "userId": "string",
    "username": "string",
    "nickname": "string",
    "avatar": "string",
    "createdAt": "string"
  }
}
```

### 删除评论

- 请求方法：DELETE
- 路径：`/posts/:postId/comments/:commentId`
- 权限：需要认证（评论作者或管理员）
- 响应：

```json
{
  "code": 200,
  "message": "success"
}
```

## 文件上传 API

### 上传文件

- 请求方法：POST
- 路径：`/upload`
- 权限：需要认证
- 请求体：multipart/form-data
  - file: 文件（支持格式：jpg、jpeg、png、gif、mp4、mov、avi、webm）
  - 文件大小限制：20MB
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "url": "string" // 文件访问URL
  }
}
```

## 管理员 API

### 获取统计数据

- 请求方法：GET
- 路径：`/admin/stats`
- 权限：需要管理员权限
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalUsers": 1000, // 总用户数
    "totalPosts": 5000, // 总帖子数
    "pendingPosts": 100, // 待审核帖子数
    "totalComments": 20000, // 总评论数
    "dailyStats": [
      {
        // 每日统计数据
        "date": "string",
        "newUsers": 10,
        "newPosts": 50,
        "newComments": 200
      }
    ],
    "tagStats": [
      {
        // 标签统计
        "tag": "string",
        "count": 100
      }
    ]
  }
}
```

### 获取待审核帖子列表

- 请求方法：GET
- 路径：`/admin/posts/pending`
- 权限：需要管理员权限
- 查询参数：
  - page: 页码（默认 1）
  - limit: 每页数量（默认 10，最大 100）
- 响应：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "limit": 10,
    "posts": [
      {
        "postId": "string",
        "title": "string",
        "content": "string",
        "type": "string",
        "tags": ["string"],
        "files": ["string"],
        "userId": "string",
        "username": "string",
        "nickname": "string",
        "avatar": "string",
        "createdAt": "string"
      }
    ]
  }
}
```

### 审核帖子

- 请求方法：PUT
- 路径：`/admin/posts/:postId/review`
- 权限：需要管理员权限
- 请求体：

```json
{
  "status": "string", // approved/rejected（必填）
  "reason": "string" // 拒绝原因（当status为rejected时建议填写，最多200个字符）
}
```

- 响应：

```json
{
  "code": 200,
  "message": "success"
}
```

## 错误码说明

- 200: 成功
- 400: 请求参数错误
- 401: 未认证或认证失败
- 403: 无权限访问
- 404: 资源不存在
- 500: 服务器内部错误

## 标签定义

系统支持的标签列表：

- 推荐
- 穿搭
- 美食
- 彩妆
- 影视
- 职场
- 情感
- 家居
- 游戏
- 旅行
- 健康
- 其他 