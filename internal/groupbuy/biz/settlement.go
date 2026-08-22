package biz

import (
	"context"
	"fmt"

	"github.com/marketing-platform/pkg/common"
)

type SettlementService struct {
	teamRepo    TeamRepo
	orderRepo   OrderRepo
	notifySvc   *NotifyService
	stockClient StockClient
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
	}
}

func (s *SettlementService) Settlement(ctx context.Context, teamID string) (*GroupBuyTeam, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyTeamNotExist.Code, err)
	}

	stockKey := fmt.Sprintf("team:%s", teamID)
	if err := s.stockClient.DeductStock(ctx, stockKey, 1); err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuySettlementFail.Code, err)
	}

	completeCount, err := s.teamRepo.IncrementTeamComplete(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuySettlementFail.Code, err)
	}

	team.CompleteCount = completeCount

	if completeCount >= team.TargetCount {
		team.TeamState = 1

		if err := s.notifySvc.CreateTeamSuccessNotify(ctx, teamID, map[string]interface{}{
			"team_id":        teamID,
			"activity_id":    team.ActivityID,
			"complete_count": completeCount,
		}); err != nil {
			return nil, fmt.Errorf("create notify task failed: %w", err)
		}
	}

	return team, nil
}
