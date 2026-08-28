package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type GroupBuyDiscount struct {
	ent.Schema
}

func (GroupBuyDiscount) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (GroupBuyDiscount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("discount_id").Unique().MaxLen(32),
		field.String("market_plan").MaxLen(8),
		field.String("market_expr").MaxLen(64),
	}
}

func (GroupBuyDiscount) Edges() []ent.Edge { return nil }
