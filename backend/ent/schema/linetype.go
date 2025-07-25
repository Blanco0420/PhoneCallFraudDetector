package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

// LineType holds the schema definition for the LineType entity.
type LineType struct {
	ent.Schema
}

// Fields of the LineType.
func (LineType) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("LineType").
			GoType(providers.LineType("")).
			Default(string(providers.LineTypeUnknown)),
	}
}

// Edges of the LineType.
func (LineType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("number", Number.Type),
	}
}
