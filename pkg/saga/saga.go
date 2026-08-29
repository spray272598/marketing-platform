// Package saga 提供编排式（orchestration）Saga 分布式事务引擎。
//
// 适用场景：一次业务操作要跨越多个服务或资源（例如本地 DB + 远程库存服务），
// 无法用单个本地事务保证原子性。Saga 的做法是把操作拆成若干步骤，
// 每个步骤定义对应的反向补偿动作；顺序执行，任一步失败就逆序补偿已完成的步骤，
// 把系统恢复到一致状态。
//
// 与"手写 try/catch + 尽力回滚"相比，本引擎的关键差异：
//  1. 补偿失败不会被静默吞掉，而是汇总进 SagaError.CompErrors，
//     便于告警与人工介入（补偿失败意味着系统处于不一致状态，必须留痕）；
//  2. 步骤之间可通过 Bag 传递数据（例如第一步算出的成团人数给后续步骤用）；
//  3. 补偿使用 context.WithoutCancel，避免主流程 ctx 取消导致补偿也被取消；
//  4. SagaLog 是可插拔接口（依赖倒置），可接入日志、审计或崩溃恢复。
package saga

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// 步骤执行的两种动作，用于错误归类。
const (
	OpAction     = "action"
	OpCompensate = "compensate"
)

// ---------- 步骤间数据传递 ----------

// Bag 在 Saga 的各个步骤之间传递中间结果，并发安全。
type Bag struct {
	mu   sync.RWMutex
	data map[string]any
}

func newBag() *Bag { return &Bag{data: make(map[string]any)} }

// Set 存放一个中间结果。
func (b *Bag) Set(key string, value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[key] = value
}

// Get 读取一个中间结果。
func (b *Bag) Get(key string) (any, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.data[key]
	return v, ok
}

// GetInt32 便捷读取 int32，缺失或类型不符时返回 (0, false)。
func (b *Bag) GetInt32(key string) (int32, bool) {
	v, ok := b.Get(key)
	if !ok {
		return 0, false
	}
	n, ok := v.(int32)
	return n, ok
}

// GetString 便捷读取 string，缺失或类型不符时返回 ("", false)。
func (b *Bag) GetString(key string) (string, bool) {
	v, ok := b.Get(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GetBool 便捷读取 bool，缺失或类型不符时返回 (false, false)。
func (b *Bag) GetBool(key string) (bool, bool) {
	v, ok := b.Get(key)
	if !ok {
		return false, false
	}
	b2, ok := v.(bool)
	return b2, ok
}

type ctxKey int

const (
	bagKey ctxKey = iota
	sagaIDKey
)

// WithBag 把 Bag 写入 context。
func WithBag(ctx context.Context, b *Bag) context.Context {
	return context.WithValue(ctx, bagKey, b)
}

// BagFrom 取出 context 中的 Bag，不存在时返回 nil。
func BagFrom(ctx context.Context) *Bag {
	if b, ok := ctx.Value(bagKey).(*Bag); ok {
		return b
	}
	return nil
}

// WithID 把 Saga 实例 ID 写入 context。
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sagaIDKey, id)
}

// IDFrom 取出当前 Saga 的实例 ID，用于串联日志与幂等键。
func IDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(sagaIDKey).(string); ok {
		return id
	}
	return ""
}

// ---------- 步骤与错误 ----------

