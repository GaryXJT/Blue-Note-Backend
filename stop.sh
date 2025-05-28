#!/bin/bash

# Blue Note Backend 停止脚本

echo "正在停止 Blue Note Backend..."

# 检查PID文件是否存在
if [ -f "logs/app.pid" ]; then
    PID=$(cat logs/app.pid)
    
    # 检查进程是否还在运行
    if ps -p $PID > /dev/null 2>&1; then
        echo "正在停止进程 $PID..."
        kill $PID
        
        # 等待进程优雅退出
        sleep 3
        
        # 如果进程仍在运行，强制杀死
        if ps -p $PID > /dev/null 2>&1; then
            echo "进程未响应，强制停止..."
            kill -9 $PID
        fi
        
        echo "Blue Note Backend 已停止"
    else
        echo "进程 $PID 未运行"
    fi
    
    # 删除PID文件
    rm -f logs/app.pid
else
    echo "未找到PID文件，尝试通过进程名停止..."
    
    # 通过进程名查找并停止
    if pgrep -f "./main" > /dev/null; then
        pkill -f "./main"
        echo "Blue Note Backend 已停止"
    else
        echo "未找到运行中的 Blue Note Backend 进程"
    fi
fi 