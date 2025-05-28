#!/bin/bash

# Blue Note Backend systemd 服务安装脚本

echo "正在安装 Blue Note Backend systemd 服务..."

# 检查是否有root权限
if [ "$EUID" -ne 0 ]; then
    echo "请使用sudo运行此脚本"
    echo "sudo ./install-service.sh"
    exit 1
fi

# 复制服务文件到systemd目录
cp blue-note.service /etc/systemd/system/

# 重新加载systemd配置
systemctl daemon-reload

# 启用服务（开机自启）
systemctl enable blue-note.service

echo "服务安装完成！"
echo ""
echo "可用命令:"
echo "  启动服务: sudo systemctl start blue-note"
echo "  停止服务: sudo systemctl stop blue-note"
echo "  重启服务: sudo systemctl restart blue-note"
echo "  查看状态: sudo systemctl status blue-note"
echo "  查看日志: sudo journalctl -u blue-note -f"
echo "  禁用开机自启: sudo systemctl disable blue-note"
echo ""
echo "注意: 使用systemd服务时，请不要同时使用start.sh脚本" 