package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/pkg/common"
)

type LotteryService struct {
	raffleSvc *biz.RaffleService
}

func NewLotteryService(raffleSvc *biz.RaffleService) *LotteryService {
	return &LotteryService{raffleSvc: raffleSvc}
}

func (s *LotteryService) QueryLotteryActivityHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}

func (s *LotteryService) QueryLotteryStrategyHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}

func (s *LotteryService) RaffleHTTP(w http.ResponseWriter, r *http.Request) {
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

	result, err := s.raffleSvc.Raffle(r.Context(), req.ActivityId, req.UserId)
	if err != nil {
		resp := common.Fail[any]("500", err.Error())
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.Success(result)
	json.NewEncoder(w).Encode(resp)
}

func (s *LotteryService) QueryUserRaffleOrderHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := common.Success[any](nil)
	json.NewEncoder(w).Encode(resp)
}
