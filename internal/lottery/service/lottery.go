package service

import (
	"context"

	pb "github.com/marketing-platform/api/lottery/v1"
	"github.com/marketing-platform/internal/lottery/biz"
)

type LotteryService struct {
	pb.UnimplementedLotteryServiceServer
	raffleSvc *biz.RaffleService
}

func NewLotteryService(raffleSvc *biz.RaffleService) *LotteryService {
	return &LotteryService{raffleSvc: raffleSvc}
}

func (s *LotteryService) QueryLotteryActivity(ctx context.Context, req *pb.QueryLotteryActivityReq) (*pb.QueryLotteryActivityResp, error) {
	return &pb.QueryLotteryActivityResp{}, nil
}

func (s *LotteryService) QueryLotteryStrategy(ctx context.Context, req *pb.QueryLotteryStrategyReq) (*pb.QueryLotteryStrategyResp, error) {
	return &pb.QueryLotteryStrategyResp{}, nil
}

func (s *LotteryService) Raffle(ctx context.Context, req *pb.RaffleReq) (*pb.RaffleResp, error) {
	result, err := s.raffleSvc.Raffle(ctx, req.ActivityId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.RaffleResp{
		AwardId:    result.AwardID,
		AwardName:  result.AwardName,
		AwardState: result.AwardState,
		AwardTime:  result.AwardTime,
	}, nil
}

func (s *LotteryService) QueryUserRaffleOrder(ctx context.Context, req *pb.QueryUserRaffleOrderReq) (*pb.QueryUserRaffleOrderResp, error) {
	return &pb.QueryUserRaffleOrderResp{}, nil
}
