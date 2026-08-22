package common

var (
	SuccessCode = ResponseCode{Code: "0000", Info: "成功"}
	FailCode    = ResponseCode{Code: "0001", Info: "失败"}

	// 秒杀
	SeckillActivityNotExist = ResponseCode{Code: "S0001", Info: "秒杀活动不存在"}
	SeckillStockNotEnough   = ResponseCode{Code: "S0002", Info: "库存不足"}
	SeckillOrderDuplicate   = ResponseCode{Code: "S0003", Info: "已参与过该活动"}
	SeckillActivityClosed   = ResponseCode{Code: "S0004", Info: "活动已关闭"}

	// 拼团
	GroupBuyActivityNotExist = ResponseCode{Code: "G0001", Info: "拼团活动不存在"}
	GroupBuyOrderNotExist    = ResponseCode{Code: "G0002", Info: "订单不存在"}
	GroupBuyTeamFull         = ResponseCode{Code: "G0003", Info: "团已满"}
	GroupBuyTeamExpired      = ResponseCode{Code: "G0004", Info: "团已过期"}

	// 抽奖
	LotteryActivityNotExist = ResponseCode{Code: "L0001", Info: "抽奖活动不存在"}
	LotteryStrategyNotExist = ResponseCode{Code: "L0002", Info: "抽奖策略不存在"}
	LotteryDrawLimit        = ResponseCode{Code: "L0003", Info: "抽奖次数已用完"}
)

type ResponseCode struct {
	Code string
	Info string
}
