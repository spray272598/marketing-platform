package data

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/seckill/biz"
)

type MQRepo struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewMQRepo(conn *amqp.Connection, ch *amqp.Channel) *MQRepo {
	return &MQRepo{conn: conn, channel: ch}
}

func (r *MQRepo) PublishOrderMessage(ctx context.Context, order *biz.SeckillOrder) error {
	queueName := "seckill_order_created"

	// 声明队列
	_, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body, _ := json.Marshal(order)

	return r.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
