.PHONY: all build clean test proto docker-up docker-down wire ent generate

# 项目名称
APP_NAME := marketing-platform

# 构建所有服务
all: build

# 构建秒杀服务
build-seckill:
	cd cmd/seckill && go build -o ../../bin/seckill .

# 构建拼团服务
build-groupbuy:
	cd cmd/groupbuy && go build -o ../../bin/groupbuy .

# 构建抽奖服务
build-lottery:
	cd cmd/lottery && go build -o ../../bin/lottery .

# 构建库存服务
build-stock:
	cd cmd/stock && go build -o ../../bin/stock .

# 构建网关
build-gateway:
	cd cmd/gateway && go build -o ../../bin/gateway .

# 构建所有
build: build-seckill build-groupbuy build-lottery build-stock build-gateway

# 运行测试
test:
	go test ./... -v

# 生成protobuf (需要安装 protoc)
proto:
	protoc --go_out=. --go-grpc_out=. api/seckill/v1/*.proto
	protoc --go_out=. --go-grpc_out=. api/groupbuy/v1/*.proto
	protoc --go_out=. --go-grpc_out=. api/lottery/v1/*.proto

# 生成Ent代码
ent:
	go run -mod=mod entgo.io/ent/cmd/ent generate ./internal/seckill/data/ent/schema

# 生成Wire依赖注入
wire:
	wire ./cmd/seckill/
	wire ./cmd/groupbuy/
	wire ./cmd/lottery/
	wire ./cmd/stock/
	wire ./cmd/gateway/

# 生成所有代码
generate: ent wire

# 安装工具
init:
	go install github.com/google/wire/cmd/wire@latest
	go install entgo.io/ent/cmd/ent@latest

# 启动基础环境
docker-up:
	docker-compose -f deploy/docker-compose-env.yml up -d

# 停止基础环境
docker-down:
	docker-compose -f deploy/docker-compose-env.yml down

# 清理
clean:
	rm -rf bin/

# 代码格式化
fmt:
	gofmt -w .

# 依赖整理
tidy:
	go mod tidy

# 运行秒杀服务
run-seckill:
	cd cmd/seckill && go run .

# 运行拼团服务
run-groupbuy:
	cd cmd/groupbuy && go run .

# 运行抽奖服务
run-lottery:
	cd cmd/lottery && go run .

# 帮助
help:
	@echo "Available commands:"
	@echo "  make all           - Build all services"
	@echo "  make build         - Build all services"
	@echo "  make build-seckill - Build seckill service"
	@echo "  make build-groupbuy - Build groupbuy service"
	@echo "  make build-lottery - Build lottery service"
	@echo "  make build-stock   - Build stock service"
	@echo "  make build-gateway - Build gateway"
	@echo "  make test          - Run tests"
	@echo "  make proto         - Generate protobuf code"
	@echo "  make ent           - Generate Ent code"
	@echo "  make wire          - Generate Wire DI"
	@echo "  make generate      - Generate all code (ent + wire)"
	@echo "  make init          - Install tools (wire, ent)"
	@echo "  make docker-up     - Start infrastructure"
	@echo "  make docker-down   - Stop infrastructure"
	@echo "  make fmt           - Format code"
	@echo "  make tidy          - Tidy dependencies"
