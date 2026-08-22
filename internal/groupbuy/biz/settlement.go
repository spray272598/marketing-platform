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
}

func NewSettlementService(teamRepo TeamRepo, orderRepo OrderRepo, notifySvc *NotifyService) *SettlementService {
	return &SettlementService{
		teamRepo:  teamRepo,
		orderRepo: orderRepo,
		notifySvc: notifySvc,
	}
}

func (s *SettlementService) Settlement(ctx context.Context, teamID string) (*GroupBuyTeam, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyTeamNotExist.Code, err)
	}

	completeCount, err := s.teamRepo.IncrementTeamComplete(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuySettlementFail.Code, err)
	}

	team.CompleteCount = completeCount

	if completeCount >= team.TargetCount {
		team.TeamState = 1

		if err := s.notifySvc.CreateTeamSuccessNotify(ctx, teamID, map[string]interface{}{
			"team_id":       teamID,
			"activity_id":   team.ActivityID,
			"complete_count": completeCount,
		}); err != nil {
			return nil, fmt.Errorf("create notify task failed: %w", err)
		}
	}

	return team, nil
}
