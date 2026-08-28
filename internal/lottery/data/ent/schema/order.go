package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type LotteryOrder struct{ ent.Schema }

func (LotteryOrder) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
func (LotteryOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("order_id").Unique().MaxLen(32),
		field.String("activity_id").MaxLen(32),
		field.Int64("user_id"),
		field.String("award_id").MaxLen(32).Optional().Nillable(),
		field.Int32("award_state").Default(0),
		field.Time("award_time").Optional().Nillable(),
	}
}
func (LotteryOrder) Edges() []ent.Edge { return nil }
