#!/bin/bash

echo "测试对象存储连接..."

# 测试基本连接
echo "1. 测试基本连接:"
curl -s -w "HTTP状态码: %{http_code}, 总时间: %{time_total}s\n" \
  "https://objectstorageapi.hzh.sealos.run" -o /dev/null

# 测试存储桶访问
echo ""
echo "2. 测试存储桶访问:"
curl -s -w "HTTP状态码: %{http_code}, 总时间: %{time_total}s\n" \
  "https://objectstorageapi.hzh.sealos.run/u6holj03-blue-note" -o /dev/null

# 测试DNS解析
echo ""
echo "3. 测试DNS解析:"
nslookup objectstorageapi.hzh.sealos.run

# 测试网络延迟
echo ""
echo "4. 测试网络延迟:"
ping -c 3 objectstorageapi.hzh.sealos.run

echo ""
echo "测试完成" 