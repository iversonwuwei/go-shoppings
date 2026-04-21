.PHONY: help init run build stop restart clean test lint fmt vet db-up db-down db-reset wire

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

BINARY_API=bin/api
BINARY_WORKER=bin/worker
MAIN_API=cmd/api/main.go
MAIN_WORKER=cmd/worker/main.go

# Docker
DC=docker-compose
DC_RUN=$(DC) run --rm
DC_EXEC=$(DC) exec

help: ## 显示帮助
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

init: ## 初始化项目（下载依赖、生成wire）
	@echo "==> 安装 Go 依赖..."
	$(GOMOD) download
	@echo "==> 生成 Wire 依赖注入..."
	$(GOCMD) generate ./...
	@echo "==> 格式化代码..."
	$(GOFMT) -s -w .
	@echo "==> 完成!"

run: ## 本地开发运行 API 服务
	@echo "==> 检查 PostgreSQL 和 Redis..."
	@$(DC) ps | grep -q "wechat-mall-postgres" || (echo "PostgreSQL 未运行，请先执行 make db-up" && exit 1)
	@$(DC) ps | grep -q "wechat-mall-redis" || (echo "Redis 未运行，请先执行 make db-up" && exit 1)
	@echo "==> 启动 API 服务 (Air 热重载)..."
	air -c .air.toml

build: ## 编译生产环境二进制
	@echo "==> 编译 API 服务..."
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_API) $(MAIN_API)
	@echo "==> 编译 Worker 服务..."
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_WORKER) $(MAIN_WORKER)
	@echo "==> 完成! 二进制文件: $(BINARY_API), $(BINARY_WORKER)"

build-dev: ## 编译开发环境二进制
	$(GOBUILD) -o $(BINARY_API) $(MAIN_API)
	$(GOBUILD) -o $(BINARY_WORKER) $(MAIN_WORKER)

test: ## 运行单元测试
	@echo "==> 运行所有测试..."
	$(GOTEST) -v -race -cover ./...

test-cover: ## 运行测试并生成覆盖率报告
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

lint: ## 运行 golangci-lint
	golangci-lint run ./...

fmt: ## 格式化代码
	$(GOFMT) -s -w .
	$(GOVET) ./...

vet: ## 代码检查
	$(GOVET) ./...

clean: ## 清理构建产物
	$(GOCLEAN)
	rm -rf $(BINARY_API) $(BINARY_WORKER) coverage.out coverage.html

db-up: ## 启动数据库和 Redis
	$(DC) up -d postgres redis
	@echo "==> 等待 PostgreSQL 就绪..."
	@sleep 10
	@$(DC) exec postgres pg_isready -U postgres -d wechat_mall_saas > /dev/null 2>&1 && echo "==> PostgreSQL 已就绪!" || echo "==> PostgreSQL 启动中..."

db-down: ## 停止数据库服务
	$(DC) down

db-reset: ## 重置数据库（删除数据重新初始化）
	@echo "==> 停止服务..."
	$(DC) down -v
	@echo "==> 删除数据卷..."
	rm -rf $(DC) -f .volumes 2>/dev/null || true
	@echo "==> 重启数据库..."
	$(DC) up -d postgres redis
	@echo "==> 等待初始化..."
	@sleep 15

psql: ## 连接 PostgreSQL
	$(DC_EXEC) postgres psql -U postgres -d wechat_mall_saas

redis-cli: ## 连接 Redis
	$(DC_EXEC) redis redis-cli

wire: ## 重新生成 Wire 依赖注入
	$(GOCMD) generate ./...

# Docker
docker-build: ## 构建 Docker 镜像
	docker build -t wechat-mall-api:latest -f Dockerfile .

docker-up: ## 启动生产环境
	$(DC) up -d

docker-down: ## 停止生产环境
	$(DC) down
