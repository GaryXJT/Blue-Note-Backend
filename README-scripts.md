# Blue Note Backend 部署脚本说明

本项目提供了多种方式来运行和管理 Blue Note Backend 服务，适用于在 Sealos 云服务器上的部署。

## 脚本文件说明

### 1. 基础运行脚本

- **start.sh** - 启动服务（后台运行）
- **stop.sh** - 停止服务
- **restart.sh** - 重启服务
- **status.sh** - 查看服务状态

### 2. 系统服务脚本

- **blue-note.service** - systemd 服务配置文件
- **install-service.sh** - 安装 systemd 服务

## 使用方法

### 方法一：使用脚本管理（推荐用于开发和测试）

1. **给脚本添加执行权限：**
   ```bash
   chmod +x *.sh
   ```

2. **启动服务：**
   ```bash
   ./start.sh
   ```

3. **查看状态：**
   ```bash
   ./status.sh
   ```

4. **查看日志：**
   ```bash
   tail -f logs/app.log
   ```

5. **停止服务：**
   ```bash
   ./stop.sh
   ```

6. **重启服务：**
   ```bash
   ./restart.sh
   ```

### 方法二：使用 systemd 服务（推荐用于生产环境）

1. **安装服务：**
   ```bash
   sudo ./install-service.sh
   ```

2. **启动服务：**
   ```bash
   sudo systemctl start blue-note
   ```

3. **查看状态：**
   ```bash
   sudo systemctl status blue-note
   ```

4. **查看日志：**
   ```bash
   sudo journalctl -u blue-note -f
   ```

5. **停止服务：**
   ```bash
   sudo systemctl stop blue-note
   ```

6. **重启服务：**
   ```bash
   sudo systemctl restart blue-note
   ```

## 特性说明

### 脚本方式特性：
- ✅ 使用 nohup 后台运行，关闭终端不会停止服务
- ✅ 自动创建日志文件和 PID 文件
- ✅ 支持进程检查和优雅停止
- ✅ 简单易用，适合开发调试

### systemd 服务特性：
- ✅ 开机自动启动
- ✅ 进程崩溃自动重启
- ✅ 系统级别的日志管理
- ✅ 更好的资源管理和监控
- ✅ 适合生产环境

## 注意事项

1. **不要同时使用两种方式**：选择其中一种方式来管理服务，避免冲突。

2. **端口冲突检查**：确保配置的端口没有被其他服务占用。

3. **权限问题**：
   - 脚本方式：确保当前用户有执行权限
   - systemd 方式：需要 sudo 权限来安装和管理服务

4. **日志管理**：
   - 脚本方式：日志保存在 `logs/app.log`
   - systemd 方式：使用 `journalctl` 查看日志

5. **MongoDB 连接**：确保 MongoDB 服务正在运行并且连接配置正确。

## 故障排除

### 服务无法启动
1. 检查 MongoDB 是否运行：`sudo systemctl status mongod`
2. 检查端口是否被占用：`netstat -tlnp | grep :端口号`
3. 查看详细日志：`./status.sh` 或 `sudo journalctl -u blue-note`

### 进程意外停止
1. 查看日志文件找出错误原因
2. 检查系统资源使用情况：`top` 或 `htop`
3. 使用 systemd 方式可以自动重启

### 无法访问服务
1. 检查防火墙设置
2. 确认服务绑定的 IP 地址和端口
3. 检查 Sealos 的网络配置和端口映射

## 快速开始

对于新部署，推荐按以下步骤操作：

```bash
# 1. 给脚本添加执行权限
chmod +x *.sh

# 2. 启动服务
./start.sh

# 3. 检查状态
./status.sh

# 4. 如果一切正常，可以安装为系统服务
sudo ./install-service.sh
sudo systemctl start blue-note
``` 