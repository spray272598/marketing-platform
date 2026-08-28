package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewNotifyService,
	NewLockService,
	NewTrialService,
	NewSettlementService,
	NewRefundService,
	NewNotifyConsumer,
)
