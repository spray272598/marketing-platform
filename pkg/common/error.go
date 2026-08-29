package common

import (
	"net/http"
	"strings"
)

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

// codeRegistry 把业务错误码映射回安全的展示文案，避免把内部错误细节泄露给客户端。
var codeRegistry = map[string]ResponseCode{
	SuccessCode.Code:               SuccessCode,
	FailCode.Code:                  FailCode,
	ParamError.Code:                ParamError,
	NotFoundError.Code:             NotFoundError,
	Unauthorized.Code:             Unauthorized,
	Forbidden.Code:                Forbidden,
	InternalError.Code:            InternalError,
	RateLimitError.Code:           RateLimitError,
	LockFailError.Code:            LockFailError,
	SeckillActivityNotExist.Code:  SeckillActivityNotExist,
	SeckillStockNotEnough.Code:    SeckillStockNotEnough,
	SeckillOrderDuplicate.Code:    SeckillOrderDuplicate,
	SeckillActivityClosed.Code:    SeckillActivityClosed,
	SeckillOrderNotFound.Code:     SeckillOrderNotFound,
	GroupBuyActivityNotExist.Code: GroupBuyActivityNotExist,
	GroupBuyOrderNotExist.Code:    GroupBuyOrderNotExist,
	GroupBuyTeamFull.Code:         GroupBuyTeamFull,
	GroupBuyTeamExpired.Code:      GroupBuyTeamExpired,
	GroupBuyDiscountNotExist.Code: GroupBuyDiscountNotExist,
	GroupBuyTrialFail.Code:        GroupBuyTrialFail,
	GroupBuyLockFail.Code:         GroupBuyLockFail,
	GroupBuySettlementFail.Code:   GroupBuySettlementFail,
	GroupBuyRefundFail.Code:       GroupBuyRefundFail,
	GroupBuyTeamNotExist.Code:     GroupBuyTeamNotExist,
	GroupBuyOrderStateInvalid.Code: GroupBuyOrderStateInvalid,
	LotteryActivityNotExist.Code:   LotteryActivityNotExist,
	LotteryStrategyNotExist.Code:   LotteryStrategyNotExist,
	LotteryDrawLimit.Code:          LotteryDrawLimit,
	LotteryAwardNotFound.Code:      LotteryAwardNotFound,
	MQPublishFail.Code:             MQPublishFail,
	MQConsumeFail.Code:             MQConsumeFail,
	NotifyTaskNotExist.Code:        NotifyTaskNotExist,
	NotifyTaskFail.Code:            NotifyTaskFail,
	NotifyTaskRetryLimit.Code:      NotifyTaskRetryLimit,
}

// HTTPStatusForCode 将业务错误码映射为合适的 HTTP 状态码。
// 通用基础设施类（内部错误、MQ、通知）归为 5xx；
// 参数/鉴权/业务校验与冲突（S/G/L 各码）归为 4xx。
func HTTPStatusForCode(code string) int {
	switch code {
	case InternalError.Code, MQPublishFail.Code, MQConsumeFail.Code,
		NotifyTaskFail.Code, NotifyTaskRetryLimit.Code, NotifyTaskNotExist.Code:
		return http.StatusInternalServerError
	case RateLimitError.Code:
		return http.StatusTooManyRequests
	case Unauthorized.Code:
		return http.StatusUnauthorized
	case Forbidden.Code:
		return http.StatusForbidden
	case NotFoundError.Code,
		SeckillActivityNotExist.Code, SeckillOrderNotFound.Code,
		GroupBuyActivityNotExist.Code, GroupBuyOrderNotExist.Code, GroupBuyTeamNotExist.Code,
		LotteryActivityNotExist.Code, LotteryStrategyNotExist.Code, LotteryAwardNotFound.Code:
		return http.StatusNotFound
	case ParamError.Code, LockFailError.Code:
		return http.StatusBadRequest
	default:
		// 秒杀/拼团/抽奖等业务码属于客户端可预期的校验/冲突，返回 4xx
		return http.StatusBadRequest
	}
}

// WriteBizError 把业务错误安全地写回 HTTP 响应：
//   - 显式设置正确的 HTTP 状态码（不再一律 200），便于监控识别失败；
//   - 只返回与错误码对应的安全文案，绝不把 err.Error() 的内部细节泄露给客户端。
func WriteBizError(w http.ResponseWriter, err error) {
	code, info, status := resolveBizError(err)
	WriteJSON(w, status, Fail[any](code, info))
}

func resolveBizError(err error) (string, string, int) {
	if err == nil {
		return InternalError.Code, InternalError.Info, http.StatusInternalServerError
	}
	msg := err.Error()
	// 业务错误形如 "<CODE>: <信息>"，取前缀 CODE 反查安全文案。
	if idx := strings.Index(msg, ": "); idx > 0 {
		if rc, ok := codeRegistry[msg[:idx]]; ok {
			return rc.Code, rc.Info, HTTPStatusForCode(rc.Code)
		}
	}
	return InternalError.Code, InternalError.Info, http.StatusInternalServerError
}
