package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Domain holds the schema definition for the Domain entity.
type Domain struct {
	ent.Schema
}

// Fields of the Domain.
func (Domain) Fields() []ent.Field {
	return []ent.Field{
		field.String("fqdn").NotEmpty().Unique(),
		field.String("owner").NotEmpty(),
		field.Bool("approved").Default(false),
	}
}

// Edges of the Domain.
func (Domain) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("delegations", Delegation.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Domain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
