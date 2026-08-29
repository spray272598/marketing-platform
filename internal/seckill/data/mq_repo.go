package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/seckill/biz"
)

type mqRepo struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	once    sync.Once
}

func NewMQRepo(data *Data) biz.MQRepo {
	return &mqRepo{conn: data.conn, channel: data.channel}
}

// ensureDeclared 仅在每个实例中声明一次队列，避免每条消息都多一次网络往返。
func (r *mqRepo) ensureDeclared() {
	r.once.Do(func() {
		if r.channel == nil {
			return
		}
		_, _ = r.channel.QueueDeclare("seckill_order_created", true, false, false, false, nil)
	})
}

func (r *mqRepo) PublishOrderMessage(ctx context.Context, order *biz.SeckillOrder) error {
	r.ensureDeclared()

	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	return r.channel.PublishWithContext(ctx,
		"",
		"seckill_order_created",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
