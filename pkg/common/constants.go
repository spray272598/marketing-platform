package common

const (
	ActivityStateInit    int32 = 0
	ActivityStateOpen    int32 = 1
	ActivityStateClosed  int32 = 2

	OrderStateInit       int32 = 0
	OrderStatePaid       int32 = 1
	OrderStateCancelled  int32 = 2
	OrderStateTimeout    int32 = 3

	TeamStateBuilding    int32 = 0
	TeamStateSuccess     int32 = 1
	TeamStateFailed      int32 = 2

	DiscountTypeZJ       = "ZJ"
	DiscountTypeMJ       = "MJ"
	DiscountTypeN        = "N"
	DiscountTypeZK       = "ZK"

	RaffleRuleBlacklist  = "rule_blacklist"
	RaffleRuleWeight     = "rule_weight"
)