// Step 定义 Saga 中的一个步骤。
// Action 是正向动作；Compensate 是反向补偿，为 nil 表示无需补偿。
// 两者都应尽量保证幂等（Saga 本身不重试，但上游可能重放整个流程）。
type Step struct {
	Name       string
	Action     func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// StepError 记录某个步骤（或其补偿）的失败详情。
type StepError struct {
	Step string // 步骤名
	Op   string // OpAction 或 OpCompensate
	Err  error
}

func (e StepError) Error() string {
	return fmt.Sprintf("saga: %s of step %q failed: %v", e.Op, e.Step, e.Err)
}

// Unwrap 支持 errors.Is/As 追溯根因。
func (e StepError) Unwrap() error { return e.Err }

// SagaError 汇总一次失败的 Saga 执行。
// CompErrors 非空意味着补偿失败、系统可能处于不一致状态，必须告警并人工介入。
type SagaError struct {
	Step        string // 触发失败的步骤
	Err         error  // 原始错误
	Compensated []string
	CompErrors  []StepError
}

func (e *SagaError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("saga: step %q failed: %v", e.Step, e.Err))
	if len(e.Compensated) > 0 {
		sb.WriteString("; compensated: " + strings.Join(e.Compensated, ", "))
	}
	if len(e.CompErrors) > 0 {
		msgs := make([]string, 0, len(e.CompErrors))
		for _, ce := range e.CompErrors {
			msgs = append(msgs, ce.Error())
		}
		sb.WriteString("; COMPENSATION FAILURES: " + strings.Join(msgs, "; "))
	}
	return sb.String()
}

// Unwrap 让调用方能直接判定业务根因（例如库存不足）。
func (e *SagaError) Unwrap() error { return e.Err }

// ---------- 执行日志（依赖倒置） ----------

// SagaLog 是可插拔的执行审计接口，便于接入监控或崩溃恢复。
// 只保留最关键的四个事件，避免实现负担过重。
type SagaLog interface {
	// OnStart 在 Saga 开始执行时调用。
	OnStart(ctx context.Context, sagaID string)
	// OnStepFailed 在某一步的正向动作失败时调用。
	OnStepFailed(ctx context.Context, sagaID, step string, err error)
	// OnCompensateFailed 在某一步的补偿动作失败时调用（需要人工介入）。
	OnCompensateFailed(ctx context.Context, sagaID, step string, err error)
	// OnAborted 在 Saga 失败并结束补偿后调用。
	OnAborted(ctx context.Context, sagaID string, err error)
}

// noopLog 是默认实现，什么都不做。
type noopLog struct{}

func (noopLog) OnStart(context.Context, string)                          {}
func (noopLog) OnStepFailed(context.Context, string, string, error)      {}
func (noopLog) OnCompensateFailed(context.Context, string, string, error) {}
func (noopLog) OnAborted(context.Context, string, error)                 {}

// SlogLog 是基于 slog 的开箱即用实现。
type SlogLog struct{ logger *slog.Logger }

// NewSlogLog 创建一个基于 slog 的审计日志实现。logger 为 nil 时使用 slog.Default()。
func NewSlogLog(logger *slog.Logger) *SlogLog {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogLog{logger: logger}
}

func (l *SlogLog) OnStart(ctx context.Context, sagaID string) {
	l.logger.Info("saga started", slog.String("saga_id", sagaID))
}
func (l *SlogLog) OnStepFailed(ctx context.Context, sagaID, step string, err error) {
	l.logger.Error("saga step failed",
		slog.String("saga_id", sagaID), slog.String("step", step), slog.Any("error", err))
}
func (l *SlogLog) OnCompensateFailed(ctx context.Context, sagaID, step string, err error) {
	l.logger.Error("saga compensation FAILED, manual intervention required",
		slog.String("saga_id", sagaID), slog.String("step", step), slog.Any("error", err))
}
func (l *SlogLog) OnAborted(ctx context.Context, sagaID string, err error) {
	l.logger.Error("saga aborted", slog.String("saga_id", sagaID), slog.Any("error", err))
}

// ---------- 协调器 ----------

// Coordinator 顺序执行步骤，失败时逆序补偿。
type Coordinator struct {
	steps  []Step
	logger *slog.Logger
	log    SagaLog
}

// New 创建一个 Saga 协调器，可传入初始步骤。
func New(steps ...Step) *Coordinator {
	return &Coordinator{
		steps:  append([]Step(nil), steps...),
		logger: slog.Default(),
		log:    noopLog{},
	}
}

