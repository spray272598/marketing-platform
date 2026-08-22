package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/lottery/biz"
)

type strategyRepo struct {
	db *Data
}

func NewStrategyRepo(data *Data) biz.StrategyRepo {
	return &strategyRepo{db: data}
}

func (r *strategyRepo) GetStrategy(ctx context.Context, strategyID string) (*biz.LotteryStrategy, error) {
	query := `SELECT id, strategy_id, rule_models FROM lottery_strategy WHERE strategy_id = ?`

	strategy := &biz.LotteryStrategy{}
	err := r.db.db.QueryRowContext(ctx, query, strategyID).Scan(
		&strategy.ID, &strategy.StrategyID, &strategy.RuleModels,
	)
	if err != nil {
		return nil, fmt.Errorf("strategy not found: %w", err)
	}
	return strategy, nil
}

func (r *strategyRepo) GetStrategyAwards(ctx context.Context, strategyID string) ([]*biz.StrategyAward, error) {
	query := `SELECT id, award_id, award_name, award_rate, award_count, rule_models 
			  FROM strategy_award WHERE strategy_id = ?`

	rows, err := r.db.db.QueryContext(ctx, query, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var awards []*biz.StrategyAward
	for rows.Next() {
		award := &biz.StrategyAward{}
		if err := rows.Scan(&award.ID, &award.AwardID, &award.AwardName,
			&award.AwardRate, &award.AwardCount, &award.RuleModels); err != nil {
			return nil, err
		}
		awards = append(awards, award)
	}
	return awards, nil
}
