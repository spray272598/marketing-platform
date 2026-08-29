package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/pkg/auth"
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
	common.WriteSuccess[any](w, nil)
}

func (s *GroupBuyService) TrialGroupBuyMarketHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		ActivityId          string `json:"activity_id"`
		MarketOriginalPrice int32  `json:"market_original_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.ActivityId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	// 试算本身不按用户区分，但仍要求已认证，避免接口被匿名刷用。
	if _, ok := auth.UserID(r.Context()); !ok {
		common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
		return
	}

	result, err := s.trialSvc.TrialMarket(r.Context(), req.ActivityId, req.MarketOriginalPrice)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

func (s *GroupBuyService) LockMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		ActivityId string `json:"activity_id"`
		Channel    string `json:"channel"`
		Source     string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.ActivityId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	// 身份只认鉴权中间件解析出的 user_id，不采信请求体自带字段，防止冒用他人身份锁单。
	userID, ok := auth.UserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
		return
	}

	order, err := s.lockSvc.LockOrder(r.Context(), req.ActivityId, userID, req.Channel, req.Source)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, order)
}

func (s *GroupBuyService) SettlementMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		TeamId string `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.TeamId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}

	team, err := s.settlementSvc.Settlement(r.Context(), req.TeamId)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, team)
}

func (s *GroupBuyService) RefundMarketPayOrderHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		OrderId string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.OrderId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}

	order, err := s.refundSvc.Refund(r.Context(), req.OrderId)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, order)
}
