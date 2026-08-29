package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/marketing-platform/pkg/common"
)

// stubStockClient 记录调用次数并支持注入失败，用于验证 Saga 的补偿路径。
type stubStockClient struct {
	deducted   int
	restored   int
	deductErr  error
	restoreErr error
}

func (s *stubStockClient) DeductStock(ctx context.Context, stockKey string, count int32) error {
	if s.deductErr != nil {
		return s.deductErr
	}
	s.deducted += int(count)
	return nil
}

func (s *stubStockClient) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	if s.restoreErr != nil {
		return s.restoreErr
	}
	s.restored += int(count)
	return nil
}

// errTeamRepo 包装 mockTeamRepo，用于注入 CompleteTeam 失败。
type errTeamRepo struct {
	*mockTeamRepo
	completeErr error
}

func (r *errTeamRepo) CompleteTeam(ctx context.Context, teamID string, targetCount, successState int32) (int32, bool, error) {
	if r.completeErr != nil {
		return 0, false, r.completeErr
	}
	return r.mockTeamRepo.CompleteTeam(ctx, teamID, targetCount, successState)
}

func newSettlementFixture(target int32) (*SettlementService, *mockTeamRepo, *mockNotifyTaskRepo, *stubStockClient) {
	teamRepo := newMockTeamRepo()
	teamRepo.CreateTeam(context.Background(), &GroupBuyTeam{
		TeamID:        "team_001",
		ActivityID:    "act_001",
		TargetCount:   target,
		CompleteCount: 0,
		TeamState:     common.TeamStateBuilding,
	})
	notifyRepo := newMockNotifyTaskRepo()
	notifySvc := NewNotifyService(notifyRepo, newMockGBMQRepo())
	stockClient := &stubStockClient{}
	svc := NewSettlementService(teamRepo, newMockGBOrderRepo(), notifySvc, stockClient)
	return svc, teamRepo, notifyRepo, stockClient
}

// TestSettlement_CompletePersistsStateAndNotifiesOnce 验证核心修复：
// 达标时团队状态必须落库（原实现只改内存对象），且成团通知只创建一次。
func TestSettlement_CompletePersistsStateAndNotifiesOnce(t *testing.T) {
	svc, teamRepo, notifyRepo, stock := newSettlementFixture(2)

	// 第一次结算：人数未满，不应成团、不应通知。
	team, err := svc.Settlement(context.Background(), "team_001")
	if err != nil {
		t.Fatalf("first settlement failed: %v", err)
	}
	if team.CompleteCount != 1 {
		t.Fatalf("expected complete_count 1, got %d", team.CompleteCount)
	}
	if team.TeamState != common.TeamStateBuilding {
		t.Fatalf("expected team still building, got %d", team.TeamState)
	}
	if len(notifyRepo.tasks) != 0 {
		t.Fatalf("expected no notify task before team completes, got %d", len(notifyRepo.tasks))
	}

	// 第二次结算：达标成团，状态必须持久化，通知创建一次。
	team, err = svc.Settlement(context.Background(), "team_001")
	if err != nil {
		t.Fatalf("second settlement failed: %v", err)
	}
	if team.CompleteCount != 2 {
		t.Fatalf("expected complete_count 2, got %d", team.CompleteCount)
	}
	if team.TeamState != common.TeamStateSuccess {
		t.Fatalf("expected team state SUCCESS persisted, got %d", team.TeamState)
	}
	// 验证确实写回了"仓储"（原 bug：状态只停留在内存对象上）。
	stored, _ := teamRepo.GetTeam(context.Background(), "team_001")
	if stored.TeamState != common.TeamStateSuccess {
		t.Fatalf("expected stored team state SUCCESS, got %d", stored.TeamState)
	}
	if len(notifyRepo.tasks) != 1 {
		t.Fatalf("expected exactly 1 notify task, got %d", len(notifyRepo.tasks))
	}
	if stock.deducted != 2 {
		t.Fatalf("expected 2 stock deductions, got %d", stock.deducted)
	}
}

