package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type StrategyAward struct{ ent.Schema }

func (StrategyAward) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
func (StrategyAward) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("strategy_id").MaxLen(32),
		field.String("award_id").MaxLen(32),
		field.String("award_name").MaxLen(64),
		field.Float("award_rate"),
		field.Int32("award_count").Default(0),
		field.String("rule_models").MaxLen(256).Optional().Nillable(),
	}
}
func (StrategyAward) Edges() []ent.Edge { return nil }
