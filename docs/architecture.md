# 营销中台 Go微服务架构设计

## 一、技术栈

| 层次 | 技术选型 | 说明 |
|------|---------|------|
| 语言 | Go 1.22+ | 高并发、高性能 |
| 框架 | Kratos v2 | B站开源，Go微服务标杆框架（服务治理） |
| RPC | gRPC + HTTP | 内部高性能RPC + 外部HTTP网关转发 |
| ORM | database/sql | 原生SQL，控制依赖复杂度 |
| 数据库 | MySQL 8.0 | 三库隔离（秒杀/拼团/抽奖） |
| 缓存 | Redis 7.x | Lua原子操作、分布式锁、库存预热 |
| 消息队列 | RabbitMQ | 异步落单、成团通知、最终一致性 |
| 注册中心 | Nacos 2.x | 服务发现 + 配置中心 |
| 可观测性 | Prometheus + 自定义Trace | Metrics + Trace + Log 三件套 |
| 日志 | slog (结构化) | 原生结构化日志，对接ELK/Loki |
| 容器化 | Docker + Docker Compose | 一键启动全栈环境 |

## 二、三服务划分

| 服务 | 端口 | 数据库 | 核心能力 |
|------|------|--------|---------|
| seckill | 18091 | marketing_seckill | 活动发布、库存预热、Redis Lua原子扣减、异步落单、一人一单原子校验 |
| groupbuy | 18092 | marketing_groupbuy | 拼团活动、折扣试算(责任链)、锁单、结算成团、退单逆向、本地消息表+定时补偿 |
| lottery | 18093 | marketing_lottery | 抽奖策略、规则树、概率计算、奖品发放 |
| stock | 18094 | marketing_stock | 统一库存服务，通过stock_key区分库存类型(product/team/prize) |
| gateway | 8080 | - | 统一入口，自建HTTP网关转发到各微服务 |

## 三、DDD分层映射

| 层次 | Go路径 | 职责 |
|------|--------|------|
| 接口定义 | api/{svc}/v1/ | Protobuf接口定义 |
| 触发/传输层 | internal/{svc}/server/ | HTTP服务器配置、路由注册 |
| 应用层 | internal/{svc}/service/ | 用例编排，编排领域服务 |
| 领域层 | internal/{svc}/biz/ | 业务逻辑、领域模型、Repository接口定义 |
| 基础设施层 | internal/{svc}/data/ | Repository接口实现（MySQL/Redis/MQ） |
| 公共层 | pkg/common/ | 公共类型/枚举/错误码/响应封装 |

## 四、关键设计模式

| 模式 | 使用场景 | Go实现 |
|------|---------|--------|
| 责任链 | 拼团试算/锁单/结算过滤 | 接口链 + slice遍历，节点可热插拔 |
| 策略模式 | 折扣计算(ZJ/MJ/N/ZK)、退单策略 | map注册 + interface |
| 工厂模式 | 策略创建 | 解耦创建和使用 |
| 仓储模式 | 数据访问抽象 | biz层定义interface，data层实现 |
| 防腐层 | DTO转换 | 保护领域模型不被外部污染 |
| 本地消息表 | 最终一致性 | notify_task表 + 定时补偿Consumer + 指数退避重试 |
| 布隆过滤器 | 秒杀防刷 | Redis Set原子校验一人一单 |

## 五、高并发设计亮点

### 5.1 Redis Lua原子扣减 + 一人一单校验

```lua
-- 原子操作：检查库存 + 检查用户是否已下单 + 扣减库存
local stock_key = KEYS[1]       -- seckill:stock:{activity_id}
local user_key = KEYS[2]        -- seckill:user:{activity_id}
local user_id = ARGV[1]

-- 1. 检查用户是否已下单
if redis.call('SISMEMBER', user_key, user_id) == 1 then
    return 2  -- 已下单
end

-- 2. 检查库存
local stock = tonumber(redis.call('GET', stock_key))
if stock == nil or stock <= 0 then
    return 0  -- 库存不足
end

-- 3. 原子扣减
redis.call('DECR', stock_key)

-- 4. 标记用户已下单
redis.call('SADD', user_key, user_id)
redis.call('EXPIRE', user_key, 3600)

return 1  -- 成功
```

**优势**：单次Lua脚本执行，保证原子性；无需分布式锁；天然防止一人一单重复扣减。

### 5.2 本地消息表 + 定时补偿（最终一致性）

```
1. 业务操作(订单创建) + 本地消息写入(notify_task表) 同一事务
2. NotifyConsumer 定时(5s)扫描未发送消息
3. 发送成功 → 标记完成
4. 发送失败 → 指数退避重试(1s→2s→4s→8s)
5. 超过最大重试次数 → 标记失败 + 告警
```

### 5.3 限流保护

- 全局限流：基于 `golang.org/x/time/rate` 的令牌桶，按IP维度限流
- Redis原子限流：秒杀接口可叠加Redis计数器限流

## 六、Design Highlights

1. **Redis Lua原子操作**：秒杀库存扣减 + 一人一单原子校验，天然防止超卖和重复下单
2. **DDD六边形架构**：领域层纯净，依赖倒置，Repository接口可替换
3. **分布式事务最终一致性**：本地消息表 + 定时补偿 + 指数退避重试
4. **高并发设计**：Redis预热 + Lua原子操作 + 异步落单
5. **设计模式实战**：责任链、策略模式、工厂模式、仓储模式
6. **可观测性三件套**：Prometheus Metrics + 自定义Trace链路 + 结构化日志
7. **Docker Compose一键部署**：完整开发环境
8. **统一库存服务抽象**：stock_key规范化(product/team/prize)，跨业务复用