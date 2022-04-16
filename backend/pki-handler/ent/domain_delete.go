// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/hm-edu/pki-handler/ent/domain"
	"github.com/hm-edu/pki-handler/ent/predicate"
)

// DomainDelete is the builder for deleting a Domain entity.
type DomainDelete struct {
	config
	hooks    []Hook
	mutation *DomainMutation
}

// Where appends a list predicates to the DomainDelete builder.
func (dd *DomainDelete) Where(ps ...predicate.Domain) *DomainDelete {
	dd.mutation.Where(ps...)
	return dd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (dd *DomainDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(dd.hooks) == 0 {
		affected, err = dd.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DomainMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			dd.mutation = mutation
			affected, err = dd.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(dd.hooks) - 1; i >= 0; i-- {
			if dd.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = dd.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, dd.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (dd *DomainDelete) ExecX(ctx context.Context) int {
	n, err := dd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (dd *DomainDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: domain.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: domain.FieldID,
			},
		},
	}
	if ps := dd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, dd.driver, _spec)
}

// DomainDeleteOne is the builder for deleting a single Domain entity.
type DomainDeleteOne struct {
	dd *DomainDelete
}

// Exec executes the deletion query.
func (ddo *DomainDeleteOne) Exec(ctx context.Context) error {
	n, err := ddo.dd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{domain.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ddo *DomainDeleteOne) ExecX(ctx context.Context) {
	ddo.dd.ExecX(ctx)
}
