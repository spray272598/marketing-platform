package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type IDMixin struct{ mixin.Schema }

func (IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
	}
}

type TimeMixin struct{ mixin.Schema }

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Optional().Nillable(),
		field.Time("updated_at").Optional().Nillable(),
	}
}

type StockItem struct{ ent.Schema }

func (StockItem) Mixin() []ent.Mixin { return []ent.Mixin{TimeMixin{}} }
func (StockItem) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("stock_key").Unique().MaxLen(64),
		field.String("stock_name").MaxLen(64),
		field.String("stock_type").MaxLen(32),
		field.Int32("stock").Default(0),
		field.Int32("total").Default(0),
	}
}
func (StockItem) Edges() []ent.Edge { return nil }
