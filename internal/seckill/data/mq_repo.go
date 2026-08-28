package data

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/seckill/biz"
)

type mqRepo struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewMQRepo(data *Data) biz.MQRepo {
	return &mqRepo{conn: data.conn, channel: data.channel}
}

func (r *mqRepo) PublishOrderMessage(ctx context.Context, order *biz.SeckillOrder) error {
	queueName := "seckill_order_created"

	_, err := r.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body, _ := json.Marshal(order)

	return r.channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
