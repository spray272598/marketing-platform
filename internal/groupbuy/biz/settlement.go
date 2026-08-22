package biz

import (
	"context"
	"fmt"

	"github.com/marketing-platform/pkg/common"
)

type SettlementService struct {
	teamRepo  TeamRepo
	orderRepo OrderRepo
	mqRepo    MQRepo
}

func NewSettlementService(teamRepo TeamRepo, orderRepo OrderRepo, mqRepo MQRepo) *SettlementService {
	return &SettlementService{
		teamRepo:  teamRepo,
		orderRepo: orderRepo,
		mqRepo:    mqRepo,
	}
}

func (s *SettlementService) Settlement(ctx context.Context, teamID string) (*GroupBuyTeam, error) {
	team, err := s.teamRepo.GetTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf(common.GroupBuyTeamExpired.Code+": %w", err)
	}

	// 增加完成数
	completeCount, err := s.teamRepo.IncrementTeamComplete(ctx, teamID)
	if err != nil {
		return nil, err
	}

	team.CompleteCount = completeCount

	// 检查是否达成目标
	if completeCount >= team.TargetCount {
		team.TeamState = common.TeamStateSuccess
		// 异步通知成团成功
		_ = s.mqRepo.PublishTeamSuccessMessage(ctx, teamID)
	}

	return team, nil
}
