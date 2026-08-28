package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type NotifyTask struct {
	ent.Schema
}

func (NotifyTask) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

func (NotifyTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().Unique().StorageKey("id"),
		field.String("task_id").Unique().MaxLen(32),
		field.String("notify_type").MaxLen(8),
		field.Int32("notify_status").Default(0),
		field.String("notify_url").MaxLen(256).Optional().Nillable(),
		field.Text("notify_data").Optional().Nillable(),
		field.String("uuid").Unique().MaxLen(64),
		field.Int32("retry_count").Default(0),
		field.Int32("max_retry").Default(3),
		field.Int64("next_time").Default(0),
	}
}

func (NotifyTask) Edges() []ent.Edge { return nil }
