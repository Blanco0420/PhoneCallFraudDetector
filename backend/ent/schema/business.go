package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Business holds the schema definition for the Business entity.
type Business struct {
	ent.Schema
}

// Fields of the Businesses.
func (Business) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").Optional(),
		field.String("WebsiteLink").Optional(),
		field.String("Overview").Optional().
			Comment("Overview of what the business does"),
	}
}

// Edges of the Businesses.
func (Business) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", Provider.Type).
			Ref("business").
			Unique().
			Required(),
		edge.To("address", Address.Type),
	}
}
