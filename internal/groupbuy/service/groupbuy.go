package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/pkg/common"
)

type GroupBuyService struct {
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

func (s *GroupBuyService) QueryGroupBuyActivityHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}

func (s *GroupBuyService) TrialGroupBuyMarketHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ActivityId          string `json:"activity_id"`
		UserId              int64  `json:"user_id"`
		MarketOriginalPrice int32  `json:"market_original_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.Fail[any]("400", "参数错误")
		json.NewEncoder(w).Encode(resp)
		return
	}

	result, err := s.trialSvc.TrialMarket(r.Context(), req.ActivityId, req.MarketOriginalPrice)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(result)
	json.NewEncoder(w).Encode(resp)
}

func (s *GroupBuyService) LockMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ActivityId string `json:"activity_id"`
		UserId     int64  `json:"user_id"`
		Channel    string `json:"channel"`
		Source     string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.Fail[any]("400", "参数错误")
		json.NewEncoder(w).Encode(resp)
		return
	}

	order, err := s.lockSvc.LockOrder(r.Context(), req.ActivityId, req.UserId, req.Channel, req.Source)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(order)
	json.NewEncoder(w).Encode(resp)
}

func (s *GroupBuyService) SettlementMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		TeamId string `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.Fail[any]("400", "参数错误")
		json.NewEncoder(w).Encode(resp)
		return
	}

	team, err := s.settlementSvc.Settlement(r.Context(), req.TeamId)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(team)
	json.NewEncoder(w).Encode(resp)
}

func (s *GroupBuyService) RefundMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		OrderId string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.Fail[any]("400", "参数错误")
		json.NewEncoder(w).Encode(resp)
		return
	}

	order, err := s.refundSvc.Refund(r.Context(), req.OrderId)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(order)
	json.NewEncoder(w).Encode(resp)
}
