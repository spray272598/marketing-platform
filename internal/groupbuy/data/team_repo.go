package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuyteam"
)

type teamRepo struct {
	data *Data
}

func NewTeamRepo(data *Data) biz.TeamRepo {
	return &teamRepo{data: data}
}

func (r *teamRepo) CreateTeam(ctx context.Context, team *biz.GroupBuyTeam) error {
	_, err := r.data.db.GroupBuyTeam.Create().
		SetTeamID(team.TeamID).
		SetActivityID(team.ActivityID).
		SetTargetCount(team.TargetCount).
		SetCompleteCount(team.CompleteCount).
		SetLockCount(team.LockCount).
		SetTeamState(team.TeamState).
		Save(ctx)
	return err
}

func (r *teamRepo) GetTeam(ctx context.Context, teamID string) (*biz.GroupBuyTeam, error) {
	po, err := r.data.db.GroupBuyTeam.Query().
		Where(groupbuyteam.TeamIDEQ(teamID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("team not found: %s", teamID)
		}
		return nil, err
	}
	return &biz.GroupBuyTeam{
		TeamID:        po.TeamID,
		ActivityID:    po.ActivityID,
		TargetCount:   po.TargetCount,
		CompleteCount: po.CompleteCount,
		LockCount:     po.LockCount,
		TeamState:     po.TeamState,
	}, nil
}

func (r *teamRepo) IncrementTeamComplete(ctx context.Context, teamID string) (int32, error) {
	_, err := r.data.db.GroupBuyTeam.Update().
		Where(groupbuyteam.TeamIDEQ(teamID)).
		AddCompleteCount(1).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	po, err := r.data.db.GroupBuyTeam.Query().
		Where(groupbuyteam.TeamIDEQ(teamID)).
		Only(ctx)
	if err != nil {
		return 0, err
	}
	return po.CompleteCount, nil
}
