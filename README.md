# Marketing Platform — Go Microservices

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A marketing middle-platform built with Go microservices. It delivers three
high-frequency promotional scenarios — **flash sale (seckill)**, **group buy**,
and **lottery** — on top of a shared **stock service**, a unified API
**gateway**, and an event-driven consistency layer.

The project is organized around DDD bounded contexts, dependency inversion, and
a small set of cross-cutting packages (`auth`, `saga`, `idgen`, `observability`,
`middleware`) so that each business domain stays focused on its own rules while
common concerns (identity, transactions, IDs, metrics) are solved once.

> This repository is maintained as a long-term open-source engineering project
> under the MIT license. Issues and discussions are welcome.

---

## Features

- **Flash sale (seckill)** — Redis-backed atomic stock deduction (Lua) on the
  hot path, async order persistence via message queue, and timeout close-order
  job.
- **Group buy** — strategy-tree discount calculation, a responsibility-chain
  trial/lock/settlement pipeline, and Saga-orchestrated settlement/refund with
  compensating actions.
- **Lottery** — rule-tree assembly, daily/monthly quota control, and prize
  granting through the unified stock service.
- **Unified stock service** — a single inventory abstraction shared by all three
  scenarios, keyed by `stock_key` (`product:{sku}`, `team:{team_id}`,
  `prize:{prize_id}`).
- **Identity & trust** — JWT bearer auth for user-facing endpoints; an internal
  token for inter-service calls. `user_id` is never taken from the request body.
- **Distributed transactions** — orchestration-style Saga with reverse-order
  compensation, plus a local outbox / RabbitMQ pipeline for eventual consistency.
- **Global, monotonic IDs** — Meituan Leaf "segment mode" for cross-service
  unique, increasing order IDs.
- **Observability** — Prometheus metrics + **Grafana** dashboards
  (datasource & panels auto-provisioned on startup), an in-process trace
  collector, and structured `slog` logging, behind a reusable middleware chain.
- **Dual protocol (HTTP + gRPC)** — every business service exposes both a
  REST/JSON HTTP API and an equivalent gRPC service, with shared `user_id`
  injection and JWT + internal-token auth on both transports.
- **Config center** — Nacos for externalized, environment-aware configuration.
- **One-command deployment** — Docker Compose brings up MySQL, Redis,
  RabbitMQ, Nacos, Prometheus, **Grafana** and all five services.

---

## Architecture

```
                         ┌──────────────────────────┐
                         │      Clients / Web        │
                         └─────────────┬────────────┘
                                       │  JWT Bearer
                                       ▼
                         ┌──────────────────────────┐
                         │   Gateway  (port 8080)     │
                         │  route + reverse proxy     │
                         │  JWT verify + body forward│
                         └─────────────┬────────────┘
            ┌──────────────────────────┼──────────────────────────┐
            │                          │                          │
            ▼                          ▼                          ▼
   ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
   │ seckill :18091/:18095 │  │ groupbuy :18092/:18096│  │ lottery  :18093/:18097│
   │  HTTP / gRPC (flash)  │  │  HTTP / gRPC (group)  │  │   HTTP / gRPC (lotto) │
   └────────┬──────────────┘  └────────┬──────────────┘  └────────┬──────────────┘
            │  X-Internal-Token         │  X-Internal-Token        │
            └─────────────┬─────────────┴───────────┬─────────────┘
                          ▼                         ▼
                 ┌─────────────────┐    ┌──────────────────────┐
                 │  stock :18094   │    │  RabbitMQ (outbox)   │
                 │ (shared stock)  │    │  eventual consistency│
                 └────────┬────────┘    └──────────┬───────────┘
                          │                        │
        ┌─────────────────┼────────────────────────┼───────────────┐
        ▼                 ▼                        ▼               ▼
   ┌──────────┐    ┌──────────────┐        ┌────────────┐  ┌──────────────┐
   │  MySQL   │    │    Redis 7    │        │    Nacos   │  │ Prometheus /  │
   │ (per ctx)│    │ (Lua deduct) │        │  config    │  │   Grafana     │
   └──────────┘    └──────────────┘        └────────────┘  └──────────────┘
```

Each business service owns its own MySQL database (seckill / groupbuy /
lottery / stock) — database-per-service isolation. Redis is shared for the
flash-sale hot path; Nacos holds environment configuration; Prometheus scrapes
the `/metrics` endpoint exposed by every service.

---

## Tech Stack

