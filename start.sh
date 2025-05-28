#!/bin/bash

# Blue Note Backend 启动脚本
# 使用nohup让程序在后台运行，即使关闭终端也不会停止

echo "正在启动 Blue Note Backend..."

# 检查是否已经有进程在运行
if pgrep -f "./main" > /dev/null; then
    echo "检测到程序已在运行，正在停止旧进程..."
    pkill -f "./main"
    sleep 2
fi

# 确保日志目录存在
mkdir -p logs

# 使用nohup在后台启动程序，输出重定向到日志文件
nohup ./main > logs/app.log 2>&1 &

# 获取进程ID
PID=$!
echo $PID > logs/app.pid

echo "Blue Note Backend 已启动"
echo "进程ID: $PID"
echo "日志文件: logs/app.log"
echo "PID文件: logs/app.pid"
echo ""
echo "使用以下命令查看日志:"
echo "  tail -f logs/app.log"
echo ""
echo "使用以下命令停止服务:"
echo "  ./stop.sh" 