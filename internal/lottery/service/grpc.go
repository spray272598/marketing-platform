package service

import (
	"context"

	v1 "github.com/marketing-platform/api/lottery/v1"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/common"
)

// 以下方法实现 v1.LotteryServiceServer 接口（gRPC），与 HTTP handler 共享 biz 调用。

func (s *LotteryService) QueryLotteryActivity(ctx context.Context, req *v1.QueryLotteryActivityReq) (*v1.QueryLotteryActivityResp, error) {
	return &v1.QueryLotteryActivityResp{ActivityId: req.GetActivityId()}, nil
}

func (s *LotteryService) QueryLotteryStrategy(ctx context.Context, req *v1.QueryLotteryStrategyReq) (*v1.QueryLotteryStrategyResp, error) {
	return &v1.QueryLotteryStrategyResp{StrategyId: req.GetStrategyId()}, nil
}

func (s *LotteryService) Raffle(ctx context.Context, req *v1.RaffleReq) (*v1.RaffleResp, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, common.Unauthorized
	}
	result, err := s.raffleSvc.Raffle(ctx, req.GetActivityId(), userID)
	if err != nil {
		return nil, err
	}
	return &v1.RaffleResp{
		AwardId:    result.AwardID,
		AwardName:  result.AwardName,
		AwardState: result.AwardState,
		AwardTime:  result.AwardTime,
	}, nil
}

func (s *LotteryService) QueryUserRaffleOrder(ctx context.Context, req *v1.QueryUserRaffleOrderReq) (*v1.QueryUserRaffleOrderResp, error) {
	return &v1.QueryUserRaffleOrderResp{}, nil
}