| Concern            | Technology                                             |
|--------------------|--------------------------------------------------------|
| Language           | Go 1.25                                                |
| Microservice frame | Kratos v3 (`transport/http`)                           |
| Dependency inject. | Wire (compile-time)                                    |
| ORM                | Ent (`entgo.io/ent`)                                   |
| Identity           | JWT (`golang-jwt/jwt/v5`)                              |
| ID generation      | Meituan Leaf segment mode (MySQL-backed)               |
| Cache              | Redis 7 (`redis/go-redis/v9`), Lua atomic scripts     |
| Message queue      | RabbitMQ (outbox / async events)                       |
| Config center      | Nacos (`nacos-group/nacos-sdk-go/v2`)                 |
| Rate limiting      | Token bucket (`golang.org/x/time/rate`)                |
| RPC / dual protocol| gRPC (`google.golang.org/grpc`), protobuf, Wire DI     |
| Observability      | Prometheus client, in-process trace collector, `slog`  |
| Visualization      | Grafana (datasource & dashboard auto-provisioning)     |
| Containerization   | Docker Compose                                         |

---

## Project Layout

```
.
├── api/                 # Protobuf/OpenAPI service contracts (per domain)
├── cmd/                 # Service entrypoints (gateway, seckill, groupbuy, lottery, stock)
│   └── <svc>/wire_gen.go
├── configs/             # Per-service Kratos bootstrap configs
├── deploy/              # Docker / MySQL init / Nacos / Prometheus assets
│   ├── docker-compose.yml
│   ├── docker-compose-env.yml
│   └── docker-compose-microservices.yml
├── docs/                # Architecture & design notes
├── internal/
│   ├── conf/            # Generated config structs
│   ├── gateway/         # Reverse proxy + auth pass-through
│   ├── seckill/         # Flash sale domain (biz / data / service / server)
│   ├── groupbuy/        # Group buy domain
│   ├── lottery/         # Lottery domain
│   └── stock/           # Shared inventory domain
└── pkg/
    ├── auth/            # JWT issue/verify + HTTP/gRPC auth middleware
    ├── saga/            # Orchestration Saga engine (compensation)
    ├── idgen/           # Leaf segment-mode ID generator
    ├── middleware/      # Trace / recovery / ratelimit / metrics / auth chain
    ├── observability/   # Prometheus metrics + trace collector
    ├── config/          # Nacos + file/env loader
    ├── ratelimit/       # Token-bucket limiter
    ├── stockclient/     # Internal client for the stock service
    ├── common/          # Errors, events, response, shutdown, trace helpers
    └── log/             # Logger setup
```

Each domain follows the same internal shape:

```
biz/      domain interfaces + business rules (pure, no framework deps)
data/     Ent schemas + repository implementations (dependency inversion)
service/  transport adapters (HTTP handlers, request/response mapping)
server/   Kratos server wiring
```

---

## Core Designs

### 1. DDD hexagonal architecture + dependency inversion

The `biz` layer declares repository interfaces; the `data` layer implements
them with Ent/Redis/RabbitMQ. Application code in `biz` depends only on
interfaces, so infrastructure is swappable (MySQL ↔ Mock) and unit-testable
without a database.

### 2. JWT identity & zero-trust user_id

`pkg/auth` issues and verifies JWT bearer tokens. A middleware extracts
`user_id` from the verified claim and injects it into `context`. Business
handlers **never** read `user_id` from the request body — this closes the
impersonation hole (any client could otherwise place orders as another user).

Two trust levels:

- **User-facing endpoints** require `Authorization: Bearer <jwt>`.
- **Inter-service endpoints** (e.g. groupbuy settlement/refund, stock
  deduct/restore) require `X-Internal-Token: <MARKETING_INTERNAL_TOKEN>`.

`MARKETING_AUTH_DISABLED=1` turns auth off for local development.

### 3. Leaf segment-mode ID generator

`pkg/idgen` implements Meituan Leaf's segment mode. Each `bizTag` batches a
range from a `SegmentStore` (MySQL `SELECT … FOR UPDATE` + `UPDATE`) into
process-local memory and dispenses IDs sequentially; the next range is fetched
only when the current one is exhausted. ID allocation is therefore nearly
lock-free against the center store, and order IDs are globally unique and
monotonically increasing across services.

### 4. Redis Lua atomic stock deduction (flash sale)

On the hot path, stock is decremented inside a single Lua script so the
read-decrement-check is atomic — no distributed lock required:

```lua
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock <= 0 then
    return 0
end
redis.call('DECR', KEYS[1])
return 1
```

After the atomic pass, the order is persisted asynchronously through the
message pipeline, and a timeout job closes unpaid orders and restores stock.