// Add 追加步骤，返回自身以支持链式调用。
func (c *Coordinator) Add(step Step) *Coordinator {
	c.steps = append(c.steps, step)
	return c
}

// WithLogger 设置用于过程日志的 slog.Logger。
func (c *Coordinator) WithLogger(l *slog.Logger) *Coordinator {
	if l != nil {
		c.logger = l
	}
	return c
}

// WithLog 设置审计日志实现。
func (c *Coordinator) WithLog(l SagaLog) *Coordinator {
	if l != nil {
		c.log = l
	}
	return c
}

// Run 顺序执行所有步骤。
// 全部成功返回 nil；任一步失败则逆序补偿已完成的步骤，并返回 *SagaError
// （可用 errors.Is/As 判定根因，用 CompErrors 判断是否需人工介入）。
func (c *Coordinator) Run(ctx context.Context) error {
	sagaID := uuid.New().String()
	ctx = WithID(ctx, sagaID)
	ctx = WithBag(ctx, newBag())

	c.log.OnStart(ctx, sagaID)
	c.logger.Info("saga running",
		slog.String("saga_id", sagaID), slog.Int("steps", len(c.steps)))

	done := make([]Step, 0, len(c.steps))
	for _, step := range c.steps {
		// 上下文已取消时不再继续推进，直接进入补偿。
		if err := ctx.Err(); err != nil {
			c.logger.Warn("saga context cancelled before step",
				slog.String("saga_id", sagaID), slog.String("step", step.Name))
			return c.abort(ctx, sagaID, step.Name, err, done)
		}

		if err := step.Action(ctx); err != nil {
			c.log.OnStepFailed(ctx, sagaID, step.Name, err)
			return c.abort(ctx, sagaID, step.Name, err, done)
		}
		c.logger.Debug("saga step done",
			slog.String("saga_id", sagaID), slog.String("step", step.Name))
		done = append(done, step)
	}

	c.logger.Info("saga completed", slog.String("saga_id", sagaID))
	return nil
}

// abort 汇总失败信息并执行补偿。
func (c *Coordinator) abort(ctx context.Context, sagaID, stepName string, err error, done []Step) error {
	compensated, compErrors := c.compensate(ctx, sagaID, done)
	sErr := &SagaError{Step: stepName, Err: err, Compensated: compensated, CompErrors: compErrors}
	c.log.OnAborted(ctx, sagaID, sErr)
	c.logger.Error("saga aborted",
		slog.String("saga_id", sagaID),
		slog.String("failed_step", stepName),
		slog.Any("error", err),
		slog.Int("compensated", len(compensated)),
		slog.Int("compensation_failures", len(compErrors)),
	)
	return sErr
}

// compensate 逆序补偿已完成的步骤。
//
// 单个补偿失败不会中断其余补偿（尽力回滚），但会被收集并返回，
// 保证"补偿失败"这件事不被吞掉。
func (c *Coordinator) compensate(ctx context.Context, sagaID string, done []Step) ([]string, []StepError) {
	compensated := make([]string, 0, len(done))
	var compErrors []StepError

	for i := len(done) - 1; i >= 0; i-- {
		step := done[i]
		if step.Compensate == nil {
			continue
		}
		// 关键：补偿脱离主 ctx 的取消链，否则主流程超时/取消会连带取消补偿，
		// 导致"该回滚的没回滚"，系统停留在不一致状态。
		compCtx := context.WithoutCancel(ctx)

		if err := step.Compensate(compCtx); err != nil {
			compErrors = append(compErrors, StepError{Step: step.Name, Op: OpCompensate, Err: err})
			c.log.OnCompensateFailed(ctx, sagaID, step.Name, err)
			continue
		}
		compensated = append(compensated, step.Name)
		c.logger.Info("saga step compensated",
			slog.String("saga_id", sagaID), slog.String("step", step.Name))
	}
	return compensated, compErrors
}
