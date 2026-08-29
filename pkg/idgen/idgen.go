// Package idgen 实现美团 Leaf「号段模式」的分布式 ID 分配。
//
// 核心思想：每个业务标签(bizTag)从 SegmentStore 批量领取一段号
// （如 [1,1000)）缓存在进程内存中顺序发放；段用尽时再向 SegmentStore
// 原子推进 max_id 领取下一段。这样取号几乎不依赖中心存储，且 ID 单调递增。
// 服务重启会丢弃当前段内未发放的号，因此单号可能存在空洞（业务订单号可接受）。
package idgen

import (
	"context"
	"fmt"
	"sync"
)

// SegmentStore 负责在中心存储上原子地推进某个业务标签的 max_id，
// 并返回推进后的新最大值。典型实现是 MySQL（SELECT .. FOR UPDATE + UPDATE）。
type SegmentStore interface {
	// AllocMax 原子地将 bizTag 对应的 max_id 增加 step，并返回新的 max_id。
	AllocMax(ctx context.Context, bizTag string, step int64) (int64, error)
}

// Generator 是号段模式的内存发放器，线程安全，按 bizTag 维护各自的号段。
type Generator struct {
	store SegmentStore
	step  int64

	mu  sync.Mutex
	cur map[string]int64 // 当前已发放到的值
	max map[string]int64 // 当前段的上界（含）
}

// NewGenerator 创建号段生成器。step 必须 > 0，否则使用默认 1000。
func NewGenerator(store SegmentStore, step int64) *Generator {
	if step <= 0 {
		step = 1000
	}
	return &Generator{
		store: store,
		step:  step,
		cur:   make(map[string]int64),
		max:   make(map[string]int64),
	}
}

// Next 返回 bizTag 的下一个 ID（从 1 开始，单调递增）。
func (g *Generator) Next(ctx context.Context, bizTag string) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 当前段内还有余量，直接递增返回。
	if g.cur[bizTag] < g.max[bizTag] {
		g.cur[bizTag]++
		return g.cur[bizTag], nil
	}

	// 段已用尽（或首次），向中心存储领取新段。
	newMax, err := g.store.AllocMax(ctx, bizTag, g.step)
	if err != nil {
		return 0, fmt.Errorf("idgen: alloc segment for %q: %w", bizTag, err)
	}
	g.max[bizTag] = newMax
	g.cur[bizTag] = newMax - g.step + 1
	return g.cur[bizTag], nil
}
