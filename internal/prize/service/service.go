package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/prize/biz"
	"github.com/marketing-platform/pkg/common"
)

type PrizeService struct {
	prizeSvc *biz.PrizeService
}

func NewPrizeService(prizeSvc *biz.PrizeService) *PrizeService {
	return &PrizeService{prizeSvc: prizeSvc}
}

func (s *PrizeService) DeductStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		PrizeID string `json:"prize_id"`
		Count   int32  `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	ok, err := s.prizeSvc.DeductStock(r.Context(), req.PrizeID, req.Count)
	if err != nil || !ok {
		json.NewEncoder(w).Encode(common.Fail[any](common.InternalError.Code, err.Error()))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"prize_id": req.PrizeID, "deducted": req.Count}))
}

func (s *PrizeService) GetStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	prizeID := r.URL.Query().Get("prize_id")
	if prizeID == "" {
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	stock, err := s.prizeSvc.GetStock(r.Context(), prizeID)
	if err != nil {
		json.NewEncoder(w).Encode(common.Fail[any](common.NotFoundError.Code, err.Error()))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"prize_id": prizeID, "stock": stock}))
}

func (s *PrizeService) RestoreStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		PrizeID string `json:"prize_id"`
		Count   int32  `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	if err := s.prizeSvc.RestoreStock(r.Context(), req.PrizeID, req.Count); err != nil {
		json.NewEncoder(w).Encode(common.Fail[any](common.InternalError.Code, err.Error()))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"prize_id": req.PrizeID, "restored": req.Count}))
}
