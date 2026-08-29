package data

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/groupbuy/biz"
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
		_, _ = r.channel.QueueDeclare("groupbuy_team_success", true, false, false, false, nil)
		_, _ = r.channel.QueueDeclare("groupbuy_refund_success", true, false, false, false, nil)
	})
}

func (r *mqRepo) PublishTeamSuccessMessage(ctx context.Context, teamID string) error {
	r.ensureDeclared()
	payload, _ := json.Marshal(map[string]string{"team_id": teamID})
	return r.channel.PublishWithContext(ctx, "", "groupbuy_team_success", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}

func (r *mqRepo) PublishRefundMessage(ctx context.Context, orderID string) error {
	r.ensureDeclared()
	payload, _ := json.Marshal(map[string]string{"order_id": orderID})
	return r.channel.PublishWithContext(ctx, "", "groupbuy_refund_success", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}
