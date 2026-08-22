FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build services
RUN go build -o /app/bin/seckill ./cmd/seckill
RUN go build -o /app/bin/groupbuy ./cmd/groupbuy
RUN go build -o /app/bin/lottery ./cmd/lottery
RUN go build -o /app/bin/prize ./cmd/prize
RUN go build -o /app/bin/gateway ./cmd/gateway

# Runtime
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy all binaries
COPY --from=builder /app/bin/seckill ./seckill
COPY --from=builder /app/bin/groupbuy ./groupbuy
COPY --from=builder /app/bin/lottery ./lottery
COPY --from=builder /app/bin/prize ./prize
COPY --from=builder /app/bin/gateway ./gateway

# Default to seckill, override in docker-compose
CMD ["./seckill"]

# Multi-stage targets
FROM builder AS seckill-build
RUN echo "seckill built"

FROM builder AS groupbuy-build
RUN echo "groupbuy built"

FROM builder AS lottery-build
RUN echo "lottery built"

FROM builder AS prize-build
RUN echo "prize built"

FROM builder AS gateway-build
RUN echo "gateway built"
