package biz

import "log/slog"

// compensate 记录补偿操作的失败。
//
// 补偿本身失败意味着系统可能停留在不一致状态（例如奖品库存已扣但订单没建成），
// 绝不能静默吞掉，必须留下错误日志以便告警与人工介入。
func compensate(action string, err error) {
	if err == nil {
		return
	}
	slog.Error("COMPENSATION FAILED, manual intervention required",
		slog.String("service", "lottery"),
		slog.String("action", action),
		slog.Any("error", err),
	)
}
