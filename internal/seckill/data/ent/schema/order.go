package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type SeckillOrder struct {
	ent.Schema
}

func (SeckillOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (SeckillOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique().
			StorageKey("id"),
		field.String("order_id").
			Unique().
			MaxLen(32),
		field.String("activity_id").
			MaxLen(32),
		field.Int64("user_id"),
		field.String("sku_id").
			MaxLen(32),
		field.Int32("order_state").
			Default(0),
		field.Time("order_time"),
		field.Time("pay_time").
			Optional().
			Nillable(),
	}
}

func (SeckillOrder) Edges() []ent.Edge {
	return nil
}
