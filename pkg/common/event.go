package common

type Event struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Timestamp int64  `json:"timestamp"`
	BizID     string `json:"biz_id"`
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
	CreateAt     string `json:"create_at"`
	UpdateAt     string `json:"update_at"`
}

var (
	NotifyStatusInitCode    = ResponseCode{Code: "N0001", Info: "初始状态"}
	NotifyStatusSuccessCode = ResponseCode{Code: "N0002", Info: "通知成功"}
	NotifyStatusRetryCode   = ResponseCode{Code: "N0003", Info: "通知重试"}
	NotifyStatusFailedCode  = ResponseCode{Code: "N0004", Info: "通知失败"}
)
