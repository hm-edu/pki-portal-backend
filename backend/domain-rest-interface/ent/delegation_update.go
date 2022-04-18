// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/hm-edu/domain-rest-interface/ent/delegation"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
	"github.com/hm-edu/domain-rest-interface/ent/predicate"
)

// DelegationUpdate is the builder for updating Delegation entities.
type DelegationUpdate struct {
	config
	hooks    []Hook
	mutation *DelegationMutation
}

// Where appends a list predicates to the DelegationUpdate builder.
func (du *DelegationUpdate) Where(ps ...predicate.Delegation) *DelegationUpdate {
	du.mutation.Where(ps...)
	return du
}

// SetUpdateTime sets the "update_time" field.
func (du *DelegationUpdate) SetUpdateTime(t time.Time) *DelegationUpdate {
	du.mutation.SetUpdateTime(t)
	return du
}

// SetUser sets the "user" field.
func (du *DelegationUpdate) SetUser(s string) *DelegationUpdate {
	du.mutation.SetUser(s)
	return du
}

// SetDomainID sets the "domain" edge to the Domain entity by ID.
func (du *DelegationUpdate) SetDomainID(id int) *DelegationUpdate {
	du.mutation.SetDomainID(id)
	return du
}

// SetDomain sets the "domain" edge to the Domain entity.
func (du *DelegationUpdate) SetDomain(d *Domain) *DelegationUpdate {
	return du.SetDomainID(d.ID)
}

// Mutation returns the DelegationMutation object of the builder.
func (du *DelegationUpdate) Mutation() *DelegationMutation {
	return du.mutation
}

// ClearDomain clears the "domain" edge to the Domain entity.
func (du *DelegationUpdate) ClearDomain() *DelegationUpdate {
	du.mutation.ClearDomain()
	return du
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (du *DelegationUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	du.defaults()
	if len(du.hooks) == 0 {
		if err = du.check(); err != nil {
			return 0, err
		}
		affected, err = du.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DelegationMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = du.check(); err != nil {
				return 0, err
			}
			du.mutation = mutation
			affected, err = du.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(du.hooks) - 1; i >= 0; i-- {
			if du.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = du.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, du.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (du *DelegationUpdate) SaveX(ctx context.Context) int {
	affected, err := du.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (du *DelegationUpdate) Exec(ctx context.Context) error {
	_, err := du.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (du *DelegationUpdate) ExecX(ctx context.Context) {
	if err := du.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (du *DelegationUpdate) defaults() {
	if _, ok := du.mutation.UpdateTime(); !ok {
		v := delegation.UpdateDefaultUpdateTime()
		du.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (du *DelegationUpdate) check() error {
	if v, ok := du.mutation.User(); ok {
		if err := delegation.UserValidator(v); err != nil {
			return &ValidationError{Name: "user", err: fmt.Errorf(`ent: validator failed for field "Delegation.user": %w`, err)}
		}
	}
	if _, ok := du.mutation.DomainID(); du.mutation.DomainCleared() && !ok {
		return errors.New(`ent: clearing a required unique edge "Delegation.domain"`)
	}
	return nil
}

func (du *DelegationUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   delegation.Table,
			Columns: delegation.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: delegation.FieldID,
			},
		},
	}
	if ps := du.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := du.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: delegation.FieldUpdateTime,
		})
	}
	if value, ok := du.mutation.User(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: delegation.FieldUser,
		})
	}
	if du.mutation.DomainCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   delegation.DomainTable,
			Columns: []string{delegation.DomainColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: domain.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.DomainIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   delegation.DomainTable,
			Columns: []string{delegation.DomainColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: domain.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, du.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{delegation.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// DelegationUpdateOne is the builder for updating a single Delegation entity.
type DelegationUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *DelegationMutation
}

// SetUpdateTime sets the "update_time" field.
func (duo *DelegationUpdateOne) SetUpdateTime(t time.Time) *DelegationUpdateOne {
	duo.mutation.SetUpdateTime(t)
	return duo
}

// SetUser sets the "user" field.
func (duo *DelegationUpdateOne) SetUser(s string) *DelegationUpdateOne {
	duo.mutation.SetUser(s)
	return duo
}

// SetDomainID sets the "domain" edge to the Domain entity by ID.
func (duo *DelegationUpdateOne) SetDomainID(id int) *DelegationUpdateOne {
	duo.mutation.SetDomainID(id)
	return duo
}

// SetDomain sets the "domain" edge to the Domain entity.
func (duo *DelegationUpdateOne) SetDomain(d *Domain) *DelegationUpdateOne {
	return duo.SetDomainID(d.ID)
}

// Mutation returns the DelegationMutation object of the builder.
func (duo *DelegationUpdateOne) Mutation() *DelegationMutation {
	return duo.mutation
}

// ClearDomain clears the "domain" edge to the Domain entity.
func (duo *DelegationUpdateOne) ClearDomain() *DelegationUpdateOne {
	duo.mutation.ClearDomain()
	return duo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (duo *DelegationUpdateOne) Select(field string, fields ...string) *DelegationUpdateOne {
	duo.fields = append([]string{field}, fields...)
	return duo
}

// Save executes the query and returns the updated Delegation entity.
func (duo *DelegationUpdateOne) Save(ctx context.Context) (*Delegation, error) {
	var (
		err  error
		node *Delegation
	)
	duo.defaults()
	if len(duo.hooks) == 0 {
		if err = duo.check(); err != nil {
			return nil, err
		}
		node, err = duo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DelegationMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = duo.check(); err != nil {
				return nil, err
			}
			duo.mutation = mutation
			node, err = duo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(duo.hooks) - 1; i >= 0; i-- {
			if duo.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = duo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, duo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (duo *DelegationUpdateOne) SaveX(ctx context.Context) *Delegation {
	node, err := duo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (duo *DelegationUpdateOne) Exec(ctx context.Context) error {
	_, err := duo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (duo *DelegationUpdateOne) ExecX(ctx context.Context) {
	if err := duo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (duo *DelegationUpdateOne) defaults() {
	if _, ok := duo.mutation.UpdateTime(); !ok {
		v := delegation.UpdateDefaultUpdateTime()
		duo.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (duo *DelegationUpdateOne) check() error {
	if v, ok := duo.mutation.User(); ok {
		if err := delegation.UserValidator(v); err != nil {
			return &ValidationError{Name: "user", err: fmt.Errorf(`ent: validator failed for field "Delegation.user": %w`, err)}
		}
	}
	if _, ok := duo.mutation.DomainID(); duo.mutation.DomainCleared() && !ok {
		return errors.New(`ent: clearing a required unique edge "Delegation.domain"`)
	}
	return nil
}

func (duo *DelegationUpdateOne) sqlSave(ctx context.Context) (_node *Delegation, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   delegation.Table,
			Columns: delegation.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: delegation.FieldID,
			},
		},
	}
	id, ok := duo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Delegation.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := duo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, delegation.FieldID)
		for _, f := range fields {
			if !delegation.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != delegation.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := duo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := duo.mutation.UpdateTime(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: delegation.FieldUpdateTime,
		})
	}
	if value, ok := duo.mutation.User(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: delegation.FieldUser,
		})
	}
	if duo.mutation.DomainCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   delegation.DomainTable,
			Columns: []string{delegation.DomainColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: domain.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.DomainIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   delegation.DomainTable,
			Columns: []string{delegation.DomainColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: domain.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Delegation{config: duo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, duo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{delegation.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}