package saga

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

// recorder 记录步骤与补偿的执行顺序。
type recorder struct {
	mu      sync.Mutex
	events  []string
	failOn  map[string]error // 正向动作按步骤名注入失败
	compErr map[string]error // 补偿动作按步骤名注入失败
}

func newRecorder() *recorder {
	return &recorder{failOn: map[string]error{}, compErr: map[string]error{}}
}

func (r *recorder) add(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

func (r *recorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.events))
	copy(out, r.events)
	return out
}

// step 构造一个记录执行情况且可注入失败的步骤。
func (r *recorder) step(name string) Step {
	return Step{
		Name: name,
		Action: func(ctx context.Context) error {
			r.add("do:" + name)
			return r.failOn[name]
		},
		Compensate: func(ctx context.Context) error {
			r.add("undo:" + name)
			return r.compErr[name]
		},
	}
}

func TestRun_AllStepsSucceed(t *testing.T) {
	r := newRecorder()
	c := New(r.step("s1"), r.step("s2"), r.step("s3"))

	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	got := r.snapshot()
	want := []string{"do:s1", "do:s2", "do:s3"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestRun_CompensatesInReverseOrder(t *testing.T) {
	r := newRecorder()
	boom := errors.New("boom")
	r.failOn["s3"] = boom

	c := New(r.step("s1"), r.step("s2"), r.step("s3"))
	err := c.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var sErr *SagaError
	if !errors.As(err, &sErr) {
		t.Fatalf("expected *SagaError, got %T", err)
	}
	if sErr.Step != "s3" || !errors.Is(sErr.Err, boom) {
		t.Fatalf("unexpected saga error: %+v", sErr)
	}

	// s3 失败后不应执行；已完成的 s2、s1 必须逆序补偿。
	got := r.snapshot()
	want := []string{"do:s1", "do:s2", "do:s3", "undo:s2", "undo:s1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if strings.Join(sErr.Compensated, ",") != "s2,s1" {
		t.Fatalf("expected compensated [s2 s1], got %v", sErr.Compensated)
	}
	if len(sErr.CompErrors) != 0 {
		t.Fatalf("expected no compensation errors, got %v", sErr.CompErrors)
	}
}

func TestRun_FirstStepFails_NoCompensation(t *testing.T) {
	r := newRecorder()
	r.failOn["s1"] = errors.New("immediate")

	c := New(r.step("s1"), r.step("s2"))
	if err := c.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	// 首步即失败，无需补偿，也不应继续推进后续步骤。
	got := r.snapshot()
	if strings.Join(got, ",") != "do:s1" {
		t.Fatalf("expected only [do:s1], got %v", got)
	}
}

// TestRun_CompensationFailureIsCollected 补偿失败不能中断其余补偿，
// 且必须汇总进 CompErrors 以便告警/人工介入。
func TestRun_CompensationFailureIsCollected(t *testing.T) {
	r := newRecorder()
	r.failOn["s3"] = errors.New("boom")
	r.compErr["s2"] = errors.New("undo failed")

	c := New(r.step("s1"), r.step("s2"), r.step("s3"))
	err := c.Run(context.Background())

	var sErr *SagaError
	if !errors.As(err, &sErr) {
		t.Fatalf("expected *SagaError, got %T", err)
	}
	// s2 补偿失败，但 s1 的补偿仍必须执行。
	if strings.Join(sErr.Compensated, ",") != "s1" {
		t.Fatalf("expected s1 compensated despite s2 failure, got %v", sErr.Compensated)
	}
	if len(sErr.CompErrors) != 1 || sErr.CompErrors[0].Step != "s2" {
		t.Fatalf("expected one compensation error on s2, got %v", sErr.CompErrors)
	}
	if sErr.CompErrors[0].Op != OpCompensate {
		t.Fatalf("expected OpCompensate, got %s", sErr.CompErrors[0].Op)
	}

	got := r.snapshot()
	want := []string{"do:s1", "do:s2", "do:s3", "undo:s2", "undo:s1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// TestCompensateUsesUncancellableContext 主流程 ctx 取消时，补偿仍必须能执行。
func TestCompensateUsesUncancellableContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var compCtxErr error
	var compensated bool
	c := New(
		Step{
			Name: "s1",
			Action: func(ctx context.Context) error {
				return nil
			},
			Compensate: func(c context.Context) error {
				compensated = true
				compCtxErr = c.Err() // 补偿拿到的 ctx 不应已取消
				return nil
			},
		},
		Step{
			Name: "s2",
			Action: func(ctx context.Context) error {
				cancel() // 模拟主流程中途取消
				return errors.New("boom")
			},
		},
	)

	if err := c.Run(ctx); err == nil {
		t.Fatal("expected error")
	}
	if !compensated {
		t.Fatal("expected compensation to run even after cancellation")
	}
	if compCtxErr != nil {
		t.Fatalf("compensation should receive an uncancelled context, got %v", compCtxErr)
	}
}

// TestRun_BagPassesDataBetweenSteps 验证步骤间可通过 Bag 传递中间结果。
func TestRun_BagPassesDataBetweenSteps(t *testing.T) {
	var seen int32
	var id string
	c := New(
		Step{
			Name: "produce",
			Action: func(ctx context.Context) error {
				BagFrom(ctx).Set("complete_count", int32(7))
				id = IDFrom(ctx)
				return nil
			},
		},
		Step{
			Name: "consume",
			Action: func(ctx context.Context) error {
				v, ok := BagFrom(ctx).GetInt32("complete_count")
				if !ok {
					t.Error("expected complete_count in bag")
				}
				seen = v
				return nil
			},
		},
	)
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seen != 7 {
		t.Fatalf("expected 7, got %d", seen)
	}
	if id == "" {
		t.Fatal("expected a non-empty saga id")
	}
}

// TestRun_StepWithoutCompensate 允许某些步骤无需补偿（例如纯本地读取）。
func TestRun_StepWithoutCompensate(t *testing.T) {
	r := newRecorder()
	r.failOn["s2"] = errors.New("boom")

	noComp := Step{
		Name:   "s1",
		Action: func(ctx context.Context) error { r.add("do:s1"); return nil },
	}
	c := New(noComp, r.step("s2"))

	if err := c.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	got := r.snapshot()
	if strings.Join(got, ",") != "do:s1,do:s2" {
		t.Fatalf("expected [do:s1 do:s2], got %v", got)
	}
}

// TestAdd_Chainable 验证 Add 可链式追加步骤。
func TestAdd_Chainable(t *testing.T) {
	r := newRecorder()
	c := New().Add(r.step("s1")).Add(r.step("s2"))
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.snapshot()) != 2 {
		t.Fatalf("expected 2 steps, got %v", r.snapshot())
	}
}

// TestSlogLog_DoesNotPanic 确保内置审计实现可正常工作。
func TestSlogLog_DoesNotPanic(t *testing.T) {
	r := newRecorder()
	r.failOn["s2"] = errors.New("boom")
	r.compErr["s1"] = errors.New("undo failed")

	c := New(r.step("s1"), r.step("s2")).WithLog(NewSlogLog(nil))
	if err := c.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
