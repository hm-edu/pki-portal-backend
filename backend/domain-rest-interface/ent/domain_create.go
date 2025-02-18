// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/hm-edu/domain-rest-interface/ent/delegation"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
)

// DomainCreate is the builder for creating a Domain entity.
type DomainCreate struct {
	config
	mutation *DomainMutation
	hooks    []Hook
}

// SetCreateTime sets the "create_time" field.
func (dc *DomainCreate) SetCreateTime(t time.Time) *DomainCreate {
	dc.mutation.SetCreateTime(t)
	return dc
}

// SetNillableCreateTime sets the "create_time" field if the given value is not nil.
func (dc *DomainCreate) SetNillableCreateTime(t *time.Time) *DomainCreate {
	if t != nil {
		dc.SetCreateTime(*t)
	}
	return dc
}

// SetUpdateTime sets the "update_time" field.
func (dc *DomainCreate) SetUpdateTime(t time.Time) *DomainCreate {
	dc.mutation.SetUpdateTime(t)
	return dc
}

// SetNillableUpdateTime sets the "update_time" field if the given value is not nil.
func (dc *DomainCreate) SetNillableUpdateTime(t *time.Time) *DomainCreate {
	if t != nil {
		dc.SetUpdateTime(*t)
	}
	return dc
}

// SetFqdn sets the "fqdn" field.
func (dc *DomainCreate) SetFqdn(s string) *DomainCreate {
	dc.mutation.SetFqdn(s)
	return dc
}

// SetOwner sets the "owner" field.
func (dc *DomainCreate) SetOwner(s string) *DomainCreate {
	dc.mutation.SetOwner(s)
	return dc
}

// SetApproved sets the "approved" field.
func (dc *DomainCreate) SetApproved(b bool) *DomainCreate {
	dc.mutation.SetApproved(b)
	return dc
}

// SetNillableApproved sets the "approved" field if the given value is not nil.
func (dc *DomainCreate) SetNillableApproved(b *bool) *DomainCreate {
	if b != nil {
		dc.SetApproved(*b)
	}
	return dc
}

// AddDelegationIDs adds the "delegations" edge to the Delegation entity by IDs.
func (dc *DomainCreate) AddDelegationIDs(ids ...int) *DomainCreate {
	dc.mutation.AddDelegationIDs(ids...)
	return dc
}

