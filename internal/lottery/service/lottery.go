package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/pkg/auth"
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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	if req.ActivityId == "" {
		common.WriteError(w, http.StatusBadRequest, common.ParamError)
		return
	}
	// 身份只认鉴权中间件解析出的 user_id，不采信请求体自带字段，防止冒用他人身份抽奖。
	userID, ok := auth.UserID(r.Context())
	if !ok {
		common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
		return
	}

	result, err := s.raffleSvc.Raffle(r.Context(), req.ActivityId, userID)
	if err != nil {
		common.WriteBizError(w, err)
		return
	}

	common.WriteSuccess(w, result)
}

func (s *LotteryService) QueryUserRaffleOrderHTTP(w http.ResponseWriter, r *http.Request) {
	common.WriteSuccess[any](w, nil)
}
