package data

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/groupbuy/biz"
)

type mqRepo struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewMQRepo(data *Data) biz.MQRepo {
	return &mqRepo{conn: data.conn, channel: data.channel}
}

func (r *mqRepo) PublishTeamSuccessMessage(ctx context.Context, teamID string) error {
	queueName := "groupbuy_team_success"
	_, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return r.channel.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(fmt.Sprintf(`{"team_id":"%s"}`, teamID)),
	})
}

func (r *mqRepo) PublishRefundMessage(ctx context.Context, orderID string) error {
	queueName := "groupbuy_refund_success"
	_, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return r.channel.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(fmt.Sprintf(`{"order_id":"%s"}`, orderID)),
	})
}
