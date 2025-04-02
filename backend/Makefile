.PHONY: build run dev clean test

# 构建应用
build:
	go build -o blue-note

# 运行应用
run:
	./blue-note

# 使用 air 运行开发模式(热重载)
dev:
	air

# 清理构建文件
clean:
	rm -f blue-note
	rm -rf ./tmp

# 运行测试
test:
	go test -v ./...

# 安装依赖
install:
	go mod download
	go install github.com/air-verse/air@latest 