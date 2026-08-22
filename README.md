# 营销中台 - Go微服务

基于DDD六边形架构的Go微服务营销中台，包含秒杀、拼团、抽奖三大核心服务。

## 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          客户端 / 前端                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Gateway (port: 8080)                       │
│                     统一入口，路由到各微服务                           │
└─────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│  seckill:18091 │          │ groupbuy:18092 │          │ lottery:18093 │
│   (秒杀服务)   │          │   (拼团服务)   │          │   (抽奖服务)   │
└───────┬───────┘          └───────┬───────┘          └───────┬───────┘
        │                           │                           │
        └───────────────────────────┼───────────────────────────┘
                                    │
                ┌───────────────────┼───────────────────────────┐
                │                   │                           │
                ▼                   ▼                           ▼
        ┌──────────────┐  ┌──────────────┐           ┌──────────────┐
        │  prize:18094 │  │  MySQL 8.0   │           │  Redis 7.x   │
        │ (公共库存服务) │  │  (数据存储)   │           │  (缓存/锁)   │
        └──────────────┘  └──────────────┘           └──────────────┘
```

## 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| 语言 | Go 1.22 | 高并发、高性能 |
| HTTP | net/http | 标准库HTTP服务器 |
| ORM | database/sql | 原生SQL |
| 数据库 | MySQL 8.0 | 三库隔离 |
| 缓存 | Redis 7.x | Lua原子操作、分布式锁 |
| 消息队列 | RabbitMQ | 异步消息、最终一致性 |
| 依赖注入 | Wire | 编译时依赖注入 |
| 容器化 | Docker Compose | 一键部署 |

## 快速开始

### 方式一：Docker Compose一键启动（推荐）

```bash
# 启动所有服务（MySQL + Redis + RabbitMQ + 五个微服务）
docker-compose -f deploy/docker-compose.yml up -d

# 查看日志
docker-compose -f deploy/docker-compose.yml logs -f

# 停止所有服务
docker-compose -f deploy/docker-compose.yml down
```

### 方式二：本地开发

```bash
# 1. 启动基础设施
docker-compose -f deploy/docker-compose-env.yml up -d

# 2. 初始化数据库
mysql -u root -proot < deploy/mysql/init.sql

# 3. 启动服务（五个终端）
go run cmd/seckill/main.go
go run cmd/groupbuy/main.go
go run cmd/lottery/main.go
go run cmd/prize/main.go
go run cmd/gateway/main.go
```

### 验证服务

```bash
# 健康检查
curl http://localhost:8080/health
curl http://localhost:18091/health
curl http://localhost:18092/health
curl http://localhost:18093/health
curl http://localhost:18094/health

# 秒杀下单
curl -X POST http://localhost:18091/api/v1/seckill/order/create \
  -H "Content-Type: application/json" \
  -d '{"activity_id":"act_001","user_id":1001}'

# 抽奖
curl -X POST http://localhost:18093/api/v1/lottery/raffle \
  -H "Content-Type: application/json" \
  -d '{"activity_id":"act_001","user_id":1001}'

# 通过网关访问
curl http://localhost:8080/api/v1/gateway/proxy/?service=seckill&path=/health
```

## 核心设计

### 1. DDD六边形架构 + 依赖倒置

```
biz (domain) 层定义接口
    ↓ 依赖
data (infrastructure) 层实现接口
```

- **领域层（biz）**：定义Repository接口，不依赖任何外部实现
- **基础设施层（data）**：实现Repository接口，可替换（MySQL/Redis/Mock）
- **应用层（service）**：编排领域服务，不关心具体实现

### 2. Redis Lua原子扣减库存

```lua
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock <= 0 then
    return 0
end
redis.call('DECR', KEYS[1])
return 1
```

**优势**：单次Lua脚本执行，原子性保证，无需分布式锁

### 3. 责任链模式（拼团试算）

```
RootNode -> MarketNode -> SwitchNode -> TagNode -> EndNode
```

每个节点可独立测试，可热插拔

### 4. 策略模式（退单）

```go
type RefundStrategy interface {
    Refund(ctx context.Context, order *GroupBuyOrder) error
}

// 未支付退单
type UnpaidRefundStrategy struct{}
// 已支付退单
type PaidRefundStrategy struct{}
```

### 5. 本地消息表（最终一致性）

```
1. 业务操作 + 消息写入同一事务
2. 定时任务扫描未发送消息
3. 发送成功后标记完成
4. 失败重试，保证最终一致
```

## API接口

### 秒杀服务 (port: 18091)

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/seckill/activity/query | GET | 查询活动 |
| /api/v1/seckill/order/create | POST | 秒杀下单 |
| /api/v1/seckill/order/query | GET | 查询订单 |
| /health | GET | 健康检查 |

### 拼团服务 (port: 18092)

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/groupbuy/activity/query | GET | 查询活动 |
| /api/v1/groupbuy/trial | POST | 试算优惠 |
| /api/v1/groupbuy/order/lock | POST | 锁单 |
| /api/v1/groupbuy/order/settlement | POST | 结算 |
| /api/v1/groupbuy/order/refund | POST | 退单 |

### 抽奖服务 (port: 18093)

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/lottery/activity/query | GET | 查询活动 |
| /api/v1/lottery/raffle | POST | 抽奖 |
| /api/v1/lottery/order/query | GET | 查询中奖记录 |

### 公共库存服务 (port: 18094)

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/prize/stock/deduct | POST | 扣减库存 |
| /api/v1/prize/stock/query | GET | 查询库存 |
| /api/v1/prize/stock/restore | POST | 恢复库存 |

### Gateway网关 (port: 8080)

| 接口 | 方法 | 说明 |
|------|------|------|
| /api/v1/gateway/proxy/ | ANY | 转发到指定服务 |
| /health | GET | 健康检查 |

## 单元测试

```bash
# 运行所有单元测试
go test ./... -v

# 运行特定服务测试
go test ./internal/seckill/biz/... -v
go test ./internal/groupbuy/biz/... -v
go test ./internal/lottery/biz/... -v

# 运行Redis集成测试（需要本地Redis）
go test ./internal/seckill/data/... -v -run TestRedis
```

### 测试覆盖

| 服务 | 测试数 | 覆盖内容 |
|------|--------|----------|
| seckill | 6 | 下单成功、库存不足、重复下单、并发扣减 |
| groupbuy | 7 | 试算(ZJ/ZK/N)、锁单、结算、退单 |
| lottery | 4 | 抽奖成功、多次抽奖、多用户 |

## 设计模式

| 模式 | 使用场景 | 说明 |
|------|---------|------|
| 责任链 | 拼团试算/锁单/结算 | 节点可热插拔 |
| 策略模式 | 折扣计算/退单 | 运行时切换策略 |
| 工厂模式 | 策略创建 | 解耦创建和使用 |
| 仓储模式 | 数据访问 | 接口抽象，可替换 |
| 防腐层 | DTO转换 | 保护领域模型 |

## 面试亮点

1. **Redis Lua原子操作**：库存扣减无需分布式锁
2. **DDD六边形架构**：领域层纯净，依赖倒置
3. **分布式事务**：本地消息表+定时补偿
4. **高并发设计**：Redis预热+异步落单
5. **设计模式**：责任链、策略模式实战应用
6. **Docker Compose**：一键部署完整环境

## License

MIT
