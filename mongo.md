手动启动 MongoDB 实例（可选）
如果您不想将 MongoDB 注册为 Windows 服务，可以直接手动启动 MongoDB 实例。请按照以下步骤操作：

打开命令行（以管理员身份运行）。

进入 MongoDB 的 bin 目录：

sh
复制
编辑
cd C:\Program Files\mongodb-win32-x86_64-windows-5.0.19\bin
创建数据库存储目录（如果还没有创建）：

sh
复制
编辑
mkdir C:\data\db
启动 MongoDB 实例：

sh
复制
编辑
mongod.exe --dbpath C:\data\db
这会启动 MongoDB 并将数据库数据存储在 C:\data\db 目录下。
