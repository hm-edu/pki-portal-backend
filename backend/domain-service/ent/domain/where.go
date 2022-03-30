// Code generated by entc, DO NOT EDIT.

package domain

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/hm-edu/domain-service/ent/predicate"
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

// CreateTime applies equality check predicate on the "create_time" field. It's identical to CreateTimeEQ.
func CreateTime(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldCreateTime), v))
	})
}

// UpdateTime applies equality check predicate on the "update_time" field. It's identical to UpdateTimeEQ.
func UpdateTime(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUpdateTime), v))
	})
}

// Fqdn applies equality check predicate on the "fqdn" field. It's identical to FqdnEQ.
func Fqdn(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldFqdn), v))
	})
}

// Owner applies equality check predicate on the "owner" field. It's identical to OwnerEQ.
func Owner(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldOwner), v))
	})
}

// Approved applies equality check predicate on the "approved" field. It's identical to ApprovedEQ.
func Approved(v bool) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldApproved), v))
	})
}

// CreateTimeEQ applies the EQ predicate on the "create_time" field.
func CreateTimeEQ(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldCreateTime), v))
	})
}

// CreateTimeNEQ applies the NEQ predicate on the "create_time" field.
func CreateTimeNEQ(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldCreateTime), v))
	})
}

// CreateTimeIn applies the In predicate on the "create_time" field.
func CreateTimeIn(vs ...time.Time) predicate.Domain {
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
		s.Where(sql.In(s.C(FieldCreateTime), v...))
	})
}

// CreateTimeNotIn applies the NotIn predicate on the "create_time" field.
func CreateTimeNotIn(vs ...time.Time) predicate.Domain {
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
		s.Where(sql.NotIn(s.C(FieldCreateTime), v...))
	})
}

// CreateTimeGT applies the GT predicate on the "create_time" field.
func CreateTimeGT(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldCreateTime), v))
	})
}

// CreateTimeGTE applies the GTE predicate on the "create_time" field.
func CreateTimeGTE(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldCreateTime), v))
	})
}

// CreateTimeLT applies the LT predicate on the "create_time" field.
func CreateTimeLT(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldCreateTime), v))
	})
}

// CreateTimeLTE applies the LTE predicate on the "create_time" field.
func CreateTimeLTE(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldCreateTime), v))
	})
}

// UpdateTimeEQ applies the EQ predicate on the "update_time" field.
func UpdateTimeEQ(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUpdateTime), v))
	})
}

// UpdateTimeNEQ applies the NEQ predicate on the "update_time" field.
func UpdateTimeNEQ(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldUpdateTime), v))
	})
}

// UpdateTimeIn applies the In predicate on the "update_time" field.
func UpdateTimeIn(vs ...time.Time) predicate.Domain {
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
		s.Where(sql.In(s.C(FieldUpdateTime), v...))
	})
}

// UpdateTimeNotIn applies the NotIn predicate on the "update_time" field.
func UpdateTimeNotIn(vs ...time.Time) predicate.Domain {
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
		s.Where(sql.NotIn(s.C(FieldUpdateTime), v...))
	})
}

// UpdateTimeGT applies the GT predicate on the "update_time" field.
func UpdateTimeGT(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldUpdateTime), v))
	})
}

// UpdateTimeGTE applies the GTE predicate on the "update_time" field.
func UpdateTimeGTE(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldUpdateTime), v))
	})
}

// UpdateTimeLT applies the LT predicate on the "update_time" field.
func UpdateTimeLT(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldUpdateTime), v))
	})
}

// UpdateTimeLTE applies the LTE predicate on the "update_time" field.
func UpdateTimeLTE(v time.Time) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldUpdateTime), v))
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

// OwnerEQ applies the EQ predicate on the "owner" field.
func OwnerEQ(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldOwner), v))
	})
}

// OwnerNEQ applies the NEQ predicate on the "owner" field.
func OwnerNEQ(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldOwner), v))
	})
}

// OwnerIn applies the In predicate on the "owner" field.
func OwnerIn(vs ...string) predicate.Domain {
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
		s.Where(sql.In(s.C(FieldOwner), v...))
	})
}

// OwnerNotIn applies the NotIn predicate on the "owner" field.
func OwnerNotIn(vs ...string) predicate.Domain {
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
		s.Where(sql.NotIn(s.C(FieldOwner), v...))
	})
}

// OwnerGT applies the GT predicate on the "owner" field.
func OwnerGT(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldOwner), v))
	})
}

// OwnerGTE applies the GTE predicate on the "owner" field.
func OwnerGTE(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldOwner), v))
	})
}

// OwnerLT applies the LT predicate on the "owner" field.
func OwnerLT(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldOwner), v))
	})
}

// OwnerLTE applies the LTE predicate on the "owner" field.
func OwnerLTE(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldOwner), v))
	})
}

// OwnerContains applies the Contains predicate on the "owner" field.
func OwnerContains(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldOwner), v))
	})
}

// OwnerHasPrefix applies the HasPrefix predicate on the "owner" field.
func OwnerHasPrefix(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldOwner), v))
	})
}

// OwnerHasSuffix applies the HasSuffix predicate on the "owner" field.
func OwnerHasSuffix(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldOwner), v))
	})
}

// OwnerEqualFold applies the EqualFold predicate on the "owner" field.
func OwnerEqualFold(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldOwner), v))
	})
}

// OwnerContainsFold applies the ContainsFold predicate on the "owner" field.
func OwnerContainsFold(v string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldOwner), v))
	})
}

// ApprovedEQ applies the EQ predicate on the "approved" field.
func ApprovedEQ(v bool) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldApproved), v))
	})
}

// ApprovedNEQ applies the NEQ predicate on the "approved" field.
func ApprovedNEQ(v bool) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldApproved), v))
	})
}

// HasDelegations applies the HasEdge predicate on the "delegations" edge.
func HasDelegations() predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(DelegationsTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, DelegationsTable, DelegationsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasDelegationsWith applies the HasEdge predicate on the "delegations" edge with a given conditions (other predicates).
func HasDelegationsWith(preds ...predicate.Delegation) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(DelegationsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, DelegationsTable, DelegationsColumn),
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
