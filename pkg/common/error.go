package common

var (
	SuccessCode = ResponseCode{Code: "0000", Info: "成功"}
	FailCode    = ResponseCode{Code: "0001", Info: "失败"}

	// 通用错误
	ParamError       = ResponseCode{Code: "C0001", Info: "参数错误"}
	NotFoundError    = ResponseCode{Code: "C0002", Info: "资源不存在"}
	Unauthorized     = ResponseCode{Code: "C0003", Info: "未授权"}
	Forbidden        = ResponseCode{Code: "C0004", Info: "禁止访问"}
	InternalError    = ResponseCode{Code: "C0005", Info: "系统内部错误"}
	RateLimitError   = ResponseCode{Code: "C0006", Info: "请求过于频繁"}
	LockFailError    = ResponseCode{Code: "C0007", Info: "获取锁失败"}

	// 秒杀 (S = Seckill)
	SeckillActivityNotExist = ResponseCode{Code: "S0001", Info: "秒杀活动不存在"}
	SeckillStockNotEnough   = ResponseCode{Code: "S0002", Info: "库存不足"}
	SeckillOrderDuplicate   = ResponseCode{Code: "S0003", Info: "已参与过该活动"}
	SeckillActivityClosed   = ResponseCode{Code: "S0004", Info: "活动已关闭"}
	SeckillOrderNotFound    = ResponseCode{Code: "S0005", Info: "订单不存在"}

	// 拼团 (G = GroupBuy)
	GroupBuyActivityNotExist  = ResponseCode{Code: "G0001", Info: "拼团活动不存在"}
	GroupBuyOrderNotExist     = ResponseCode{Code: "G0002", Info: "订单不存在"}
	GroupBuyTeamFull          = ResponseCode{Code: "G0003", Info: "团已满"}
	GroupBuyTeamExpired       = ResponseCode{Code: "G0004", Info: "团已过期"}
	GroupBuyDiscountNotExist  = ResponseCode{Code: "G0005", Info: "折扣配置不存在"}
	GroupBuyTrialFail         = ResponseCode{Code: "G0006", Info: "试算失败"}
	GroupBuyLockFail          = ResponseCode{Code: "G0007", Info: "锁单失败"}
	GroupBuySettlementFail    = ResponseCode{Code: "G0008", Info: "结算失败"}
	GroupBuyRefundFail        = ResponseCode{Code: "G0009", Info: "退单失败"}
	GroupBuyTeamNotExist      = ResponseCode{Code: "G0010", Info: "团队不存在"}
	GroupBuyOrderStateInvalid = ResponseCode{Code: "G0011", Info: "订单状态无效"}

	// 抽奖 (L = Lottery)
	LotteryActivityNotExist = ResponseCode{Code: "L0001", Info: "抽奖活动不存在"}
	LotteryStrategyNotExist = ResponseCode{Code: "L0002", Info: "抽奖策略不存在"}
	LotteryDrawLimit        = ResponseCode{Code: "L0003", Info: "抽奖次数已用完"}
	LotteryAwardNotFound    = ResponseCode{Code: "L0004", Info: "奖品不存在"}

	// 消息队列 (M = MQ)
	MQPublishFail = ResponseCode{Code: "M0001", Info: "消息发送失败"}
	MQConsumeFail = ResponseCode{Code: "M0002", Info: "消息消费失败"}

	// 通知 (N = Notify)
	NotifyTaskNotExist   = ResponseCode{Code: "N0001", Info: "通知任务不存在"}
	NotifyTaskFail       = ResponseCode{Code: "N0002", Info: "通知任务失败"}
	NotifyTaskRetryLimit = ResponseCode{Code: "N0003", Info: "通知重试次数超限"}
)

type ResponseCode struct {
	Code string
	Info string
}

func (r ResponseCode) Error() string {
	return r.Code + ": " + r.Info
}
