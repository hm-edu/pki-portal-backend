// Code generated by entc, DO NOT EDIT.

package domain

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/hm-edu/pki-handler/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// Fqdn applies equality check predicate on the "fqdn" field. It's identical to FqdnEQ.
func Fqdn(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldFqdn), v))
	})
}

// FqdnEQ applies the EQ predicate on the "fqdn" field.
func FqdnEQ(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldFqdn), v))
	})
}

// FqdnNEQ applies the NEQ predicate on the "fqdn" field.
func FqdnNEQ(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldFqdn), v))
	})
}

// FqdnIn applies the In predicate on the "fqdn" field.
func FqdnIn(vs ...string) predicate.Domain {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Domain(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldFqdn), v...))
	})
}

// FqdnNotIn applies the NotIn predicate on the "fqdn" field.
func FqdnNotIn(vs ...string) predicate.Domain {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Domain(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldFqdn), v...))
	})
}

// FqdnGT applies the GT predicate on the "fqdn" field.
func FqdnGT(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldFqdn), v))
	})
}

// FqdnGTE applies the GTE predicate on the "fqdn" field.
func FqdnGTE(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldFqdn), v))
	})
}

// FqdnLT applies the LT predicate on the "fqdn" field.
func FqdnLT(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldFqdn), v))
	})
}

// FqdnLTE applies the LTE predicate on the "fqdn" field.
func FqdnLTE(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldFqdn), v))
	})
}

// FqdnContains applies the Contains predicate on the "fqdn" field.
func FqdnContains(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldFqdn), v))
	})
}

// FqdnHasPrefix applies the HasPrefix predicate on the "fqdn" field.
func FqdnHasPrefix(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldFqdn), v))
	})
}

// FqdnHasSuffix applies the HasSuffix predicate on the "fqdn" field.
func FqdnHasSuffix(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldFqdn), v))
	})
}

// FqdnEqualFold applies the EqualFold predicate on the "fqdn" field.
func FqdnEqualFold(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldFqdn), v))
	})
}

// FqdnContainsFold applies the ContainsFold predicate on the "fqdn" field.
func FqdnContainsFold(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldFqdn), v))
	})
}

// HasCertificates applies the HasEdge predicate on the "certificates" edge.
func HasCertificates() predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(CertificatesTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, CertificatesTable, CertificatesPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCertificatesWith applies the HasEdge predicate on the "certificates" edge with a given conditions (other predicates).
func HasCertificatesWith(preds ...predicate.Certificate) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(CertificatesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, CertificatesTable, CertificatesPrimaryKey...),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Domain) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Domain) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Domain) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		p(s.Not())
	})
}
