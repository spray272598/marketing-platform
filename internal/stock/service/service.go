package service

import (
	"encoding/json"
	"net/http"

	"github.com/marketing-platform/internal/stock/biz"
	"github.com/marketing-platform/pkg/common"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewStockService)

type StockService struct {
	stockSvc *biz.StockService
}

func NewStockService(stockSvc *biz.StockService) *StockService {
	return &StockService{stockSvc: stockSvc}
}

func (s *StockService) DeductStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, "method not allowed"))
		return
	}
	var req struct {
		StockKey string `json:"stock_key"`
		Count    int32  `json:"count"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	if req.StockKey == "" || req.Count <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, "stock_key required, count must be positive"))
		return
	}
	ok, err := s.stockSvc.DeductStock(r.Context(), req.StockKey, req.Count)
	if err != nil || !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.Fail[any](common.InternalError.Code, common.InternalError.Info))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"stock_key": req.StockKey, "deducted": req.Count}))
}

func (s *StockService) GetStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stockKey := r.URL.Query().Get("stock_key")
	if stockKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	stock, err := s.stockSvc.GetStock(r.Context(), stockKey)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.Fail[any](common.NotFoundError.Code, common.NotFoundError.Info))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"stock_key": stockKey, "stock": stock}))
}

func (s *StockService) RestoreStockHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, "method not allowed"))
		return
	}
	var req struct {
		StockKey string `json:"stock_key"`
		Count    int32  `json:"count"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, common.ParamError.Info))
		return
	}
	if req.StockKey == "" || req.Count <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.Fail[any](common.ParamError.Code, "stock_key required, count must be positive"))
		return
	}
	if err := s.stockSvc.RestoreStock(r.Context(), req.StockKey, req.Count); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.Fail[any](common.InternalError.Code, common.InternalError.Info))
		return
	}
	json.NewEncoder(w).Encode(common.Success(map[string]interface{}{"stock_key": req.StockKey, "restored": req.Count}))
}