// TestSettlement_IdempotentAfterSuccess 成团后重复结算必须幂等：
// 既不重复扣库存，也不重复创建成团通知，更不会让完成数继续增长。
func TestSettlement_IdempotentAfterSuccess(t *testing.T) {
	svc, teamRepo, notifyRepo, stock := newSettlementFixture(1)

	for i := 0; i < 3; i++ {
		if _, err := svc.Settlement(context.Background(), "team_001"); err != nil {
			t.Fatalf("settlement %d failed: %v", i+1, err)
		}
	}

	if stock.deducted != 1 {
		t.Fatalf("expected exactly 1 stock deduction across repeated settlements, got %d", stock.deducted)
	}
	if len(notifyRepo.tasks) != 1 {
		t.Fatalf("expected exactly 1 notify task across repeated settlements, got %d", len(notifyRepo.tasks))
	}
	stored, _ := teamRepo.GetTeam(context.Background(), "team_001")
	if stored.CompleteCount != 1 {
		t.Fatalf("expected complete_count to stay at target 1, got %d", stored.CompleteCount)
	}
	if stored.TeamState != common.TeamStateSuccess {
		t.Fatalf("expected team state SUCCESS, got %d", stored.TeamState)
	}
}

// TestSettlement_CompensatesStockWhenTeamNotSettleable 库存已扣但本地推进失败时，
// Saga 必须把库存补偿回去，避免"库存少了但订单没成"。
func TestSettlement_CompensatesStockWhenTeamNotSettleable(t *testing.T) {
	svc, teamRepo, _, stock := newSettlementFixture(2)

	// 让第二步（本地自增）失败，触发逆序补偿。
	// 注意包装的是 fixture 里的 teamRepo，保证 GetTeam 仍能查到团队。
	svc.teamRepo = &errTeamRepo{mockTeamRepo: teamRepo, completeErr: ErrTeamNotSettleable}

	_, err := svc.Settlement(context.Background(), "team_001")
	if err == nil {
		t.Fatal("expected settlement to fail")
	}
	if stock.deducted != 1 {
		t.Fatalf("expected stock deducted once, got %d", stock.deducted)
	}
	if stock.restored != 1 {
		t.Fatalf("expected stock restored (compensated) once, got %d", stock.restored)
	}
}

// TestSettlement_NotifyFailureRollsBackTeamState 成团通知创建失败时，
// 团队必须退回"进行中"，这样后续重试可以重新成团并补发通知。
func TestSettlement_NotifyFailureRollsBackTeamState(t *testing.T) {
	svc, teamRepo, notifyRepo, stock := newSettlementFixture(1)
	notifyRepo.createErr = errors.New("notify db down")

	if _, err := svc.Settlement(context.Background(), "team_001"); err == nil {
		t.Fatal("expected settlement to fail when notify task cannot be created")
	}

	// 库存与完成数都应被补偿回滚。
	if stock.deducted != 1 || stock.restored != 1 {
		t.Fatalf("expected stock deducted and compensated once, got deducted=%d restored=%d",
			stock.deducted, stock.restored)
	}
	stored, _ := teamRepo.GetTeam(context.Background(), "team_001")
	if stored.CompleteCount != 0 {
		t.Fatalf("expected complete_count rolled back to 0, got %d", stored.CompleteCount)
	}
	if stored.TeamState != common.TeamStateBuilding {
		t.Fatalf("expected team state rolled back to BUILDING for retry, got %d", stored.TeamState)
	}

	// 故障恢复后重试，应能正常成团并补发通知。
	notifyRepo.createErr = nil
	if _, err := svc.Settlement(context.Background(), "team_001"); err != nil {
		t.Fatalf("retry after recovery failed: %v", err)
	}
	stored, _ = teamRepo.GetTeam(context.Background(), "team_001")
	if stored.TeamState != common.TeamStateSuccess {
		t.Fatalf("expected SUCCESS after retry, got %d", stored.TeamState)
	}
	if len(notifyRepo.tasks) != 1 {
		t.Fatalf("expected 1 notify task after retry, got %d", len(notifyRepo.tasks))
	}
}
