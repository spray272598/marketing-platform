package lotteryv1

import (
	"context"
	"google.golang.org/grpc"
)

type LotteryServiceServer interface {
	QueryLotteryActivity(context.Context, *QueryLotteryActivityReq) (*QueryLotteryActivityResp, error)
	QueryLotteryStrategy(context.Context, *QueryLotteryStrategyReq) (*QueryLotteryStrategyResp, error)
	Raffle(context.Context, *RaffleReq) (*RaffleResp, error)
	QueryUserRaffleOrder(context.Context, *QueryUserRaffleOrderReq) (*QueryUserRaffleOrderResp, error)
}

type QueryLotteryActivityReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
}

type QueryLotteryActivityResp struct {
	ActivityId    string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	ActivityName  string `protobuf:"bytes,2,opt,name=activity_name,json=activityName,proto3" json:"activity_name,omitempty"`
	StrategyId    string `protobuf:"bytes,3,opt,name=strategy_id,json=strategyId,proto3" json:"strategy_id,omitempty"`
	ActivityState int32  `protobuf:"varint,4,opt,name=activity_state,json=activityState,proto3" json:"activity_state,omitempty"`
}

type QueryLotteryStrategyReq struct {
	StrategyId string `protobuf:"bytes,1,opt,name=strategy_id,json=strategyId,proto3" json:"strategy_id,omitempty"`
}

type StrategyAward struct {
	AwardId    string  `protobuf:"bytes,1,opt,name=award_id,json=awardId,proto3" json:"award_id,omitempty"`
	AwardName  string  `protobuf:"bytes,2,opt,name=award_name,json=awardName,proto3" json:"award_name,omitempty"`
	AwardRate  float64 `protobuf:"fixed64,3,opt,name=award_rate,json=awardRate,proto3" json:"award_rate,omitempty"`
	AwardCount int32   `protobuf:"varint,4,opt,name=award_count,json=awardCount,proto3" json:"award_count,omitempty"`
}

type StrategyRule struct {
	RuleModel string `protobuf:"bytes,1,opt,name=rule_model,json=ruleModel,proto3" json:"rule_model,omitempty"`
	RuleValue string `protobuf:"bytes,2,opt,name=rule_value,json=ruleValue,proto3" json:"rule_value,omitempty"`
}

type QueryLotteryStrategyResp struct {
	StrategyId string           `protobuf:"bytes,1,opt,name=strategy_id,json=strategyId,proto3" json:"strategy_id,omitempty"`
	Awards     []*StrategyAward `protobuf:"bytes,2,rep,name=awards,proto3" json:"awards,omitempty"`
	Rules      []*StrategyRule  `protobuf:"bytes,3,rep,name=rules,proto3" json:"rules,omitempty"`
}

type RaffleReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId     int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type RaffleResp struct {
	AwardId    string `protobuf:"bytes,1,opt,name=award_id,json=awardId,proto3" json:"award_id,omitempty"`
	AwardName  string `protobuf:"bytes,2,opt,name=award_name,json=awardName,proto3" json:"award_name,omitempty"`
	AwardState int32  `protobuf:"varint,3,opt,name=award_state,json=awardState,proto3" json:"award_state,omitempty"`
	AwardTime  string `protobuf:"bytes,4,opt,name=award_time,json=awardTime,proto3" json:"award_time,omitempty"`
}

type QueryUserRaffleOrderReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId     int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type QueryUserRaffleOrderResp struct {
	TotalCount int32 `protobuf:"varint,1,opt,name=total_count,json=totalCount,proto3" json:"total_count,omitempty"`
	DayCount   int32 `protobuf:"varint,2,opt,name=day_count,json=dayCount,proto3" json:"day_count,omitempty"`
	MonthCount int32 `protobuf:"varint,3,opt,name=month_count,json=monthCount,proto3" json:"month_count,omitempty"`
	UsedCount  int32 `protobuf:"varint,4,opt,name=used_count,json=usedCount,proto3" json:"used_count,omitempty"`
}

func RegisterLotteryServiceServer(s *grpc.Server, srv LotteryServiceServer) {}

type UnimplementedLotteryServiceServer struct{}

func (UnimplementedLotteryServiceServer) QueryLotteryActivity(context.Context, *QueryLotteryActivityReq) (*QueryLotteryActivityResp, error) {
	return nil, nil
}

func (UnimplementedLotteryServiceServer) QueryLotteryStrategy(context.Context, *QueryLotteryStrategyReq) (*QueryLotteryStrategyResp, error) {
	return nil, nil
}

func (UnimplementedLotteryServiceServer) Raffle(context.Context, *RaffleReq) (*RaffleResp, error) {
	return nil, nil
}

func (UnimplementedLotteryServiceServer) QueryUserRaffleOrder(context.Context, *QueryUserRaffleOrderReq) (*QueryUserRaffleOrderResp, error) {
	return nil, nil
}
