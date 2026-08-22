package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
)

type teamRepo struct {
	db *Data
}

func NewTeamRepo(data *Data) biz.TeamRepo {
	return &teamRepo{db: data}
}

func (r *teamRepo) CreateTeam(ctx context.Context, team *biz.GroupBuyTeam) error {
	query := `INSERT INTO groupbuy_team (team_id, activity_id, target_count, complete_count, lock_count, team_state) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.db.ExecContext(ctx, query, team.TeamID, team.ActivityID, team.TargetCount, team.CompleteCount, team.LockCount, team.TeamState)
	return err
}

func (r *teamRepo) GetTeam(ctx context.Context, teamID string) (*biz.GroupBuyTeam, error) {
	query := `SELECT id, team_id, activity_id, target_count, complete_count, lock_count, team_state FROM groupbuy_team WHERE team_id = ?`
	team := &biz.GroupBuyTeam{}
	err := r.db.db.QueryRowContext(ctx, query, teamID).Scan(&team.ID, &team.TeamID, &team.ActivityID, &team.TargetCount, &team.CompleteCount, &team.LockCount, &team.TeamState)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}
	return team, nil
}

func (r *teamRepo) IncrementTeamComplete(ctx context.Context, teamID string) (int32, error) {
	query := `UPDATE groupbuy_team SET complete_count = complete_count + 1, update_time = NOW() WHERE team_id = ?`
	_, err := r.db.db.ExecContext(ctx, query, teamID)
	if err != nil {
		return 0, err
	}

	var count int32
	err = r.db.db.QueryRowContext(ctx, `SELECT complete_count FROM groupbuy_team WHERE team_id = ?`, teamID).Scan(&count)
	return count, err
}
