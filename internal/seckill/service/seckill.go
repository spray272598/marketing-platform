package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/pkg/common"
)

type SeckillService struct {
	tradeSvc *biz.TradeService
}

func NewSeckillService(tradeSvc *biz.TradeService) *SeckillService {
	return &SeckillService{tradeSvc: tradeSvc}
}

func (s *SeckillService) QuerySeckillActivityHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: implement
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}

func (s *SeckillService) CreateSeckillOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ActivityId string `json:"activity_id"`
		UserId     int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.Fail[any]("400", "参数错误")
		json.NewEncoder(w).Encode(resp)
		return
	}

	order, err := s.tradeSvc.CreateSeckillOrder(r.Context(), req.ActivityId, req.UserId)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(order)
	json.NewEncoder(w).Encode(resp)
}

func (s *SeckillService) QuerySeckillOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: implement
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}
