package service

import (
	"context"

	pb "github.com/marketing-platform/api/seckill/v1"
	"github.com/marketing-platform/internal/seckill/biz"
)

type SeckillService struct {
	pb.UnimplementedSeckillServiceServer
	tradeSvc *biz.TradeService
}

func NewSeckillService(tradeSvc *biz.TradeService) *SeckillService {
	return &SeckillService{tradeSvc: tradeSvc}
}

func (s *SeckillService) QuerySeckillActivity(ctx context.Context, req *pb.QuerySeckillActivityReq) (*pb.QuerySeckillActivityResp, error) {
	// TODO: implement
	return &pb.QuerySeckillActivityResp{}, nil
}

func (s *SeckillService) CreateSeckillOrder(ctx context.Context, req *pb.CreateSeckillOrderReq) (*pb.CreateSeckillOrderResp, error) {
	order, err := s.tradeSvc.CreateSeckillOrder(ctx, req.ActivityId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateSeckillOrderResp{
		OrderId:     order.OrderID,
		ActivityId:  order.ActivityID,
		OrderState:  order.OrderState,
		OrderTime:   order.OrderTime,
	}, nil
}

func (s *SeckillService) QuerySeckillOrder(ctx context.Context, req *pb.QuerySeckillOrderReq) (*pb.QuerySeckillOrderResp, error) {
	order, err := s.tradeSvc.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &pb.QuerySeckillOrderResp{
		OrderId:    order.OrderID,
		ActivityId: order.ActivityID,
		UserId:     order.UserID,
		SkuId:      order.SkuID,
		OrderState: order.OrderState,
		OrderTime:  order.OrderTime,
		PayTime:    order.PayTime,
	}, nil
}
