package biz

type GroupBuyActivity struct {
	ID            int64  `json:"id"`
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	DiscountID    string `json:"discount_id"`
	GroupType     int32  `json:"group_type"`
	TargetCount   int32  `json:"target_count"`
	ValidTime     int32  `json:"valid_time"`
	TagID         string `json:"tag_id"`
	ActivityState int32  `json:"activity_state"`
}

type GroupBuyDiscount struct {
	ID          int64  `json:"id"`
	DiscountID  string `json:"discount_id"`
	MarketPlan  string `json:"market_plan"`
	MarketExpr  string `json:"market_expr"`
}

type GroupBuyOrder struct {
	ID          int64  `json:"id"`
	OrderID     string `json:"order_id"`
	TeamID      string `json:"team_id"`
	UserID      int64  `json:"user_id"`
	ActivityID  string `json:"activity_id"`
	BizID       string `json:"biz_id"`
	OrderState  int32  `json:"order_state"`
	CreateAt    string `json:"create_at"`
}

type GroupBuyTeam struct {
	ID            int64  `json:"id"`
	TeamID        string `json:"team_id"`
	ActivityID    string `json:"activity_id"`
	TargetCount   int32  `json:"target_count"`
	CompleteCount int32  `json:"complete_count"`
	LockCount     int32  `json:"lock_count"`
	TeamState     int32  `json:"team_state"`
}

type NotifyTask struct {
	ID           int64  `json:"id"`
	TaskID       string `json:"task_id"`
	NotifyType   string `json:"notify_type"`
	NotifyStatus int32  `json:"notify_status"`
	NotifyURL    string `json:"notify_url"`
	NotifyData   string `json:"notify_data"`
	UUID         string `json:"uuid"`
	RetryCount   int32  `json:"retry_count"`
	MaxRetry     int32  `json:"max_retry"`
	NextTime     int64  `json:"next_time"`
}

const (
	NotifyStatusInit    int32 = 0
	NotifyStatusSuccess int32 = 1
	NotifyStatusRetry   int32 = 2
	NotifyStatusFailed  int32 = 3

	NotifyTypeHTTP = "HTTP"
	NotifyTypeMQ   = "MQ"
)
