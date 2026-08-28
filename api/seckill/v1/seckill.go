package seckillv1

import "context"

type SeckillServiceServer interface {
	QuerySeckillActivity(context.Context, *QuerySeckillActivityReq) (*QuerySeckillActivityResp, error)
	CreateSeckillOrder(context.Context, *CreateSeckillOrderReq) (*CreateSeckillOrderResp, error)
	QuerySeckillOrder(context.Context, *QuerySeckillOrderReq) (*QuerySeckillOrderResp, error)
}

type UnimplementedSeckillServiceServer struct{}

func (UnimplementedSeckillServiceServer) QuerySeckillActivity(context.Context, *QuerySeckillActivityReq) (*QuerySeckillActivityResp, error) {
	return nil, nil
}

func (UnimplementedSeckillServiceServer) CreateSeckillOrder(context.Context, *CreateSeckillOrderReq) (*CreateSeckillOrderResp, error) {
	return nil, nil
}

func (UnimplementedSeckillServiceServer) QuerySeckillOrder(context.Context, *QuerySeckillOrderReq) (*QuerySeckillOrderResp, error) {
	return nil, nil
}

type QuerySeckillActivityReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
}

type QuerySeckillActivityResp struct {
	ActivityId    string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	ActivityName  string `protobuf:"bytes,2,opt,name=activity_name,json=activityName,proto3" json:"activity_name,omitempty"`
	SkuId         string `protobuf:"bytes,3,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
	TotalCount    int32  `protobuf:"varint,4,opt,name=total_count,json=totalCount,proto3" json:"total_count,omitempty"`
	LimitCount    int32  `protobuf:"varint,5,opt,name=limit_count,json=limitCount,proto3" json:"limit_count,omitempty"`
	ActivityState int32  `protobuf:"varint,6,opt,name=activity_state,json=activityState,proto3" json:"activity_state,omitempty"`
	StartTime     string `protobuf:"bytes,7,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime       string `protobuf:"bytes,8,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
}

type CreateSeckillOrderReq struct {
	ActivityId string `protobuf:"bytes,1,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId     int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type CreateSeckillOrderResp struct {
	OrderId    string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	ActivityId string `protobuf:"bytes,2,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	OrderState int32  `protobuf:"varint,3,opt,name=order_state,json=orderState,proto3" json:"order_state,omitempty"`
	OrderTime  string `protobuf:"bytes,4,opt,name=order_time,json=orderTime,proto3" json:"order_time,omitempty"`
}

type QuerySeckillOrderReq struct {
	OrderId string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	UserId  int64  `protobuf:"varint,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type QuerySeckillOrderResp struct {
	OrderId    string `protobuf:"bytes,1,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	ActivityId string `protobuf:"bytes,2,opt,name=activity_id,json=activityId,proto3" json:"activity_id,omitempty"`
	UserId     int64  `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	SkuId      string `protobuf:"bytes,4,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
	OrderState int32  `protobuf:"varint,5,opt,name=order_state,json=orderState,proto3" json:"order_state,omitempty"`
	OrderTime  string `protobuf:"bytes,6,opt,name=order_time,json=orderTime,proto3" json:"order_time,omitempty"`
	PayTime    string `protobuf:"bytes,7,opt,name=pay_time,json=payTime,proto3" json:"pay_time,omitempty"`
}

func (x *QuerySeckillActivityReq) GetActivityId() string {
	if x != nil {
		return x.ActivityId
	}
	return ""
}

func (x *CreateSeckillOrderReq) GetActivityId() string {
	if x != nil {
		return x.ActivityId
	}
	return ""
}

func (x *CreateSeckillOrderReq) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *QuerySeckillOrderReq) GetOrderId() string {
	if x != nil {
		return x.OrderId
	}
	return ""
}
