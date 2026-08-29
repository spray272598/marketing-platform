package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/marketing-platform/internal/groupbuy/biz"
)

type mqRepo struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	mu       sync.Mutex
	declared bool
}

func NewMQRepo(data *Data) biz.MQRepo {
	return &mqRepo{conn: data.conn, channel: data.channel}
}

// ensureDeclared 只在首次发布前声明一次队列，避免每条消息都多一次网络往返。
//
// 两个关键点：
//  1. 声明失败必须向上返回错误——RabbitMQ 向不存在的队列发布消息时，
//     在 mandatory=false 下会**静默丢弃**，导致成团/退款通知永久丢失；
//  2. 失败后不置 declared，允许下次重试；同时处理 channel 为 nil 的情况，
//     避免直接 nil 解引用导致进程崩溃。
func (r *mqRepo) ensureDeclared() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.declared {
		return nil
	}
	if r.channel == nil {
		return errors.New("rabbitmq channel is nil")
	}
	if _, err := r.channel.QueueDeclare("groupbuy_team_success", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue groupbuy_team_success: %w", err)
	}
	if _, err := r.channel.QueueDeclare("groupbuy_refund_success", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue groupbuy_refund_success: %w", err)
	}
	r.declared = true
	return nil
}

func (r *mqRepo) PublishTeamSuccessMessage(ctx context.Context, teamID string) error {
	if err := r.ensureDeclared(); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{"team_id": teamID})
	if err != nil {
		return fmt.Errorf("marshal team success message: %w", err)
	}
	return r.channel.PublishWithContext(ctx, "", "groupbuy_team_success", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}

func (r *mqRepo) PublishRefundMessage(ctx context.Context, orderID string) error {
	if err := r.ensureDeclared(); err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{"order_id": orderID})
	if err != nil {
		return fmt.Errorf("marshal refund message: %w", err)
	}
	return r.channel.PublishWithContext(ctx, "", "groupbuy_refund_success", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}
