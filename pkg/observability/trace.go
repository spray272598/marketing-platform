package observability

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Service   string
	SpanName  string
	StartTime int64
	EndTime   int64
	Tags      map[string]string
	Logs      []string
	Children  []*TraceContext
	logger    *slog.Logger
	mu        sync.Mutex
}

func NewTraceContext(traceID, spanID, parentID, service, spanName string, logger *slog.Logger) *TraceContext {
	return &TraceContext{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Service:   service,
		SpanName:  spanName,
		StartTime: time.Now().UnixMilli(),
		Tags:      make(map[string]string),
		Logs:      make([]string, 0),
		Children:  make([]*TraceContext, 0),
		logger:    logger,
	}
}

func (t *TraceContext) Finish() {
	t.EndTime = time.Now().UnixMilli()
	duration := t.EndTime - t.StartTime

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.logger != nil {
		t.logger.Info("span completed",
			slog.String("trace_id", t.TraceID),
			slog.String("span_id", t.SpanID),
			slog.String("parent_id", t.ParentID),
			slog.String("span_name", t.SpanName),
			slog.String("service", t.Service),
			slog.Int64("duration_ms", duration),
			slog.Any("tags", t.Tags),
		)
	}
}

func (t *TraceContext) AddTag(key, value string) *TraceContext {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Tags[key] = value
	return t
}

func (t *TraceContext) AddLog(msg string) *TraceContext {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Logs = append(t.Logs, msg)
	return t
}

func (t *TraceContext) ChildSpan(ctx context.Context, spanName string) *TraceContext {
	child := NewTraceContext(t.TraceID, uuid.New().String(), t.SpanID, t.Service, spanName, t.logger)
	t.mu.Lock()
	t.Children = append(t.Children, child)
	t.mu.Unlock()
	return child
}

func (t *TraceContext) DurationMs() int64 {
	if t.EndTime == 0 {
		return time.Now().UnixMilli() - t.StartTime
	}
	return t.EndTime - t.StartTime
}

func (t *TraceContext) String() string {
	return fmt.Sprintf("Trace{trace=%s, span=%s, parent=%s, name=%s, svc=%s, dur=%dms}",
		t.TraceID, t.SpanID, t.ParentID, t.SpanName, t.Service, t.DurationMs())
}

type TraceCollector struct {
	serviceName string
	traces      map[string]*TraceContext
	mu          sync.RWMutex
	logger      *slog.Logger
	metrics     *Metrics
}

var globalCollector *TraceCollector
var once sync.Once

func GetTraceCollector() *TraceCollector {
	once.Do(func() {
		globalCollector = &TraceCollector{
			serviceName: "default",
			traces:      make(map[string]*TraceContext),
			logger:      slog.Default(),
		}
	})
	return globalCollector
}

func NewTraceCollector(serviceName string, logger *slog.Logger, metrics *Metrics) *TraceCollector {
	return &TraceCollector{
		serviceName: serviceName,
		traces:      make(map[string]*TraceContext),
		logger:      logger,
		metrics:     metrics,
	}
}

func (tc *TraceCollector) StartRootSpan(ctx context.Context, spanName string) *TraceContext {
	traceID := getOrCreateTraceID(ctx)
	spanID := uuid.New().String()

	root := NewTraceContext(traceID, spanID, "", tc.serviceName, spanName, tc.logger)

	tc.mu.Lock()
	tc.traces[traceID] = root
	tc.mu.Unlock()

	if tc.logger != nil {
		tc.logger.Debug("root span started",
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
			slog.String("span_name", spanName),
			slog.String("service", tc.serviceName),
		)
	}

	// 记录trace启动指标
	if tc.metrics != nil {
		tc.metrics.RequestTotal.WithLabelValues("trace", spanName).Inc()
	}

	return root
}

func (tc *TraceCollector) CompleteTrace(trace *TraceContext) {
	trace.Finish()

	tc.mu.RLock()
	delete(tc.traces, trace.TraceID)
	tc.mu.RUnlock()

	// 记录trace完成指标
	if tc.metrics != nil {
		duration := float64(trace.DurationMs()) / 1000.0
		tc.metrics.RequestDuration.WithLabelValues("trace", trace.SpanName).Observe(duration)
		tc.metrics.RequestStatus.WithLabelValues("trace", trace.SpanName, "200").Inc()
	}
}

func (tc *TraceCollector) GetActiveTraces() map[string]*TraceContext {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	result := make(map[string]*TraceContext)
	for k, v := range tc.traces {
		result[k] = v
	}
	return result
}

func (tc *TraceCollector) GetTraceByID(traceID string) (*TraceContext, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	t, ok := tc.traces[traceID]
	return t, ok
}

func (tc *TraceCollector) InjectTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

func (tc *TraceCollector) ExtractTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

func getOrCreateTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return uuid.New().String()
}

func FormatTrace(trace *TraceContext) string {
	return fmt.Sprintf("Trace{id=%s, span=%s, service=%s, duration=%dms, tags=%v, logs=%d, children=%d}",
		trace.TraceID, trace.SpanID, trace.Service,
		trace.DurationMs(), trace.Tags, len(trace.Logs), len(trace.Children))
}