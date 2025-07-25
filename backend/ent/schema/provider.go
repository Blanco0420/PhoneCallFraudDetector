package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Providers holds the schema definition for the Providers entity.
type Provider struct {
	ent.Schema
}

// Fields of the Providers.
func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.Text("Name").Unique(),
	}
}

// Edges of the Providers.
func (Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("number", Number.Type).
			Ref("provider").
			Unique().
			Required(),
		edge.To("comment", Comment.Type),
		edge.To("business", Business.Type).
			Unique(),
	}
}
