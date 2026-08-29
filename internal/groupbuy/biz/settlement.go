package biz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/saga"
)

type SettlementService struct {
	teamRepo    TeamRepo
	orderRepo   OrderRepo
	notifySvc   *NotifyService
	stockClient StockClient
	logger      *slog.Logger
}

type StockClient interface {
	DeductStock(ctx context.Context, stockKey string, count int32) error
	RestoreStock(ctx context.Context, stockKey string, count int32) error
}

func NewSettlementService(teamRepo TeamRepo, orderRepo OrderRepo, notifySvc *NotifyService, stockClient StockClient) *SettlementService {
	return &SettlementService{
		teamRepo:    teamRepo,
		orderRepo:   orderRepo,
		notifySvc:   notifySvc,
		stockClient: stockClient,
		logger:      slog.Default(),
	}
}

// SetLogger 允许外部注入 logger（保持构造签名不变，避免影响 wire 装配）。
func (s *SettlementService) SetLogger(logger *slog.Logger) {
	if logger != nil {
		s.logger = logger
	}
}

// teamCompletedKey 是 Saga 步骤之间传递"本次是否成团"的键。
const teamCompletedKey = "team_completed"

// Settlement 结算一次拼团参与。
//
// 该流程跨越"远程库存服务"与"本地数据库"，无法用单个本地事务覆盖，
// 因此用 Saga 编排：顺序执行各步骤，任一步失败则逆序补偿已完成的步骤。
//
// 步骤：
//  1. deduct_stock            跨服务扣减库存（补偿：归还库存）
//  2. complete_team           本地原子自增完成数，达标则流转成团（补偿：回滚完成数与状态）
//  3. create_team_success_notify 成团则落一条通知任务（本地消息表，由后台任务保证投递）
//
// 幂等性由第 2 步的条件更新保证：只有真正把状态从"进行中"改为"成团"的
// 那一次调用会触发通知，重复结算不会重复扣库存或重复发奖。
func (s *SettlementService) Settlement(ctx context.Context, teamID string) (*GroupBuyTeam, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyTeamNotExist.Code, err)
	}
	// 幂等：已成团直接返回，避免重复扣库存与重复通知。
	if team.TeamState == common.TeamStateSuccess {
		return team, nil
	}
	if team.TeamState != common.TeamStateBuilding {
		return nil, fmt.Errorf("%s: %w", common.GroupBuySettlementFail.Code, ErrTeamNotSettleable)
	}

	stockKey := fmt.Sprintf("team:%s", teamID)

	c := saga.New(
		// 1) 扣库存：这是唯一无法纳入本地事务的一步，只能靠补偿回滚。
		saga.Step{
			Name: "deduct_stock",
			Action: func(ctx context.Context) error {
				return s.stockClient.DeductStock(ctx, stockKey, 1)
			},
			Compensate: func(ctx context.Context) error {
				return s.stockClient.RestoreStock(ctx, stockKey, 1)
			},
		},
		// 2) 自增完成数并在达标时流转成团状态（原子且幂等）。
		saga.Step{
			Name: "complete_team",
			Action: func(ctx context.Context) error {
				count, completed, err := s.teamRepo.CompleteTeam(ctx, teamID, team.TargetCount, common.TeamStateSuccess)
				if err != nil {
					return err
				}
				team.CompleteCount = count
				if completed {
					team.TeamState = common.TeamStateSuccess
				}
				if bag := saga.BagFrom(ctx); bag != nil {
					bag.Set(teamCompletedKey, completed)
				}
				return nil
			},
			Compensate: func(ctx context.Context) error {
				return s.teamRepo.RollbackTeamComplete(ctx, teamID, common.TeamStateBuilding)
			},
		},
		// 3) 成团则创建通知任务（本地消息表，保证最终一致）。
		saga.Step{
			Name: "create_team_success_notify",
			Action: func(ctx context.Context) error {
				completed := false
				if bag := saga.BagFrom(ctx); bag != nil {
					completed, _ = bag.GetBool(teamCompletedKey)
				}
				if !completed {
					return nil // 尚未成团，无需通知
				}
				return s.notifySvc.CreateTeamSuccessNotify(ctx, teamID, map[string]interface{}{
					"team_id":        teamID,
					"activity_id":    team.ActivityID,
					"complete_count": team.CompleteCount,
				})
			},
			// 无需反向补偿：第 2 步的补偿会把团队退回"进行中"，
			// 后续重试可以重新成团并补建这条通知。
		},
	).WithLogger(s.logger).WithLog(saga.NewSlogLog(s.logger))

	if err := c.Run(ctx); err != nil {
		// 补偿失败意味着系统可能停留在不一致状态，必须显式告警以便人工介入。
		var sErr *saga.SagaError
		if errors.As(err, &sErr) && len(sErr.CompErrors) > 0 {
			s.logger.Error("SETTLEMENT SAGA COMPENSATION FAILED, manual intervention required",
				slog.String("team_id", teamID),
				slog.String("saga_error", err.Error()),
			)
		}
		return nil, fmt.Errorf("%s: %w", common.GroupBuySettlementFail.Code, err)
	}
	return team, nil
}
