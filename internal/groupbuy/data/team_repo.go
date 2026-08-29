package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuyteam"
	"github.com/marketing-platform/pkg/common"
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

// CompleteTeam 原子地完成"完成数 +1"，并在达标时把团队流转为成团状态。
//
// 实现要点（保证幂等与安全）：
//  1. 自增带条件：仅当团队仍处于"进行中"且完成数尚未达到目标时才自增，
//     避免并发下完成数超过目标人数；
//  2. 状态流转用条件更新（WHERE team_state = buildingState），
//     只有真正从"进行中"改为"成团"才算本次成团，从而保证通知只创建一次。
func (r *teamRepo) CompleteTeam(ctx context.Context, teamID string, targetCount, successState int32) (int32, bool, error) {
	building := common.TeamStateBuilding

	n, err := r.data.db.GroupBuyTeam.Update().
		Where(
			groupbuyteam.TeamIDEQ(teamID),
			groupbuyteam.TeamStateEQ(building),
			groupbuyteam.CompleteCountLT(targetCount),
		).
		AddCompleteCount(1).
		Save(ctx)
	if err != nil {
		return 0, false, err
	}

	// 未发生自增：团队不存在、已结束或人数已满。
	// 明确报错，让上层 Saga 知道本次没有实际推进，从而触发补偿（回滚库存）。
	if n == 0 {
		return 0, false, biz.ErrTeamNotSettleable
	}

	po, err := r.data.db.GroupBuyTeam.Query().
		Where(groupbuyteam.TeamIDEQ(teamID)).
		Only(ctx)
	if err != nil {
		return 0, false, err
	}

	// 达标则尝试流转状态；条件更新保证只有一个并发请求能成功成团。
	completed := false
	if po.CompleteCount >= targetCount {
		k, err := r.data.db.GroupBuyTeam.Update().
			Where(
				groupbuyteam.TeamIDEQ(teamID),
				groupbuyteam.TeamStateEQ(building),
			).
			SetTeamState(successState).
			Save(ctx)
		if err != nil {
			return po.CompleteCount, false, err
		}
		completed = k == 1
	}
	return po.CompleteCount, completed, nil
}

// RollbackTeamComplete 回滚完成数与团队状态（Saga 补偿）。
// 完成数下限为 0；状态退回"进行中"，使后续重试能重新成团并补发通知。
func (r *teamRepo) RollbackTeamComplete(ctx context.Context, teamID string, buildingState int32) error {
	_, err := r.data.db.GroupBuyTeam.Update().
		Where(
			groupbuyteam.TeamIDEQ(teamID),
			groupbuyteam.CompleteCountGT(0),
		).
		AddCompleteCount(-1).
		SetTeamState(buildingState).
		Save(ctx)
	return err
}
