#!/bin/bash

echo "测试文件上传功能..."

# 创建一个测试图片文件
echo "创建测试图片文件..."
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > test-image.png

# 获取JWT token（需要先登录）
echo "正在获取认证token..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "q",
    "password": "123456",
    "captchaId": "test",
    "captchaCode": "test"
  }')

echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "无法获取token，请检查登录信息"
    exit 1
fi

echo "获取到token: ${TOKEN:0:50}..."

# 测试文件上传
echo ""
echo "测试文件上传..."
UPLOAD_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/upload" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-image.png" \
  -F "type=image")

echo "上传响应: $UPLOAD_RESPONSE"

# 清理测试文件
rm -f test-image.png

echo ""
echo "测试完成" 