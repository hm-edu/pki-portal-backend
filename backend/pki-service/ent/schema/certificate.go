package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/hm-edu/pki-service/ent/hook"
)

// Certificate holds the schema definition for the Certificate entity.
type Certificate struct {
	ent.Schema
}

// Fields of the Certificate.
func (Certificate) Fields() []ent.Field {
	return []ent.Field{
		field.Int("sslId").Optional(),
		field.String("serial").Optional().Unique(),
		field.String("commonName").NotEmpty(),
		field.Time("notBefore").Nillable().Optional(),
		field.Time("notAfter").Optional(),
		field.String("issuedBy").Nillable().Optional(),
		field.String("source").Nillable().Optional(),
		field.Time("created").Nillable().Optional(),
		field.Enum("status").Values("Invalid", "Requested", "Approved", "Declined", "Applied", "Issued", "Revoked", "Expired", "Replaced", "Rejected", "Unmanaged", "SAApproved", "Init").Default("Invalid"),
	}
}

// Edges of the Certificate.
func (Certificate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("domains", Domain.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
	}
}

// Mixin adds default time fields to this model.
func (Certificate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Hooks of the certificate.
func (Certificate) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.Reject(ent.OpDelete | ent.OpDeleteOne),
	}
}
