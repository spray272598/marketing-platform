package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type GroupBuyOrder struct {
	ent.Schema
}

func (GroupBuyOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (GroupBuyOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("order_id").Unique().MaxLen(32),
		field.String("team_id").MaxLen(32),
		field.Int64("user_id"),
		field.String("activity_id").MaxLen(32),
		field.String("biz_id").Unique().MaxLen(64),
		field.Int32("order_state").Default(0),
	}
}

func (GroupBuyOrder) Edges() []ent.Edge { return nil }
