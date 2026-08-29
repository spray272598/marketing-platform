package idgen

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// fakeStore 是 SegmentStore 的测试替身：每次 AllocMax 把内部 max 增加 step 并返回。
type fakeStore struct {
	mu    sync.Mutex
	calls int
	max   int64
	fail  bool
	failN int // 第几次调用故意失败（0 表示不失败）
}

func (f *fakeStore) AllocMax(ctx context.Context, bizTag string, step int64) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.fail && f.calls == f.failN {
		return 0, errors.New("boom")
	}
	f.max += step
	return f.max, nil
}

func TestNext_SequentialWithinSegment(t *testing.T) {
	fs := &fakeStore{}
	g := NewGenerator(fs, 1000)

	prev := int64(0)
	for i := 1; i <= 1000; i++ {
		id, err := g.Next(context.Background(), "order")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != int64(i) {
			t.Fatalf("expected %d, got %d", i, id)
		}
		if id <= prev {
			t.Fatalf("id not increasing: %d <= %d", id, prev)
		}
		prev = id
	}
	// 段内发放不应触发第二次分配
	if fs.calls != 1 {
		t.Fatalf("expected 1 store call, got %d", fs.calls)
	}
}

func TestNext_RefillOnExhaust(t *testing.T) {
	fs := &fakeStore{}
	g := NewGenerator(fs, 10)

	for i := 1; i <= 25; i++ {
		id, err := g.Next(context.Background(), "order")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != int64(i) {
			t.Fatalf("expected %d, got %d", i, id)
		}
	}
	// 25 个 ID，step=10 → 第 1、11、21 次各触发一次分配，共 3 次
	if fs.calls != 3 {
		t.Fatalf("expected 3 store calls, got %d", fs.calls)
	}
}

func TestNext_StoreError(t *testing.T) {
	fs := &fakeStore{fail: true, failN: 1}
	g := NewGenerator(fs, 1000)
	if _, err := g.Next(context.Background(), "order"); err == nil {
		t.Fatal("expected error from store")
	}
}

func TestNext_Concurrent(t *testing.T) {
	fs := &fakeStore{}
	g := NewGenerator(fs, 1000)

	const n = 5000
	ids := make([]int64, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			id, err := g.Next(context.Background(), "order")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			ids[i] = id
		}(i)
	}
	wg.Wait()

	seen := make(map[int64]struct{}, n)
	for _, id := range ids {
		if id == 0 {
			t.Fatal("zero id returned")
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != n {
		t.Fatalf("expected %d unique ids, got %d", n, len(seen))
	}
}
