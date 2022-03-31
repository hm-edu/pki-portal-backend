package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// Delegation holds the schema definition for the Delegation entity.
type Delegation struct {
	ent.Schema
}

// Fields of the Delegation.
func (Delegation) Fields() []ent.Field {
	return []ent.Field{
		field.String("user").NotEmpty(),
	}
}

// Edges of the Delegation.
func (Delegation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("domain", Domain.Type).
			Ref("delegations").
			Unique().Required(),
	}
}

// Indexes of the Delegation.
func (Delegation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user").
			Edges("domain").
			Unique(),
	}
}

// Mixin adds default time fields to this model.
func (Delegation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
