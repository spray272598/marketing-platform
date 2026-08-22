package biz

type SeckillActivity struct {
	ID            int64  `json:"id"`
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	SkuID         string `json:"sku_id"`
	TotalCount    int32  `json:"total_count"`
	LimitCount    int32  `json:"limit_count"`
	ActivityState int32  `json:"activity_state"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
}

type SeckillOrder struct {
	ID           int64  `json:"id"`
	OrderID      string `json:"order_id"`
	ActivityID   string `json:"activity_id"`
	UserID       int64  `json:"user_id"`
	SkuID        string `json:"sku_id"`
	OrderState   int32  `json:"order_state"`
	OrderTime    string `json:"order_time"`
	PayTime      string `json:"pay_time"`
}

type SeckillStock struct {
	ActivityID string `json:"activity_id"`
	StockCount int32  `json:"stock_count"`
}