// AddDelegations adds the "delegations" edges to the Delegation entity.
func (dc *DomainCreate) AddDelegations(d ...*Delegation) *DomainCreate {
	ids := make([]int, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return dc.AddDelegationIDs(ids...)
}

// Mutation returns the DomainMutation object of the builder.
func (dc *DomainCreate) Mutation() *DomainMutation {
	return dc.mutation
}

// Save creates the Domain in the database.
func (dc *DomainCreate) Save(ctx context.Context) (*Domain, error) {
	dc.defaults()
	return withHooks(ctx, dc.sqlSave, dc.mutation, dc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (dc *DomainCreate) SaveX(ctx context.Context) *Domain {
	v, err := dc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dc *DomainCreate) Exec(ctx context.Context) error {
	_, err := dc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dc *DomainCreate) ExecX(ctx context.Context) {
	if err := dc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (dc *DomainCreate) defaults() {
	if _, ok := dc.mutation.CreateTime(); !ok {
		v := domain.DefaultCreateTime()
		dc.mutation.SetCreateTime(v)
	}
	if _, ok := dc.mutation.UpdateTime(); !ok {
		v := domain.DefaultUpdateTime()
		dc.mutation.SetUpdateTime(v)
	}
	if _, ok := dc.mutation.Approved(); !ok {
		v := domain.DefaultApproved
		dc.mutation.SetApproved(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dc *DomainCreate) check() error {
	if _, ok := dc.mutation.CreateTime(); !ok {
		return &ValidationError{Name: "create_time", err: errors.New(`ent: missing required field "Domain.create_time"`)}
	}
	if _, ok := dc.mutation.UpdateTime(); !ok {
		return &ValidationError{Name: "update_time", err: errors.New(`ent: missing required field "Domain.update_time"`)}
	}
	if _, ok := dc.mutation.Fqdn(); !ok {
		return &ValidationError{Name: "fqdn", err: errors.New(`ent: missing required field "Domain.fqdn"`)}
	}
	if v, ok := dc.mutation.Fqdn(); ok {
		if err := domain.FqdnValidator(v); err != nil {
			return &ValidationError{Name: "fqdn", err: fmt.Errorf(`ent: validator failed for field "Domain.fqdn": %w`, err)}
		}
	}
	if _, ok := dc.mutation.Owner(); !ok {
		return &ValidationError{Name: "owner", err: errors.New(`ent: missing required field "Domain.owner"`)}
	}
	if v, ok := dc.mutation.Owner(); ok {
		if err := domain.OwnerValidator(v); err != nil {
			return &ValidationError{Name: "owner", err: fmt.Errorf(`ent: validator failed for field "Domain.owner": %w`, err)}
		}
	}
	if _, ok := dc.mutation.Approved(); !ok {
		return &ValidationError{Name: "approved", err: errors.New(`ent: missing required field "Domain.approved"`)}
	}
	return nil
}

func (dc *DomainCreate) sqlSave(ctx context.Context) (*Domain, error) {
	if err := dc.check(); err != nil {
		return nil, err
	}
	_node, _spec := dc.createSpec()
	if err := sqlgraph.CreateNode(ctx, dc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	dc.mutation.id = &_node.ID
	dc.mutation.done = true
	return _node, nil
}

func (dc *DomainCreate) createSpec() (*Domain, *sqlgraph.CreateSpec) {
	var (
		_node = &Domain{config: dc.config}
		_spec = sqlgraph.NewCreateSpec(domain.Table, sqlgraph.NewFieldSpec(domain.FieldID, field.TypeInt))
	)
	if value, ok := dc.mutation.CreateTime(); ok {
		_spec.SetField(domain.FieldCreateTime, field.TypeTime, value)
		_node.CreateTime = value
	}
	if value, ok := dc.mutation.UpdateTime(); ok {
		_spec.SetField(domain.FieldUpdateTime, field.TypeTime, value)
		_node.UpdateTime = value
	}
	if value, ok := dc.mutation.Fqdn(); ok {
		_spec.SetField(domain.FieldFqdn, field.TypeString, value)
		_node.Fqdn = value
	}
	if value, ok := dc.mutation.Owner(); ok {
		_spec.SetField(domain.FieldOwner, field.TypeString, value)
		_node.Owner = value
	}
	if value, ok := dc.mutation.Approved(); ok {
		_spec.SetField(domain.FieldApproved, field.TypeBool, value)
		_node.Approved = value
	}
	if nodes := dc.mutation.DelegationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   domain.DelegationsTable,
			Columns: []string{domain.DelegationsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(delegation.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// DomainCreateBulk is the builder for creating many Domain entities in bulk.
type DomainCreateBulk struct {
	config
	err      error
	builders []*DomainCreate
}

// Save creates the Domain entities in the database.
func (dcb *DomainCreateBulk) Save(ctx context.Context) ([]*Domain, error) {
	if dcb.err != nil {
		return nil, dcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(dcb.builders))
	nodes := make([]*Domain, len(dcb.builders))
	mutators := make([]Mutator, len(dcb.builders))
	for i := range dcb.builders {
		func(i int, root context.Context) {
			builder := dcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*DomainMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, dcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, dcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, dcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (dcb *DomainCreateBulk) SaveX(ctx context.Context) []*Domain {
	v, err := dcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dcb *DomainCreateBulk) Exec(ctx context.Context) error {
	_, err := dcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcb *DomainCreateBulk) ExecX(ctx context.Context) {
	if err := dcb.Exec(ctx); err != nil {
		panic(err)
	}
}
