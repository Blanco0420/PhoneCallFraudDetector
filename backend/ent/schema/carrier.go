package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Carrier holds the schema definition for the Carrier entity.
type Carrier struct {
	ent.Schema
}

// Fields of the Carrier.
func (Carrier) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").Optional().Unique(),
	}
}

// Edges of the Carrier.
func (Carrier) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("number", Number.Type),
	}
}
