package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type GroupBuyTeam struct {
	ent.Schema
}

func (GroupBuyTeam) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (GroupBuyTeam) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("team_id").Unique().MaxLen(32),
		field.String("activity_id").MaxLen(32),
		field.Int32("target_count").Default(2),
		field.Int32("complete_count").Default(0),
		field.Int32("lock_count").Default(0),
		field.Int32("team_state").Default(0),
	}
}

func (GroupBuyTeam) Edges() []ent.Edge { return nil }
