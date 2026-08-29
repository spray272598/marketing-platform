package biz

import "log/slog"

// compensate 记录补偿操作的失败。
//
// 补偿本身失败意味着系统可能停留在不一致状态（例如订单已创建但团队没建成），
// 绝不能静默吞掉，必须留下错误日志以便告警与人工介入。
//
// 说明：结算与退款链路已改用 pkg/saga 编排，其补偿失败会汇总进
// SagaError.CompErrors；本 helper 用于其余链路的手动补偿。
func compensate(action string, err error) {
	if err == nil {
		return
	}
	slog.Error("COMPENSATION FAILED, manual intervention required",
		slog.String("service", "groupbuy"),
		slog.String("action", action),
		slog.Any("error", err),
	)
}
