#!/bin/bash

# Blue Note Backend 状态检查脚本

echo "Blue Note Backend 状态检查"
echo "=========================="

# 检查PID文件
if [ -f "logs/app.pid" ]; then
    PID=$(cat logs/app.pid)
    echo "PID文件存在: $PID"
    
    # 检查进程是否在运行
    if ps -p $PID > /dev/null 2>&1; then
        echo "状态: 运行中 ✓"
        echo "进程ID: $PID"
        
        # 显示进程信息
        echo ""
        echo "进程详情:"
        ps -p $PID -o pid,ppid,cmd,etime,pcpu,pmem
        
        # 显示端口占用情况
        echo ""
        echo "端口占用情况:"
        netstat -tlnp 2>/dev/null | grep $PID || echo "未找到端口信息"
        
    else
        echo "状态: 已停止 ✗"
        echo "PID文件存在但进程未运行"
    fi
else
    echo "PID文件不存在"
    
    # 检查是否有相关进程在运行
    if pgrep -f "./main" > /dev/null; then
        echo "状态: 运行中 ✓ (但无PID文件)"
        echo ""
        echo "相关进程:"
        pgrep -f "./main" | xargs ps -p
    else
        echo "状态: 已停止 ✗"
    fi
fi

# 检查日志文件
echo ""
echo "日志文件状态:"
if [ -f "logs/app.log" ]; then
    LOG_SIZE=$(du -h logs/app.log | cut -f1)
    LOG_LINES=$(wc -l < logs/app.log)
    echo "日志文件: logs/app.log ($LOG_SIZE, $LOG_LINES 行)"
    echo ""
    echo "最近10行日志:"
    echo "-------------"
    tail -10 logs/app.log
else
    echo "日志文件不存在"
fi