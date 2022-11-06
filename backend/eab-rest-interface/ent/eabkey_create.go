// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/hm-edu/eab-rest-interface/ent/eabkey"
)

// EABKeyCreate is the builder for creating a EABKey entity.
type EABKeyCreate struct {
	config
	mutation *EABKeyMutation
	hooks    []Hook
}

// SetUser sets the "user" field.
func (ekc *EABKeyCreate) SetUser(s string) *EABKeyCreate {
	ekc.mutation.SetUser(s)
	return ekc
}

// SetEabKey sets the "eabKey" field.
func (ekc *EABKeyCreate) SetEabKey(s string) *EABKeyCreate {
	ekc.mutation.SetEabKey(s)
	return ekc
}

// SetComment sets the "comment" field.
func (ekc *EABKeyCreate) SetComment(s string) *EABKeyCreate {
	ekc.mutation.SetComment(s)
	return ekc
}

// SetNillableComment sets the "comment" field if the given value is not nil.
func (ekc *EABKeyCreate) SetNillableComment(s *string) *EABKeyCreate {
	if s != nil {
		ekc.SetComment(*s)
	}
	return ekc
}

// Mutation returns the EABKeyMutation object of the builder.
func (ekc *EABKeyCreate) Mutation() *EABKeyMutation {
	return ekc.mutation
}

// Save creates the EABKey in the database.
func (ekc *EABKeyCreate) Save(ctx context.Context) (*EABKey, error) {
	var (
		err  error
		node *EABKey
	)
	if len(ekc.hooks) == 0 {
		if err = ekc.check(); err != nil {
			return nil, err
		}
		node, err = ekc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*EABKeyMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = ekc.check(); err != nil {
				return nil, err
			}
			ekc.mutation = mutation
			if node, err = ekc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(ekc.hooks) - 1; i >= 0; i-- {
			if ekc.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = ekc.hooks[i](mut)
		}
		v, err := mut.Mutate(ctx, ekc.mutation)
		if err != nil {
			return nil, err
		}
		nv, ok := v.(*EABKey)
		if !ok {
			return nil, fmt.Errorf("unexpected node type %T returned from EABKeyMutation", v)
		}
		node = nv
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (ekc *EABKeyCreate) SaveX(ctx context.Context) *EABKey {
	v, err := ekc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ekc *EABKeyCreate) Exec(ctx context.Context) error {
	_, err := ekc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ekc *EABKeyCreate) ExecX(ctx context.Context) {
	if err := ekc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ekc *EABKeyCreate) check() error {
	if _, ok := ekc.mutation.User(); !ok {
		return &ValidationError{Name: "user", err: errors.New(`ent: missing required field "EABKey.user"`)}
	}
	if _, ok := ekc.mutation.EabKey(); !ok {
		return &ValidationError{Name: "eabKey", err: errors.New(`ent: missing required field "EABKey.eabKey"`)}
	}
	return nil
}

func (ekc *EABKeyCreate) sqlSave(ctx context.Context) (*EABKey, error) {
	_node, _spec := ekc.createSpec()
	if err := sqlgraph.CreateNode(ctx, ekc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	return _node, nil
}

func (ekc *EABKeyCreate) createSpec() (*EABKey, *sqlgraph.CreateSpec) {
	var (
		_node = &EABKey{config: ekc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: eabkey.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: eabkey.FieldID,
			},
		}
	)
	if value, ok := ekc.mutation.User(); ok {
		_spec.SetField(eabkey.FieldUser, field.TypeString, value)
		_node.User = value
	}
	if value, ok := ekc.mutation.EabKey(); ok {
		_spec.SetField(eabkey.FieldEabKey, field.TypeString, value)
		_node.EabKey = value
	}
	if value, ok := ekc.mutation.Comment(); ok {
		_spec.SetField(eabkey.FieldComment, field.TypeString, value)
		_node.Comment = value
	}
	return _node, _spec
}

// EABKeyCreateBulk is the builder for creating many EABKey entities in bulk.
type EABKeyCreateBulk struct {
	config
	builders []*EABKeyCreate
}

// Save creates the EABKey entities in the database.
func (ekcb *EABKeyCreateBulk) Save(ctx context.Context) ([]*EABKey, error) {
	specs := make([]*sqlgraph.CreateSpec, len(ekcb.builders))
	nodes := make([]*EABKey, len(ekcb.builders))
	mutators := make([]Mutator, len(ekcb.builders))
	for i := range ekcb.builders {
		func(i int, root context.Context) {
			builder := ekcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*EABKeyMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ekcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ekcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ekcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ekcb *EABKeyCreateBulk) SaveX(ctx context.Context) []*EABKey {
	v, err := ekcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ekcb *EABKeyCreateBulk) Exec(ctx context.Context) error {
	_, err := ekcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ekcb *EABKeyCreateBulk) ExecX(ctx context.Context) {
	if err := ekcb.Exec(ctx); err != nil {
		panic(err)
	}
}
