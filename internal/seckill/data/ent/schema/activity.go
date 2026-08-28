package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type SeckillActivity struct {
	ent.Schema
}

func (SeckillActivity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (SeckillActivity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique().
			StorageKey("id"),
		field.String("activity_id").
			Unique().
			MaxLen(32),
		field.String("activity_name").
			MaxLen(64),
		field.String("sku_id").
			MaxLen(32),
		field.Int32("total_count").
			Default(0),
		field.Int32("limit_count").
			Default(1),
		field.Int32("activity_state").
			Default(0),
		field.Time("start_time"),
		field.Time("end_time"),
	}
}

func (SeckillActivity) Edges() []ent.Edge {
	return nil
}
