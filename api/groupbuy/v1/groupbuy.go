package groupbuyv1

import (
	"context"
	"google.golang.org/grpc"
)

type GroupBuyServiceServer interface {
	QueryGroupBuyActivity(context.Context, *QueryGroupBuyActivityReq) (*QueryGroupBuyActivityResp, error)
	TrialGroupBuyMarket(context.Context, *TrialGroupBuyMarketReq) (*TrialGroupBuyMarketResp, error)
	LockMarketPayOrder(context.Context, *LockMarketPayOrderReq) (*LockMarketPayOrderResp, error)
	SettlementMarketPayOrder(context.Context, *SettlementMarketPayOrderReq) (*SettlementMarketPayOrderResp, error)
	RefundMarketPayOrder(context.Context, *RefundMarketPayOrderReq) (*RefundMarketPayOrderResp, error)
}

type QueryGroupBuyActivityReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
}

type QueryGroupBuyActivityResp struct {
	ActivityId    string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	ActivityName  string `protobuf:"bytes,2,opt,name=activity_name,json=activityName,proto3" json:"activity_name,omitempty"`
	DiscountId    string `protobuf:"bytes,3,opt,name=discount_id,json=discountId,proto3" json:"discount_id,omitempty"`
	GroupType     int32  `protobuf:"varint,4,opt,name=group_type,json=groupType,proto3" json:"group_type,omitempty"`
	TargetCount   int32  `protobuf:"varint,5,opt,name=target_count,json=targetCount,proto3" json:"target_count,omitempty"`
	ActivityState int32  `protobuf:"varint,6,opt,name=activity_state,json=activityState,proto3" json:"activity_state,omitempty"`
}

type TrialGroupBuyMarketReq struct {
	ActivityId           string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId               int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	MarketOriginalPrice  int32  `protobuf:"varint,3,opt,name=market_original_price,json=marketOriginalPrice,proto3" json:"market_original_price,omitempty"`
	Channel              string `protobuf:"bytes,4,opt,name=channel,proto3" json:"channel,omitempty"`
	Source               string `protobuf:"bytes,5,opt,name=source,proto3" json:"source,omitempty"`
}

type TrialGroupBuyMarketResp struct {
	ActivityId          string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	MarketPlan          int32  `protobuf:"varint,2,opt,name=market_plan,json=marketPlan,proto3" json:"market_plan,omitempty"`
	MarketRule          string `protobuf:"bytes,3,opt,name=market_rule,json=marketRule,proto3" json:"market_rule,omitempty"`
	MarketDiscountAmount int32 `protobuf:"varint,4,opt,name=market_discount_amount,json=marketDiscountAmount,proto3" json:"market_discount_amount,omitempty"`
	MarketPayAmount     int32  `protobuf:"varint,5,opt,name=market_pay_amount,json=marketPayAmount,proto3" json:"market_pay_amount,omitempty"`
}

type LockMarketPayOrderReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId     int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Channel    string `protobuf:"bytes,3,opt,name=channel,proto3" json:"channel,omitempty"`
	Source     string `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
}

type LockMarketPayOrderResp struct {
	OrderId    string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	TeamId     string `protobuf:"bytes,2,opt,name=team_id,json=teamId,proto3" json:"team_id,omitempty"`
	OrderState int32  `protobuf:"varint,3,opt,name=order_state,json=orderState,proto3" json:"order_state,omitempty"`
}

type SettlementMarketPayOrderReq struct {
	TeamId string `protobuf:"bytes,1,opt,name=team_id,json=teamId,proto3" json:"team_id,omitempty"`
}

type SettlementMarketPayOrderResp struct {
	TeamId        string `protobuf:"bytes,1,opt,name=team_id,json=teamId,proto3" json:"team_id,omitempty"`
	TeamState     int32  `protobuf:"varint,2,opt,name=team_state,json=teamState,proto3" json:"team_state,omitempty"`
	CompleteCount int32  `protobuf:"varint,3,opt,name=complete_count,json=completeCount,proto3" json:"complete_count,omitempty"`
}

type RefundMarketPayOrderReq struct {
	OrderId string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	UserId  int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type RefundMarketPayOrderResp struct {
	OrderId    string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	OrderState int32  `protobuf:"varint,2,opt,name=order_state,json=orderState,proto3" json:"order_state,omitempty"`
}

func RegisterGroupBuyServiceServer(s *grpc.Server, srv GroupBuyServiceServer) {}

type UnimplementedGroupBuyServiceServer struct{}

func (UnimplementedGroupBuyServiceServer) QueryGroupBuyActivity(context.Context, *QueryGroupBuyActivityReq) (*QueryGroupBuyActivityResp, error) {
	return nil, nil
}

func (UnimplementedGroupBuyServiceServer) TrialGroupBuyMarket(context.Context, *TrialGroupBuyMarketReq) (*TrialGroupBuyMarketResp, error) {
	return nil, nil
}

func (UnimplementedGroupBuyServiceServer) LockMarketPayOrder(context.Context, *LockMarketPayOrderReq) (*LockMarketPayOrderResp, error) {
	return nil, nil
}

func (UnimplementedGroupBuyServiceServer) SettlementMarketPayOrder(context.Context, *SettlementMarketPayOrderReq) (*SettlementMarketPayOrderResp, error) {
	return nil, nil
}

func (UnimplementedGroupBuyServiceServer) RefundMarketPayOrder(context.Context, *RefundMarketPayOrderReq) (*RefundMarketPayOrderResp, error) {
	return nil, nil
}
