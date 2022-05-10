package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EABKey holds the schema definition for the EABKey entity.
type EABKey struct {
	ent.Schema
}

// Fields of the EABKey.
func (EABKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("user"),
		field.String("eabKey").Unique(),
	}
}

// Indexes defines the indices for the EABKey entity.
func (EABKey) Indexes() []ent.Index {
	return []ent.Index{
		// unique index.
		index.Fields("user", "eabKey").
			Unique(),
	}
}
