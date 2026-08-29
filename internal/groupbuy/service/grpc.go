package service

import (
	"context"

	v1 "github.com/marketing-platform/api/groupbuy/v1"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/common"
)

// 以下方法实现 v1.GroupBuyServiceServer 接口（gRPC）。它们与 HTTP handler
// 共享同一套 biz 调用，只是入参/出参换成 protobuf 生成的类型。

func (s *GroupBuyService) QueryGroupBuyActivity(ctx context.Context, req *v1.QueryGroupBuyActivityReq) (*v1.QueryGroupBuyActivityResp, error) {
	// 活动查询在 HTTP 层同样是占位实现；这里回显 activity_id，保证 gRPC 接口可调用。
	return &v1.QueryGroupBuyActivityResp{ActivityId: req.GetActivityId()}, nil
}

func (s *GroupBuyService) TrialGroupBuyMarket(ctx context.Context, req *v1.TrialGroupBuyMarketReq) (*v1.TrialGroupBuyMarketResp, error) {
	if _, ok := auth.UserID(ctx); !ok {
		return nil, common.Unauthorized
	}
	result, err := s.trialSvc.TrialMarket(ctx, req.GetActivityId(), req.GetMarketOriginalPrice())
	if err != nil {
		return nil, err
	}
	return &v1.TrialGroupBuyMarketResp{
		ActivityId:           result.ActivityID,
		MarketPlan:           result.MarketPlan,
		MarketRule:           result.MarketRule,
		MarketDiscountAmount: result.MarketDiscountAmt,
		MarketPayAmount:      result.MarketPayAmount,
	}, nil
}

func (s *GroupBuyService) LockMarketPayOrder(ctx context.Context, req *v1.LockMarketPayOrderReq) (*v1.LockMarketPayOrderResp, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, common.Unauthorized
	}
	order, err := s.lockSvc.LockOrder(ctx, req.GetActivityId(), userID, req.GetChannel(), req.GetSource())
	if err != nil {
		return nil, err
	}
	return &v1.LockMarketPayOrderResp{
		OrderId:    order.OrderID,
		TeamId:     order.TeamID,
		OrderState: order.OrderState,
	}, nil
}

func (s *GroupBuyService) SettlementMarketPayOrder(ctx context.Context, req *v1.SettlementMarketPayOrderReq) (*v1.SettlementMarketPayOrderResp, error) {
	team, err := s.settlementSvc.Settlement(ctx, req.GetTeamId())
	if err != nil {
		return nil, err
	}
	return &v1.SettlementMarketPayOrderResp{
		TeamId:        team.TeamID,
		TeamState:     team.TeamState,
		CompleteCount: team.CompleteCount,
	}, nil
}

func (s *GroupBuyService) RefundMarketPayOrder(ctx context.Context, req *v1.RefundMarketPayOrderReq) (*v1.RefundMarketPayOrderResp, error) {
	order, err := s.refundSvc.Refund(ctx, req.GetOrderId())
	if err != nil {
		return nil, err
	}
	return &v1.RefundMarketPayOrderResp{
		OrderId:    order.OrderID,
		OrderState: order.OrderState,
	}, nil
}