### 5. Saga distributed transaction (group buy settlement / refund)

`pkg/saga` is an orchestration engine: a `Coordinator` runs `Step{Action,
Compensate}` in order and, on failure, runs compensations in **reverse order**.
Compensation runs under `context.WithoutCancel` so a cancelled parent context
does not abort the rollback; partial failures are collected in `SagaError`.
Settlement deducts cross-service stock (compensate: restore), idempotently
persists team state, then emits a "team formed" notification.

### 6. Local outbox + RabbitMQ (eventual consistency)

Long-running side effects (notifications, async order persistence) are written
to a local message table inside the business transaction, then a scanner /
RabbitMQ publisher relays them. This keeps the business write and the event
emit atomic at the source and lets consumers retry toward eventual consistency.

### 7. Unified stock abstraction

The `stock` service is the single owner of inventory. Callers address stock by
`stock_key`:

```
product:{sku_id}   - goods stock (flash sale deduct)
team:{team_id}     - team slots  (group buy deduct)
prize:{prize_id}   - prize stock (lottery deduct)
```

All three scenarios go through one service, so deduction, restoration, and
queries share one consistency boundary.

### 8. Reusable middleware chain

`pkg/middleware` composes, in order: **trace** (UUID trace/span id) →
**recovery** → **ratelimit** (token bucket per key) → **metrics** → **auth**
(JWT / internal token). Each stage is a standard `func(http.Handler)
http.Handler`, so it is testable and reorderable independently.

### 9. Observability

- **Metrics**: Prometheus counters/histograms for request total, duration,
  status, DB connections, cache hit-rate, and per-business gauges, exposed at
  `/metrics` on **both** the HTTP and gRPC ports of every service for scraping.
- **Dashboards**: Grafana is auto-provisioned on startup — a Prometheus
  datasource and a `Marketing / Marketing Platform Overview` dashboard
  (QPS, p95 latency, status codes, goroutines, heap, GC) are loaded from
  `deploy/grafana/provisioning/` with no manual UI setup.
- **Tracing**: an in-process `TraceCollector` builds a span tree
  (trace/span/parent, tags, logs) propagated through the middleware chain.
- **Logging**: structured `slog` output, wired through Kratos' logger.

### 10. Design patterns in use

| Pattern            | Where                                        |
|--------------------|----------------------------------------------|
| Responsibility chain | Group buy trial / lock / settlement nodes |
| Strategy           | Discount calculation, refund strategies      |
| Factory            | Strategy / node creation                     |
| Repository         | `biz` interfaces implemented by `data`       |
| Anti-corruption layer | DTO ↔ domain model mapping               |

---

## Getting Started

### Option A — Docker Compose (full stack)

```bash
# Bring up MySQL, Redis, RabbitMQ, Nacos, Prometheus, Grafana and all five services.
docker compose -f deploy/docker-compose-microservices.yml up -d

# Tail logs
docker compose -f deploy/docker-compose-microservices.yml logs -f

# Tear down
docker compose -f deploy/docker-compose-microservices.yml down
```

After the stack is up:

- **Grafana** dashboards: <http://localhost:3000> (admin / admin).
- **Prometheus** UI: <http://localhost:9090>.
- **Metrics** (Prometheus scrape targets): `/metrics` on every service
  (HTTP `:18091`–`:18094`, gRPC `:18095`–`:18097`).
- **gRPC** endpoints are served alongside HTTP on each business service
  (seckill `:18095`, groupbuy `:18096`, lottery `:18097`).

Set the auth secrets before starting (defaults are dev placeholders):

```bash
export MARKETING_AUTH_SECRET="a-strong-secret"
export MARKETING_INTERNAL_TOKEN="a-strong-internal-token"
```

### Option B — Local development

```bash
# 1. Start only the infrastructure.
docker compose -f deploy/docker-compose-env.yml up -d

# 2. Initialize databases (per-service schemas under deploy/mysql).
mysql -u root -proot < deploy/mysql/init.sql

# 3. Run the five services (separate terminals).
export MARKETING_AUTH_SECRET="dev-secret" MARKETING_INTERNAL_TOKEN="dev-internal"
go run ./cmd/gateway
go run ./cmd/seckill
go run ./cmd/groupbuy
go run ./cmd/lottery
go run ./cmd/stock
```

### Health checks

```bash
curl http://localhost:8080/health          # gateway
curl http://localhost:18091/health         # seckill
curl http://localhost:18092/health         # groupbuy
curl http://localhost:18093/health         # lottery
curl http://localhost:18094/health         # stock
curl http://localhost:18094/metrics        # prometheus scrape target
```

