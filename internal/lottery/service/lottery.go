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
	common.WriteSuccess[any](w, nil)
}

func (s *LotteryService) QueryLotteryStrategyHTTP(w http.ResponseWriter, r *http.Request) {
	common.WriteSuccess[any](w, nil)
}

func (s *LotteryService) RaffleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteError(w, http.StatusMethodNotAllowed, common.ParamError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req struct {
		ActivityId string `json:"activity_id"`
		UserId     int64  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.ActivityId == "" || req.UserId <= 0 {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}

	result, err := s.raffleSvc.Raffle(r.Context(), req.ActivityId, req.UserId)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

func (s *LotteryService) QueryUserRaffleOrderHTTP(w http.ResponseWriter, r *http.Request) {
	common.WriteSuccess[any](w, nil)
}
