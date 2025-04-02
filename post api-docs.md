# 蓝调笔记 API 文档

## 目录

1. [文件上传](#1-文件上传)
2. [创建笔记](#2-创建笔记)
3. [草稿管理](#3-草稿管理)
   - [保存草稿](#31-保存草稿)
   - [获取草稿列表](#32-获取草稿列表)
   - [获取草稿详情](#33-获取草稿详情)
   - [删除草稿](#34-删除草稿)
   - [发布草稿](#35-发布草稿)

## 1. 文件上传

上传图片或视频文件。

### 请求

```
POST /api/v1/upload
```

**请求头**
```
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

**表单参数**

| 参数名 | 类型   | 必填 | 描述                      |
|--------|--------|------|---------------------------|
| file   | File   | 是   | 要上传的文件              |
| type   | String | 是   | 文件类型，"image"或"video" |

**文件限制**
- 图片：最大10MB，支持JPG、PNG、GIF、WebP格式
- 视频：最大100MB，支持MP4、WebM、MOV、AVI格式

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "文件上传成功",
  "data": {
    "url": "https://example.com/image/2023/05/01/filename.jpg",
    "type": "image",
    "size": 1024000,
    "name": "filename.jpg"
  }
}
```

**错误响应**
```json
{
  "code": 40003,
  "message": "请求参数错误",
  "error": "未找到文件"
}
```

```json
{
  "code": 40004,
  "message": "文件大小超过限制(10MB)"
}
```

```json
{
  "code": 40005,
  "message": "不支持的图片格式，请上传JPG、PNG、GIF或WebP格式"
}
```

## 2. 创建笔记

创建并发布新笔记。

### 请求

```
POST /api/v1/posts
```

**请求头**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**请求体**
```json
{
  "title": "笔记标题",
  "content": "笔记内容...",
  "type": "image",
  "tags": ["标签1", "标签2"],
  "files": [
    "https://example.com/image/2023/05/01/image1.jpg",
    "https://example.com/image/2023/05/01/image2.jpg"
  ],
  "isDraft": false
}
```

**参数说明**

| 参数名  | 类型    | 必填 | 描述                                |
|---------|---------|------|-------------------------------------|
| title   | String  | 是   | 笔记标题，最大长度100               |
| content | String  | 是   | 笔记内容                            |
| type    | String  | 是   | 笔记类型，"image"或"video"          |
| tags    | Array   | 是   | 标签列表                            |
| files   | Array   | 是   | 文件URL列表，由文件上传接口返回     |
| isDraft | Boolean | 否   | 是否保存为草稿，默认false（发布）   |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "postId": "60d21b4667d0d8992e610c85",
    "title": "笔记标题",
    "content": "笔记内容...",
    "type": "image",
    "tags": ["标签1", "标签2"],
    "files": [
      "https://example.com/image/2023/05/01/image1.jpg",
      "https://example.com/image/2023/05/01/image2.jpg"
    ],
    "status": "pending",
    "userId": "60d21b4667d0d8992e610c86",
    "username": "user123",
    "nickname": "用户昵称",
    "avatar": "https://example.com/avatar.jpg",
    "likes": 0,
    "comments": 0,
    "createdAt": "2023-05-01T12:00:00Z",
    "updatedAt": "2023-05-01T12:00:00Z"
  }
}
```

**错误响应**
```json
{
  "code": 400,
  "message": "请求参数错误"
}
```

```json
{
  "code": 500,
  "message": "创建帖子失败"
}
```

## 3. 草稿管理

### 3.1 保存草稿

保存笔记草稿。

### 请求

```
POST /api/v1/posts/draft
```

**请求头**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**URL参数**

| 参数名  | 类型   | 必填 | 描述                     |
|---------|--------|------|--------------------------|
| draftId | String | 否   | 草稿ID，更新草稿时需提供 |

**请求体**
```json
{
  "title": "草稿标题",
  "content": "草稿内容...",
  "type": "image",
  "tags": ["标签1", "标签2"],
  "files": [
    "https://example.com/image/2023/05/01/image1.jpg"
  ]
}
```

**参数说明**

| 参数名  | 类型   | 必填 | 描述                            |
|---------|--------|------|----------------------------------|
| title   | String | 否   | 草稿标题，最大长度100           |
| content | String | 否   | 草稿内容                        |
| type    | String | 是   | 笔记类型，"image"或"video"      |
| tags    | Array  | 否   | 标签列表                        |
| files   | Array  | 否   | 文件URL列表，由文件上传接口返回 |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "草稿保存成功",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "title": "草稿标题",
    "content": "草稿内容...",
    "type": "image",
    "tags": ["标签1", "标签2"],
    "files": [
      "https://example.com/image/2023/05/01/image1.jpg"
    ],
    "status": "draft",
    "userId": "60d21b4667d0d8992e610c86",
    "username": "user123",
    "nickname": "用户昵称",
    "avatar": "https://example.com/avatar.jpg",
    "createdAt": "2023-05-01T12:00:00Z",
    "updatedAt": "2023-05-01T12:00:00Z"
  }
}
```

**错误响应**
```json
{
  "code": 40003,
  "message": "请求参数错误",
  "error": "错误详情"
}
```

```json
{
  "code": 50001,
  "message": "保存草稿失败",
  "error": "错误详情"
}
```

### 3.2 获取草稿列表

获取当前用户的草稿列表。

### 请求

```
GET /api/v1/posts/drafts
```

**请求头**
```
Authorization: Bearer {token}
```

**URL参数**

| 参数名 | 类型 | 必填 | 描述                   |
|--------|------|------|------------------------|
| page   | Int  | 否   | 页码，默认1            |
| limit  | Int  | 否   | 每页数量，默认10，最大100 |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "id": "60d21b4667d0d8992e610c85",
        "title": "草稿标题1",
        "content": "草稿内容1...",
        "type": "image",
        "tags": ["标签1", "标签2"],
        "files": ["https://example.com/image1.jpg"],
        "createdAt": "2023-05-01T12:00:00Z",
        "updatedAt": "2023-05-01T12:00:00Z"
      },
      {
        "id": "60d21b4667d0d8992e610c86",
        "title": "草稿标题2",
        "content": "草稿内容2...",
        "type": "video",
        "tags": ["标签3"],
        "files": ["https://example.com/video1.mp4"],
        "createdAt": "2023-05-01T11:00:00Z",
        "updatedAt": "2023-05-01T11:00:00Z"
      }
    ]
  }
}
```

**错误响应**
```json
{
  "code": 50001,
  "message": "获取草稿列表失败",
  "error": "错误详情"
}
```

### 3.3 获取草稿详情

获取指定草稿的详细信息。

### 请求

```
GET /api/v1/posts/draft/{draftId}
```

**请求头**
```
Authorization: Bearer {token}
```

**URL参数**

| 参数名  | 类型   | 必填 | 描述   |
|---------|--------|------|--------|
| draftId | String | 是   | 草稿ID |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "title": "草稿标题",
    "content": "草稿内容...",
    "type": "image",
    "tags": ["标签1", "标签2"],
    "files": [
      "https://example.com/image/2023/05/01/image1.jpg"
    ],
    "status": "draft",
    "userId": "60d21b4667d0d8992e610c86",
    "username": "user123",
    "nickname": "用户昵称",
    "avatar": "https://example.com/avatar.jpg",
    "createdAt": "2023-05-01T12:00:00Z",
    "updatedAt": "2023-05-01T12:00:00Z"
  }
}
```

**错误响应**
```json
{
  "code": 50001,
  "message": "获取草稿详情失败",
  "error": "草稿不存在或不属于当前用户"
}
```

### 3.4 删除草稿

删除指定的草稿。

### 请求

```
DELETE /api/v1/posts/draft/{draftId}
```

**请求头**
```
Authorization: Bearer {token}
```

**URL参数**

| 参数名  | 类型   | 必填 | 描述   |
|---------|--------|------|--------|
| draftId | String | 是   | 草稿ID |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "草稿删除成功"
}
```

**错误响应**
```json
{
  "code": 50001,
  "message": "删除草稿失败",
  "error": "草稿不存在或不属于当前用户"
}
```

### 3.5 发布草稿

将草稿发布为正式笔记。

### 请求

```
POST /api/v1/posts/draft/{draftId}/publish
```

**请求头**
```
Authorization: Bearer {token}
Content-Type: application/json
```

**URL参数**

| 参数名  | 类型   | 必填 | 描述   |
|---------|--------|------|--------|
| draftId | String | 是   | 草稿ID |

**请求体**
```json
{
  "title": "最终标题",
  "content": "最终内容...",
  "tags": ["标签1", "标签2"],
  "files": [
    "https://example.com/image/2023/05/01/image1.jpg",
    "https://example.com/image/2023/05/01/image2.jpg"
  ]
}
```

**参数说明**

| 参数名  | 类型   | 必填 | 描述                            |
|---------|--------|------|----------------------------------|
| title   | String | 否   | 标题，不提供则使用草稿标题       |
| content | String | 否   | 内容，不提供则使用草稿内容       |
| tags    | Array  | 否   | 标签列表，不提供则使用草稿标签   |
| files   | Array  | 否   | 文件URL列表，不提供则使用草稿文件 |

### 响应

**成功响应 (200 OK)**
```json
{
  "code": 0,
  "message": "发布成功",
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "title": "最终标题",
    "content": "最终内容...",
    "type": "image",
    "tags": ["标签1", "标签2"],
    "files": [
      "https://example.com/image/2023/05/01/image1.jpg",
      "https://example.com/image/2023/05/01/image2.jpg"
    ],
    "status": "pending",
    "userId": "60d21b4667d0d8992e610c86",
    "username": "user123",
    "nickname": "用户昵称",
    "avatar": "https://example.com/avatar.jpg",
    "createdAt": "2023-05-01T12:00:00Z",
    "updatedAt": "2023-05-01T13:00:00Z"
  }
}
```

**错误响应**
```json
{
  "code": 40003,
  "message": "请求参数错误",
  "error": "错误详情"
}
```

```json
{
  "code": 50001,
  "message": "发布草稿失败",
  "error": "草稿不存在或不属于当前用户"
}
``` 