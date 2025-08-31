.PHONY: help build run test clean deps migrate swagger docker-build docker-run

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build        - 构建项目"
	@echo "  run          - 运行项目"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  deps         - 安装依赖"
	@echo "  migrate      - 数据库迁移"
	@echo "  swagger      - 生成 Swagger 文档"
	@echo "  sls-sync     - 同步 SLS 规则到数据库"
	@echo "  sls-status   - 检查 SLS 连接状态"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-run   - 运行 Docker 容器"

# 安装依赖
deps:
	go mod download
	go mod tidy

# 构建项目
build:
	go build -o bin/sls-migrate main.go

# 运行项目
run:
	go run main.go

# 运行测试
test:
	go test ./...

# 清理构建文件
clean:
	rm -rf bin/
	go clean

# 数据库迁移
migrate:
	@echo "请确保 MySQL 服务已启动，并配置了正确的环境变量"
	@echo "然后运行: go run main.go"

# 生成 Swagger 文档
swagger:
	swag init -g main.go -o docs

# 构建 Docker 镜像
docker-build:
	docker build -t sls-migrate:latest .

# 运行 Docker 容器
docker-run:
	docker run -p 8080:8080 --env-file .env sls-migrate:latest

# 开发模式运行（自动重载）
dev:
	@echo "安装 air 工具: go install github.com/cosmtrek/air@latest"
	air

# 检查代码质量
lint:
	golangci-lint run

# 格式化代码
fmt:
	go fmt ./...
	goimports -w .

# 检查依赖安全漏洞
security:
	govulncheck ./...

# SLS 相关命令
sls-sync:
	@echo "同步 SLS 规则到数据库..."
	@curl -X POST http://localhost:8080/api/v1/sls/sync

sls-status:
	@echo "检查 SLS 连接状态..."
	@curl http://localhost:8080/api/v1/sls/status

# 测试 SLS 功能
test-sls:
	@echo "测试 SLS 功能..."
	@curl http://localhost:8080/api/v1/sls/alerts

# 测试 SLS 服务编译
test-sls-compile:
	@echo "测试 SLS 服务编译..."
	@go build -o /tmp/test-sls .
	@echo "SLS 服务编译成功！"
