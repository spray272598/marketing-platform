package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type LotteryActivity struct{ ent.Schema }

func (LotteryActivity) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
func (LotteryActivity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("activity_id").Unique().MaxLen(32),
		field.String("activity_name").MaxLen(64),
		field.String("strategy_id").MaxLen(32),
		field.Int32("activity_state").Default(0),
	}
}
func (LotteryActivity) Edges() []ent.Edge { return nil }
