FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/seckill ./cmd/seckill
RUN go build -o /app/bin/groupbuy ./cmd/groupbuy
RUN go build -o /app/bin/lottery ./cmd/lottery
RUN go build -o /app/bin/stock ./cmd/stock
RUN go build -o /app/bin/gateway ./cmd/gateway

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/seckill ./seckill
COPY --from=builder /app/bin/groupbuy ./groupbuy
COPY --from=builder /app/bin/lottery ./lottery
COPY --from=builder /app/bin/stock ./stock
COPY --from=builder /app/bin/gateway ./gateway

CMD ["./seckill"]

FROM builder AS seckill
RUN echo "seckill built"

FROM builder AS groupbuy
RUN echo "groupbuy built"

FROM builder AS lottery
RUN echo "lottery built"

FROM builder AS stock
RUN echo "stock built"

FROM builder AS gateway
RUN echo "gateway built"
