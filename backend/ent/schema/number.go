package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Number holds the schema definition for the Number entity.
type Number struct {
	ent.Schema
}

// Fields of the Numbers.
func (Number) Fields() []ent.Field {
	return []ent.Field{
		field.String("Number"),
	}
}

// Edges of the Numbers.
func (Number) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("caller", Caller.Type).
			Ref("number").
			Unique().
			Required(),
		edge.From("linetype", LineType.Type).
			Ref("number").
			Unique().
			Required(),
		edge.From("carrier", Carrier.Type).
			Ref("number").
			Unique(),
		edge.To("provider", Provider.Type),
	}
}
