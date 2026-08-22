# 营销中台 - Go微服务

基于DDD六边形架构的Go微服务营销中台，包含秒杀、拼团、抽奖三大核心服务。

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 框架 | Kratos v2 | B站开源微服务框架 |
| RPC | gRPC + gRPC-Gateway | 高性能RPC + HTTP兼容 |
| ORM | GORM v2 | Go ORM |
| 数据库 | MySQL 8.x | 三库隔离 |
| 缓存 | Redis 7.x | 分布式锁/Lua原子操作 |
| 消息队列 | RabbitMQ | 异步消息 |
| 注册中心 | Nacos 2.x | 服务发现+配置中心 |
| 监控 | Prometheus + Grafana | 指标监控 |
| 链路追踪 | Jaeger | OpenTelemetry |
| 容器化 | Docker Compose | 一键部署 |

## 项目结构

```
marketing-platform/
├── api/                    # Protobuf接口定义
│   ├── seckill/v1/
│   ├── groupbuy/v1/
│   └── lottery/v1/
├── cmd/                    # 各服务入口
│   ├── seckill/
│   ├── groupbuy/
│   └── lottery/
├── internal/               # 内部实现(DDD分层)
│   ├── seckill/            # 秒杀服务
│   │   ├── biz/            # 领域层
│   │   ├── data/           # 基础设施层
│   │   ├── service/        # 应用层
│   │   └── server/         # 服务器配置
│   ├── groupbuy/           # 拼团服务
│   └── lottery/            # 抽奖服务
├── pkg/                    # 公共包
├── configs/                # 配置文件
└── deploy/                 # Docker部署
```

## 快速开始

### 1. 启动基础设施

```bash
make docker-up
```

启动MySQL、Redis、RabbitMQ、Nacos

### 2. 初始化数据库

```bash
mysql -u root -proot < deploy/mysql/init.sql
```

### 3. 编译服务

```bash
make build
```

### 4. 运行服务

```bash
make run-seckill    # 秒杀服务 :18081(gRPC) :18091(HTTP)
make run-groupbuy   # 拼团服务 :18082(gRPC) :18092(HTTP)
make run-lottery    # 抽奖服务 :18083(gRPC) :18093(HTTP)
```

## 服务端口

| 服务 | gRPC端口 | HTTP端口 | 数据库 |
|------|---------|---------|--------|
| seckill-market | 18081 | 18091 | marketing_seckill |
| groupbuy-market | 18082 | 18092 | marketing_groupbuy |
| lottery-market | 18083 | 18093 | marketing_lottery |

## 核心设计模式

- **DDD六边形架构**: 领域层纯净，依赖倒置
- **责任链模式**: 拼团试算/锁单/结算过滤链
- **策略模式**: 折扣计算、退单策略
- **规则树**: 抽奖策略装配
- **本地消息表**: 分布式事务最终一致性
- **Redis Lua**: 秒杀库存原子扣减
