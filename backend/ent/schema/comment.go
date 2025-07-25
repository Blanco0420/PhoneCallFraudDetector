package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Comments holds the schema definition for the Comments entity.
type Comment struct {
	ent.Schema
}

// Fields of the Comments.
func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Time("PostDate"),
		field.String("CommentText"),
		field.Int("CommentFraudScore").
			Default(0),
	}
}

// Edges of the Comments.
func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", Provider.Type).
			Ref("comment").
			Unique().
			Required(),
	}
}
