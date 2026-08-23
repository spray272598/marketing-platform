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
	StartTime int64
	EndTime   int64
	Tags      map[string]string
	Logs      []string
	Children  []*TraceContext
	mu        sync.Mutex
}

type TraceCollector struct {
	serviceName string
	traces      map[string]*TraceContext
	mu          sync.RWMutex
	logger      *slog.Logger
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

func NewTraceCollector(serviceName string, logger *slog.Logger) *TraceCollector {
	return &TraceCollector{
		serviceName: serviceName,
		traces:      make(map[string]*TraceContext),
		logger:      logger,
	}
}

func (tc *TraceCollector) StartTrace(ctx context.Context, spanName string) *TraceContext {
	traceID := getOrCreateTraceID(ctx)
	spanID := uuid.New().String()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	trace := &TraceContext{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   tc.serviceName,
		StartTime: time.Now().UnixMilli(),
		Tags:      make(map[string]string),
	}
	tc.traces[traceID] = trace

	if tc.logger != nil {
		tc.logger.Debug("trace started",
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
			slog.String("span_name", spanName),
		)
	}

	return trace
}

func (tc *TraceCollector) EndTrace(trace *TraceContext) {
	trace.EndTime = time.Now().UnixMilli()

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	delete(tc.traces, trace.TraceID)

	duration := trace.EndTime - trace.StartTime
	if tc.logger != nil {
		tc.logger.Info("trace completed",
			slog.String("trace_id", trace.TraceID),
			slog.String("span_id", trace.SpanID),
			slog.Int64("duration_ms", duration),
		)
	}
}

func (tc *TraceCollector) AddTag(trace *TraceContext, key, value string) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	trace.Tags[key] = value
}

func (tc *TraceCollector) AddLog(trace *TraceContext, log string) {
	trace.mu.Lock()
	defer trace.mu.Unlock()
	trace.Logs = append(trace.Logs, log)
}

func getOrCreateTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return uuid.New().String()
}

func FormatTrace(trace *TraceContext) string {
	return fmt.Sprintf("Trace{id=%s, span=%s, service=%s, duration=%dms, tags=%v, logs=%d}",
		trace.TraceID, trace.SpanID, trace.Service,
		trace.EndTime-trace.StartTime, trace.Tags, len(trace.Logs))
}
