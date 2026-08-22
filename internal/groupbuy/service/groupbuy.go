package service

import (
	"context"

	pb "github.com/marketing-platform/api/groupbuy/v1"
	"github.com/marketing-platform/internal/groupbuy/biz"
)

type GroupBuyService struct {
	pb.UnimplementedGroupBuyServiceServer
	trialSvc      *biz.TrialService
	lockSvc       *biz.LockService
	settlementSvc *biz.SettlementService
	refundSvc     *biz.RefundService
}

func NewGroupBuyService(
	trialSvc *biz.TrialService,
	lockSvc *biz.LockService,
	settlementSvc *biz.SettlementService,
	refundSvc *biz.RefundService,
) *GroupBuyService {
	return &GroupBuyService{
		trialSvc:      trialSvc,
		lockSvc:       lockSvc,
		settlementSvc: settlementSvc,
		refundSvc:     refundSvc,
	}
}

func (s *GroupBuyService) QueryGroupBuyActivity(ctx context.Context, req *pb.QueryGroupBuyActivityReq) (*pb.QueryGroupBuyActivityResp, error) {
	return &pb.QueryGroupBuyActivityResp{}, nil
}

func (s *GroupBuyService) TrialGroupBuyMarket(ctx context.Context, req *pb.TrialGroupBuyMarketReq) (*pb.TrialGroupBuyMarketResp, error) {
	result, err := s.trialSvc.TrialMarket(ctx, req.ActivityId, req.MarketOriginalPrice)
	if err != nil {
		return nil, err
	}
	return &pb.TrialGroupBuyMarketResp{
		ActivityId:           result.ActivityID,
		MarketPlan:           0,
		MarketRule:           result.MarketRule,
		MarketDiscountAmount: result.MarketDiscountAmt,
		MarketPayAmount:      result.MarketPayAmount,
	}, nil
}

func (s *GroupBuyService) LockMarketPayOrder(ctx context.Context, req *pb.LockMarketPayOrderReq) (*pb.LockMarketPayOrderResp, error) {
	order, err := s.lockSvc.LockOrder(ctx, req.ActivityId, req.UserId, req.Channel, req.Source)
	if err != nil {
		return nil, err
	}
	return &pb.LockMarketPayOrderResp{
		OrderId:    order.OrderID,
		TeamId:     order.TeamID,
		OrderState: order.OrderState,
	}, nil
}

func (s *GroupBuyService) SettlementMarketPayOrder(ctx context.Context, req *pb.SettlementMarketPayOrderReq) (*pb.SettlementMarketPayOrderResp, error) {
	team, err := s.settlementSvc.Settlement(ctx, req.TeamId)
	if err != nil {
		return nil, err
	}
	return &pb.SettlementMarketPayOrderResp{
		TeamId:        team.TeamID,
		TeamState:     team.TeamState,
		CompleteCount: team.CompleteCount,
	}, nil
}

func (s *GroupBuyService) RefundMarketPayOrder(ctx context.Context, req *pb.RefundMarketPayOrderReq) (*pb.RefundMarketPayOrderResp, error) {
	order, err := s.refundSvc.Refund(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &pb.RefundMarketPayOrderResp{
		OrderId:    order.OrderID,
		OrderState: order.OrderState,
	}, nil
}
