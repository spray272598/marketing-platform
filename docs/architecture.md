# 营销中台 Go微服务架构设计

## 一、技术栈

| 层次 | 技术选型 | 说明 |
|------|---------|------|
| 框架 | Kratos v2 | B站开源，Go微服务标杆框架 |
| RPC | gRPC + gRPC-Gateway | 内部高性能RPC + 外部HTTP兼容 |
| ORM | GORM v2 | Go生态最流行ORM |
| 数据库 | MySQL 8.x | 三库隔离（秒杀/拼团/抽奖） |
| 缓存 | Redis 7.x | 分布式锁、库存预热、Lua原子操作 |
| 消息队列 | RabbitMQ | 异步落单、成团通知、最终一致性 |
| 注册中心 | Nacos 2.x | 服务发现 + 配置中心（与Java项目一致） |
| 链路追踪 | Jaeger | OpenTelemetry标准 |
| 监控指标 | Prometheus + Grafana | Kratos内置埋点 |
| 日志 | Zap | 结构化日志，可对接ELK |
| 容器化 | Docker + Docker Compose | 一键启动全栈环境 |

## 二、三服务划分

| 服务 | 端口(gRPC) | 端口(HTTP) | 数据库 | 核心能力 |
|------|-----------|-----------|-------|---------|
| seckill-market | 18081 | 18091 | marketing_seckill | 活动发布、库存预热、Redis Lua扣减、异步落单、超时关单 |
| groupbuy-market | 18082 | 18092 | marketing_groupbuy | 拼团活动、折扣试算、锁单、结算成团、退单逆向 |
| lottery-market | 18083 | 18093 | marketing_lottery | 抽奖策略、规则树、概率计算、奖品发放 |

## 三、DDD分层映射（Java -> Go/Kratos）

| Java模块 | Go路径 | 职责 |
|----------|--------|------|
| trigger | internal/{svc}/server/ | gRPC/HTTP服务器配置 |
| api | api/{svc}/v1/ | Protobuf接口定义 |
| app (service) | internal/{svc}/service/ | 应用层-用例编排 |
| domain | internal/{svc}/biz/ | 领域层-业务逻辑 |
| infrastructure | internal/{svc}/data/ | 基础设施层-仓储实现 |
| types | pkg/common/ | 公共类型/枚举/错误码 |

## 四、关键设计模式

| 模式 | 使用场景 | Go实现 |
|------|---------|--------|
| 责任链 | 拼团试算/锁单/结算过滤 | 接口链 + slice遍历 |
| 策略模式 | 折扣计算(ZJ/MJ/N/ZK)、退单策略 | map注册 + interface |
| 规则树 | 抽奖策略装配 | DB存储 + 递归遍历 |
| 仓储模式 | 数据访问抽象 | interface + struct实现 |
| 防腐层 | 入站/出站DTO转换 | assembler包 |
| 本地消息表 | 最终一致性 | notify_task表 + 定时补偿 |

## 五、秋招面试亮点

1. **Redis Lua原子操作**：秒杀库存扣减，单次脚本执行INCR+判断+SET
2. **gRPC双向流**：服务间高性能通信
3. **Nacos动态配置**：DCC热更新，无需重启
4. **本地消息表+定时补偿**：分布式事务最终一致性
5. **Docker Compose一键部署**：完整开发环境
6. **DDD六边形架构**：领域层纯净，依赖倒置
