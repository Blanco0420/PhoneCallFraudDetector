package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Caller holds the schema definition for the Caller entity.
type Caller struct {
	ent.Schema
}

// Fields of the Callers.
func (Caller) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("IsFraud"),
		field.Int("FraudScore"),
	}
}

// Edges of the Callers.
func (Caller) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("number", Number.Type),
	}
}
