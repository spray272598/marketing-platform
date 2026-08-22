package biz

type LotteryActivity struct {
	ID            int64  `json:"id"`
	ActivityID    string `json:"activity_id"`
	ActivityName  string `json:"activity_name"`
	StrategyID    string `json:"strategy_id"`
	ActivityState int32  `json:"activity_state"`
}

type LotteryStrategy struct {
	ID         int64  `json:"id"`
	StrategyID string `json:"strategy_id"`
	RuleModels string `json:"rule_models"`
}

type StrategyAward struct {
	ID        int64   `json:"id"`
	AwardID   string  `json:"award_id"`
	AwardName string  `json:"award_name"`
	AwardRate float64 `json:"award_rate"`
	AwardCount int32  `json:"award_count"`
	RuleModels string `json:"rule_models"`
}

type LotteryOrder struct {
	ID          int64  `json:"id"`
	OrderID     string `json:"order_id"`
	ActivityID  string `json:"activity_id"`
	UserID      int64  `json:"user_id"`
	AwardID     string `json:"award_id"`
	AwardState  int32  `json:"award_state"`
	AwardTime   string `json:"award_time"`
}