### Example calls

```bash
# Obtain a JWT for user 1001 (auth issuer endpoint / your token service)
TOKEN=$(curl -s ... | jq -r .token)

# Flash sale order (user_id comes from the token, not the body)
curl -X POST http://localhost:8080/api/v1/seckill/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"activity_id":"act_001"}'

# Group buy trial
curl -X POST http://localhost:8080/api/v1/groupbuy/trial \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"activity_id":"act_001","goods_id":"g_1","source":"NEW"}'

# Lottery raffle
curl -X POST http://localhost:8080/api/v1/lottery/raffle \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"activity_id":"act_001"}'
```

> Settlement and refund are internal operations. They are invoked by the
> system (e.g. the close-order job or a completed team) with the
> `X-Internal-Token` header, not by end users.

---

## API Reference

All user-facing routes are exposed through the gateway at `:8080` and routed to
the owning service. Internal routes are marked **[internal]** and require
`X-Internal-Token`.

### Gateway (`:8080`)

| Route                     | Method | Notes                         |
|---------------------------|--------|-------------------------------|
| `/api/v1/gateway/proxy/`  | ANY    | Reverse proxy to a service    |
| `/health`                 | GET    | Health check                  |

### Flash sale (`:18091`)

| Route                                | Method | Auth            |
|-------------------------------------|--------|-----------------|
| `/api/v1/seckill/activity/query`    | GET    | JWT (optional)  |
| `/api/v1/seckill/order/create`      | POST   | JWT Bearer      |
| `/api/v1/seckill/order/query`       | GET    | JWT Bearer      |

### Group buy (`:18092`)

| Route                                  | Method | Auth                     |
|---------------------------------------|--------|--------------------------|
| `/api/v1/groupbuy/activity/query`     | GET    | JWT (optional)           |
| `/api/v1/groupbuy/trial`              | POST   | JWT Bearer               |
| `/api/v1/groupbuy/lock`               | POST   | JWT Bearer               |
| `/api/v1/groupbuy/settlement`         | POST   | **[internal]** token     |
| `/api/v1/groupbuy/refund`             | POST   | **[internal]** token     |

### Lottery (`:18093`)

| Route                                  | Method | Auth            |
|---------------------------------------|--------|-----------------|
| `/api/v1/lottery/activity/query`      | GET    | JWT (optional)  |
| `/api/v1/lottery/strategy/query`      | GET    | JWT (optional)  |
| `/api/v1/lottery/raffle`              | POST   | JWT Bearer      |
| `/api/v1/lottery/order/query`         | GET    | JWT Bearer      |

### Stock (`:18094`)

| Route                     | Method | Auth                     |
|---------------------------|--------|--------------------------|
| `/api/v1/stock/query`     | GET    | **[internal]** token     |
| `/api/v1/stock/deduct`    | POST   | **[internal]** token     |
| `/api/v1/stock/restore`   | POST   | **[internal]** token     |

---

## Testing

```bash
# Run the full suite
go test ./...

# Domain logic only
go test ./internal/seckill/biz/... ./internal/groupbuy/biz/... ./internal/lottery/biz/...

# Redis integration tests (requires a local Redis)
go test ./internal/seckill/data/... -run TestRedis
```

Tests cover flash-sale success / insufficient stock / duplicate order /
concurrent deduction, group-buy trial (ZJ/ZK/N) / lock / settlement / refund,
lottery success / repeated draws / multi-user, plus `pkg/auth` and `pkg/saga`
unit suites (incl. compensation ordering and `alg=none` rejection).

---

## Roadmap

Implemented:

- [x] Flash sale with Redis Lua deduction + async persistence + close-order job
- [x] Group buy: strategy tree + responsibility chain + Saga settlement/refund
- [x] Lottery: rule tree + quota + prize granting
- [x] Unified stock service
- [x] JWT auth (user + internal token) and zero-trust `user_id`
- [x] Leaf segment-mode global IDs
- [x] Local outbox + RabbitMQ eventual consistency
- [x] Nacos config center
- [x] Observability: Prometheus metrics, Grafana auto-provisioned dashboards, trace collector, `slog`
- [x] Dual protocol: HTTP + gRPC on every business service, shared auth
- [x] Docker Compose full-stack deployment (incl. Prometheus + Grafana)

Planned:

- [ ] CI/CD pipeline (build → test → image → deploy)
- [ ] Load-test baseline and a published benchmark report
- [ ] Additional promotional scenarios (e.g. full reduction, coupon)
- [ ] Sharding / read-write split for larger data scale

---

## License

[MIT](LICENSE)
