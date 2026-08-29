package service

import (
	"context"
	"encoding/json"
	"net/http"

	v1 "github.com/marketing-platform/api/seckill/v1"
	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/common"
)

type SeckillService struct {
	v1.UnimplementedSeckillServiceServer
	tradeSvc    *biz.TradeService
	activityRepo biz.ActivityRepo
}

func NewSeckillService(tradeSvc *biz.TradeService, activityRepo biz.ActivityRepo) *SeckillService {
	return &SeckillService{
		tradeSvc:    tradeSvc,
		activityRepo: activityRepo,
	}
}

func (s *SeckillService) QuerySeckillActivity(ctx context.Context, req *v1.QuerySeckillActivityReq) (*v1.QuerySeckillActivityResp, error) {
	activity, err := s.activityRepo.GetActivity(ctx, req.GetActivityId())
	if err != nil {
		return nil, err
	}
	return &v1.QuerySeckillActivityResp{
		ActivityId:    activity.ActivityID,
		ActivityName:  activity.ActivityName,
		SkuId:         activity.SkuID,
		TotalCount:    activity.TotalCount,
		LimitCount:    activity.LimitCount,
		ActivityState: activity.ActivityState,
		StartTime:     activity.StartTime,
		EndTime:       activity.EndTime,
	}, nil
}

func (s *SeckillService) CreateSeckillOrder(ctx context.Context, req *v1.CreateSeckillOrderReq) (*v1.CreateSeckillOrderResp, error) {
	// 身份只认认证信息里解析出的 user_id，不采信调用方传入的 req.UserId。
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, common.Unauthorized
	}
	order, err := s.tradeSvc.CreateSeckillOrder(ctx, req.GetActivityId(), userID)
	if err != nil {
		return nil, err
	}
	return &v1.CreateSeckillOrderResp{
		OrderId:    order.OrderID,
		ActivityId: order.ActivityID,
		OrderState: order.OrderState,
		OrderTime:  order.OrderTime,
	}, nil
}

func (s *SeckillService) QuerySeckillOrder(ctx context.Context, req *v1.QuerySeckillOrderReq) (*v1.QuerySeckillOrderResp, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, common.Unauthorized
	}
	order, err := s.tradeSvc.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, common.SeckillOrderNotFound
	}
	// 归属校验：只能查询自己的订单。
	if order.UserID != userID {
		return nil, common.Forbidden
	}
	return &v1.QuerySeckillOrderResp{
		OrderId:     order.OrderID,
		ActivityId:  order.ActivityID,
		UserId:      order.UserID,
		SkuId:       order.SkuID,
		OrderState:  order.OrderState,
		OrderTime:   order.OrderTime,
		PayTime:     order.PayTime,
	}, nil
}

// HTTP handler wrappers for backward compatibility with existing gateway
func (s *SeckillService) QuerySeckillActivityHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	activityID := r.URL.Query().Get("activity_id")
	if activityID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	activity, err := s.activityRepo.GetActivity(r.Context(), activityID)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}
	common.WriteSuccess(w, activity)
}

func (s *SeckillService) CreateSeckillOrderHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		ActivityId string `json:"activity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.ActivityId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	// 身份只认鉴权中间件解析出的 user_id，不采信请求体自带字段，防止冒用他人身份下单。
	userID, ok := auth.UserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
		return
	}
	order, err := s.tradeSvc.CreateSeckillOrder(r.Context(), req.ActivityId, userID)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}
	common.WriteSuccess(w, order)
}

func (s *SeckillService) QuerySeckillOrderHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	userID, ok := auth.UserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
		return
	}
	order, err := s.tradeSvc.GetOrder(r.Context(), orderID)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}
	// repo 可能在查不到时返回 (nil, nil)，先判空再做归属校验，避免空指针。
	if order == nil {
		common.WriteError(w, http.StatusNotFound, common.SeckillOrderNotFound)
		return
	}
	// 归属校验：只能查询自己的订单，避免仅凭 order_id 越权查看他人订单。
	if order.UserID != userID {
		common.WriteError(w, http.StatusForbidden, common.Forbidden)
		return
	}
	common.WriteSuccess(w, order)
}
