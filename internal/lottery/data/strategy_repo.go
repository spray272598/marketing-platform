package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data/ent"
	"github.com/marketing-platform/internal/lottery/data/ent/lotterystrategy"
	"github.com/marketing-platform/internal/lottery/data/ent/strategyaward"
)

type strategyRepo struct{ data *Data }

func NewStrategyRepo(data *Data) biz.StrategyRepo { return &strategyRepo{data: data} }

func (r *strategyRepo) GetStrategy(ctx context.Context, strategyID string) (*biz.LotteryStrategy, error) {
	po, err := r.data.db.LotteryStrategy.Query().
		Where(lotterystrategy.StrategyIDEQ(strategyID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("strategy not found: %s", strategyID)
		}
		return nil, err
	}
	s := &biz.LotteryStrategy{StrategyID: po.StrategyID}
	if po.RuleModels != nil {
		s.RuleModels = *po.RuleModels
	}
	return s, nil
}

func (r *strategyRepo) GetStrategyAwards(ctx context.Context, strategyID string) ([]*biz.StrategyAward, error) {
	pos, err := r.data.db.StrategyAward.Query().
		Where(strategyaward.StrategyIDEQ(strategyID)).All(ctx)
	if err != nil {
		return nil, err
	}
	awards := make([]*biz.StrategyAward, 0, len(pos))
	for _, po := range pos {
		a := &biz.StrategyAward{
			AwardID:    po.AwardID,
			AwardName:  po.AwardName,
			AwardRate:  po.AwardRate,
			AwardCount: po.AwardCount,
		}
		if po.RuleModels != nil {
			a.RuleModels = *po.RuleModels
		}
		awards = append(awards, a)
	}
	return awards, nil
}
