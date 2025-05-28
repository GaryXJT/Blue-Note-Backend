#!/bin/bash

# Blue Note Backend 重启脚本

echo "正在重启 Blue Note Backend..."

# 停止服务
./stop.sh

# 等待一下确保完全停止
sleep 2

# 启动服务
./start.sh

echo "重启完成" 