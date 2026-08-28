package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type GroupBuyActivity struct {
	ent.Schema
}

func (GroupBuyActivity) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (GroupBuyActivity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("activity_id").Unique().MaxLen(32),
		field.String("activity_name").MaxLen(64),
		field.String("discount_id").MaxLen(32),
		field.Int32("group_type").Default(0),
		field.Int32("target_count").Default(2),
		field.Int32("valid_time").Default(24),
		field.String("tag_id").MaxLen(32).Optional().Nillable(),
		field.Int32("activity_state").Default(0),
	}
}

func (GroupBuyActivity) Edges() []ent.Edge { return nil }
