package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type LotteryStrategy struct{ ent.Schema }

func (LotteryStrategy) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
func (LotteryStrategy) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("strategy_id").Unique().MaxLen(32),
		field.String("rule_models").MaxLen(256).Optional().Nillable(),
	}
}
func (LotteryStrategy) Edges() []ent.Edge { return nil }
