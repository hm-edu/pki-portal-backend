// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hm-edu/domain-service/ent/delegation"
	"github.com/hm-edu/domain-service/ent/domain"
	"github.com/hm-edu/domain-service/ent/predicate"

	"entgo.io/ent"
)

const (
	// Operation types.
	OpCreate    = ent.OpCreate
	OpDelete    = ent.OpDelete
	OpDeleteOne = ent.OpDeleteOne
	OpUpdate    = ent.OpUpdate
	OpUpdateOne = ent.OpUpdateOne

	// Node types.
	TypeDelegation = "Delegation"
	TypeDomain     = "Domain"
)

// DelegationMutation represents an operation that mutates the Delegation nodes in the graph.
type DelegationMutation struct {
	config
	op            Op
	typ           string
	id            *int
	create_time   *time.Time
	update_time   *time.Time
	user          *string
	clearedFields map[string]struct{}
	domain        *int
	cleareddomain bool
	done          bool
	oldValue      func(context.Context) (*Delegation, error)
	predicates    []predicate.Delegation
}

var _ ent.Mutation = (*DelegationMutation)(nil)

// delegationOption allows management of the mutation configuration using functional options.
type delegationOption func(*DelegationMutation)

// newDelegationMutation creates new mutation for the Delegation entity.
func newDelegationMutation(c config, op Op, opts ...delegationOption) *DelegationMutation {
	m := &DelegationMutation{
		config:        c,
		op:            op,
		typ:           TypeDelegation,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withDelegationID sets the ID field of the mutation.
func withDelegationID(id int) delegationOption {
	return func(m *DelegationMutation) {
		var (
			err   error
			once  sync.Once
			value *Delegation
		)
		m.oldValue = func(ctx context.Context) (*Delegation, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Delegation.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withDelegation sets the old Delegation of the mutation.
func withDelegation(node *Delegation) delegationOption {
	return func(m *DelegationMutation) {
		m.oldValue = func(context.Context) (*Delegation, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m DelegationMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m DelegationMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *DelegationMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *DelegationMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Delegation.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *DelegationMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *DelegationMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Delegation entity.
// If the Delegation object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DelegationMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *DelegationMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *DelegationMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *DelegationMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Delegation entity.
// If the Delegation object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DelegationMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *DelegationMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetUser sets the "user" field.
func (m *DelegationMutation) SetUser(s string) {
	m.user = &s
}

// User returns the value of the "user" field in the mutation.
func (m *DelegationMutation) User() (r string, exists bool) {
	v := m.user
	if v == nil {
		return
	}
	return *v, true
}

// OldUser returns the old "user" field's value of the Delegation entity.
// If the Delegation object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DelegationMutation) OldUser(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUser is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUser requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUser: %w", err)
	}
	return oldValue.User, nil
}

// ResetUser resets all changes to the "user" field.
func (m *DelegationMutation) ResetUser() {
	m.user = nil
}

// SetDomainID sets the "domain" edge to the Domain entity by id.
func (m *DelegationMutation) SetDomainID(id int) {
	m.domain = &id
}

// ClearDomain clears the "domain" edge to the Domain entity.
func (m *DelegationMutation) ClearDomain() {
	m.cleareddomain = true
}

// DomainCleared reports if the "domain" edge to the Domain entity was cleared.
func (m *DelegationMutation) DomainCleared() bool {
	return m.cleareddomain
}

// DomainID returns the "domain" edge ID in the mutation.
func (m *DelegationMutation) DomainID() (id int, exists bool) {
	if m.domain != nil {
		return *m.domain, true
	}
	return
}

// DomainIDs returns the "domain" edge IDs in the mutation.
// Note that IDs always returns len(IDs) <= 1 for unique edges, and you should use
// DomainID instead. It exists only for internal usage by the builders.
func (m *DelegationMutation) DomainIDs() (ids []int) {
	if id := m.domain; id != nil {
		ids = append(ids, *id)
	}
	return
}

// ResetDomain resets all changes to the "domain" edge.
func (m *DelegationMutation) ResetDomain() {
	m.domain = nil
	m.cleareddomain = false
}

// Where appends a list predicates to the DelegationMutation builder.
func (m *DelegationMutation) Where(ps ...predicate.Delegation) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *DelegationMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Delegation).
func (m *DelegationMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *DelegationMutation) Fields() []string {
	fields := make([]string, 0, 3)
	if m.create_time != nil {
		fields = append(fields, delegation.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, delegation.FieldUpdateTime)
	}
	if m.user != nil {
		fields = append(fields, delegation.FieldUser)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *DelegationMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case delegation.FieldCreateTime:
		return m.CreateTime()
	case delegation.FieldUpdateTime:
		return m.UpdateTime()
	case delegation.FieldUser:
		return m.User()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *DelegationMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case delegation.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case delegation.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case delegation.FieldUser:
		return m.OldUser(ctx)
	}
	return nil, fmt.Errorf("unknown Delegation field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DelegationMutation) SetField(name string, value ent.Value) error {
	switch name {
	case delegation.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case delegation.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case delegation.FieldUser:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUser(v)
		return nil
	}
	return fmt.Errorf("unknown Delegation field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *DelegationMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *DelegationMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DelegationMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Delegation numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *DelegationMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *DelegationMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *DelegationMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Delegation nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *DelegationMutation) ResetField(name string) error {
	switch name {
	case delegation.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case delegation.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case delegation.FieldUser:
		m.ResetUser()
		return nil
	}
	return fmt.Errorf("unknown Delegation field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *DelegationMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.domain != nil {
		edges = append(edges, delegation.EdgeDomain)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *DelegationMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case delegation.EdgeDomain:
		if id := m.domain; id != nil {
			return []ent.Value{*id}
		}
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *DelegationMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *DelegationMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *DelegationMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.cleareddomain {
		edges = append(edges, delegation.EdgeDomain)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *DelegationMutation) EdgeCleared(name string) bool {
	switch name {
	case delegation.EdgeDomain:
		return m.cleareddomain
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *DelegationMutation) ClearEdge(name string) error {
	switch name {
	case delegation.EdgeDomain:
		m.ClearDomain()
		return nil
	}
	return fmt.Errorf("unknown Delegation unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *DelegationMutation) ResetEdge(name string) error {
	switch name {
	case delegation.EdgeDomain:
		m.ResetDomain()
		return nil
	}
	return fmt.Errorf("unknown Delegation edge %s", name)
}

// DomainMutation represents an operation that mutates the Domain nodes in the graph.
type DomainMutation struct {
	config
	op                 Op
	typ                string
	id                 *int
	create_time        *time.Time
	update_time        *time.Time
	fqdn               *string
	owner              *string
	approved           *bool
	clearedFields      map[string]struct{}
	delegations        map[int]struct{}
	removeddelegations map[int]struct{}
	cleareddelegations bool
	done               bool
	oldValue           func(context.Context) (*Domain, error)
	predicates         []predicate.Domain
}

var _ ent.Mutation = (*DomainMutation)(nil)

// domainOption allows management of the mutation configuration using functional options.
type domainOption func(*DomainMutation)

// newDomainMutation creates new mutation for the Domain entity.
func newDomainMutation(c config, op Op, opts ...domainOption) *DomainMutation {
	m := &DomainMutation{
		config:        c,
		op:            op,
		typ:           TypeDomain,
		clearedFields: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// withDomainID sets the ID field of the mutation.
func withDomainID(id int) domainOption {
	return func(m *DomainMutation) {
		var (
			err   error
			once  sync.Once
			value *Domain
		)
		m.oldValue = func(ctx context.Context) (*Domain, error) {
			once.Do(func() {
				if m.done {
					err = errors.New("querying old values post mutation is not allowed")
				} else {
					value, err = m.Client().Domain.Get(ctx, id)
				}
			})
			return value, err
		}
		m.id = &id
	}
}

// withDomain sets the old Domain of the mutation.
func withDomain(node *Domain) domainOption {
	return func(m *DomainMutation) {
		m.oldValue = func(context.Context) (*Domain, error) {
			return node, nil
		}
		m.id = &node.ID
	}
}

// Client returns a new `ent.Client` from the mutation. If the mutation was
// executed in a transaction (ent.Tx), a transactional client is returned.
func (m DomainMutation) Client() *Client {
	client := &Client{config: m.config}
	client.init()
	return client
}

// Tx returns an `ent.Tx` for mutations that were executed in transactions;
// it returns an error otherwise.
func (m DomainMutation) Tx() (*Tx, error) {
	if _, ok := m.driver.(*txDriver); !ok {
		return nil, errors.New("ent: mutation is not running in a transaction")
	}
	tx := &Tx{config: m.config}
	tx.init()
	return tx, nil
}

// ID returns the ID value in the mutation. Note that the ID is only available
// if it was provided to the builder or after it was returned from the database.
func (m *DomainMutation) ID() (id int, exists bool) {
	if m.id == nil {
		return
	}
	return *m.id, true
}

// IDs queries the database and returns the entity ids that match the mutation's predicate.
// That means, if the mutation is applied within a transaction with an isolation level such
// as sql.LevelSerializable, the returned ids match the ids of the rows that will be updated
// or updated by the mutation.
func (m *DomainMutation) IDs(ctx context.Context) ([]int, error) {
	switch {
	case m.op.Is(OpUpdateOne | OpDeleteOne):
		id, exists := m.ID()
		if exists {
			return []int{id}, nil
		}
		fallthrough
	case m.op.Is(OpUpdate | OpDelete):
		return m.Client().Domain.Query().Where(m.predicates...).IDs(ctx)
	default:
		return nil, fmt.Errorf("IDs is not allowed on %s operations", m.op)
	}
}

// SetCreateTime sets the "create_time" field.
func (m *DomainMutation) SetCreateTime(t time.Time) {
	m.create_time = &t
}

// CreateTime returns the value of the "create_time" field in the mutation.
func (m *DomainMutation) CreateTime() (r time.Time, exists bool) {
	v := m.create_time
	if v == nil {
		return
	}
	return *v, true
}

// OldCreateTime returns the old "create_time" field's value of the Domain entity.
// If the Domain object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DomainMutation) OldCreateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldCreateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldCreateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldCreateTime: %w", err)
	}
	return oldValue.CreateTime, nil
}

// ResetCreateTime resets all changes to the "create_time" field.
func (m *DomainMutation) ResetCreateTime() {
	m.create_time = nil
}

// SetUpdateTime sets the "update_time" field.
func (m *DomainMutation) SetUpdateTime(t time.Time) {
	m.update_time = &t
}

// UpdateTime returns the value of the "update_time" field in the mutation.
func (m *DomainMutation) UpdateTime() (r time.Time, exists bool) {
	v := m.update_time
	if v == nil {
		return
	}
	return *v, true
}

// OldUpdateTime returns the old "update_time" field's value of the Domain entity.
// If the Domain object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DomainMutation) OldUpdateTime(ctx context.Context) (v time.Time, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldUpdateTime is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldUpdateTime requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldUpdateTime: %w", err)
	}
	return oldValue.UpdateTime, nil
}

// ResetUpdateTime resets all changes to the "update_time" field.
func (m *DomainMutation) ResetUpdateTime() {
	m.update_time = nil
}

// SetFqdn sets the "fqdn" field.
func (m *DomainMutation) SetFqdn(s string) {
	m.fqdn = &s
}

// Fqdn returns the value of the "fqdn" field in the mutation.
func (m *DomainMutation) Fqdn() (r string, exists bool) {
	v := m.fqdn
	if v == nil {
		return
	}
	return *v, true
}

// OldFqdn returns the old "fqdn" field's value of the Domain entity.
// If the Domain object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DomainMutation) OldFqdn(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldFqdn is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldFqdn requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldFqdn: %w", err)
	}
	return oldValue.Fqdn, nil
}

// ResetFqdn resets all changes to the "fqdn" field.
func (m *DomainMutation) ResetFqdn() {
	m.fqdn = nil
}

// SetOwner sets the "owner" field.
func (m *DomainMutation) SetOwner(s string) {
	m.owner = &s
}

// Owner returns the value of the "owner" field in the mutation.
func (m *DomainMutation) Owner() (r string, exists bool) {
	v := m.owner
	if v == nil {
		return
	}
	return *v, true
}

// OldOwner returns the old "owner" field's value of the Domain entity.
// If the Domain object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DomainMutation) OldOwner(ctx context.Context) (v string, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldOwner is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldOwner requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldOwner: %w", err)
	}
	return oldValue.Owner, nil
}

// ResetOwner resets all changes to the "owner" field.
func (m *DomainMutation) ResetOwner() {
	m.owner = nil
}

// SetApproved sets the "approved" field.
func (m *DomainMutation) SetApproved(b bool) {
	m.approved = &b
}

// Approved returns the value of the "approved" field in the mutation.
func (m *DomainMutation) Approved() (r bool, exists bool) {
	v := m.approved
	if v == nil {
		return
	}
	return *v, true
}

// OldApproved returns the old "approved" field's value of the Domain entity.
// If the Domain object wasn't provided to the builder, the object is fetched from the database.
// An error is returned if the mutation operation is not UpdateOne, or the database query fails.
func (m *DomainMutation) OldApproved(ctx context.Context) (v bool, err error) {
	if !m.op.Is(OpUpdateOne) {
		return v, errors.New("OldApproved is only allowed on UpdateOne operations")
	}
	if m.id == nil || m.oldValue == nil {
		return v, errors.New("OldApproved requires an ID field in the mutation")
	}
	oldValue, err := m.oldValue(ctx)
	if err != nil {
		return v, fmt.Errorf("querying old value for OldApproved: %w", err)
	}
	return oldValue.Approved, nil
}

// ResetApproved resets all changes to the "approved" field.
func (m *DomainMutation) ResetApproved() {
	m.approved = nil
}

// AddDelegationIDs adds the "delegations" edge to the Delegation entity by ids.
func (m *DomainMutation) AddDelegationIDs(ids ...int) {
	if m.delegations == nil {
		m.delegations = make(map[int]struct{})
	}
	for i := range ids {
		m.delegations[ids[i]] = struct{}{}
	}
}

// ClearDelegations clears the "delegations" edge to the Delegation entity.
func (m *DomainMutation) ClearDelegations() {
	m.cleareddelegations = true
}

// DelegationsCleared reports if the "delegations" edge to the Delegation entity was cleared.
func (m *DomainMutation) DelegationsCleared() bool {
	return m.cleareddelegations
}

// RemoveDelegationIDs removes the "delegations" edge to the Delegation entity by IDs.
func (m *DomainMutation) RemoveDelegationIDs(ids ...int) {
	if m.removeddelegations == nil {
		m.removeddelegations = make(map[int]struct{})
	}
	for i := range ids {
		delete(m.delegations, ids[i])
		m.removeddelegations[ids[i]] = struct{}{}
	}
}

// RemovedDelegations returns the removed IDs of the "delegations" edge to the Delegation entity.
func (m *DomainMutation) RemovedDelegationsIDs() (ids []int) {
	for id := range m.removeddelegations {
		ids = append(ids, id)
	}
	return
}

// DelegationsIDs returns the "delegations" edge IDs in the mutation.
func (m *DomainMutation) DelegationsIDs() (ids []int) {
	for id := range m.delegations {
		ids = append(ids, id)
	}
	return
}

// ResetDelegations resets all changes to the "delegations" edge.
func (m *DomainMutation) ResetDelegations() {
	m.delegations = nil
	m.cleareddelegations = false
	m.removeddelegations = nil
}

// Where appends a list predicates to the DomainMutation builder.
func (m *DomainMutation) Where(ps ...predicate.Domain) {
	m.predicates = append(m.predicates, ps...)
}

// Op returns the operation name.
func (m *DomainMutation) Op() Op {
	return m.op
}

// Type returns the node type of this mutation (Domain).
func (m *DomainMutation) Type() string {
	return m.typ
}

// Fields returns all fields that were changed during this mutation. Note that in
// order to get all numeric fields that were incremented/decremented, call
// AddedFields().
func (m *DomainMutation) Fields() []string {
	fields := make([]string, 0, 5)
	if m.create_time != nil {
		fields = append(fields, domain.FieldCreateTime)
	}
	if m.update_time != nil {
		fields = append(fields, domain.FieldUpdateTime)
	}
	if m.fqdn != nil {
		fields = append(fields, domain.FieldFqdn)
	}
	if m.owner != nil {
		fields = append(fields, domain.FieldOwner)
	}
	if m.approved != nil {
		fields = append(fields, domain.FieldApproved)
	}
	return fields
}

// Field returns the value of a field with the given name. The second boolean
// return value indicates that this field was not set, or was not defined in the
// schema.
func (m *DomainMutation) Field(name string) (ent.Value, bool) {
	switch name {
	case domain.FieldCreateTime:
		return m.CreateTime()
	case domain.FieldUpdateTime:
		return m.UpdateTime()
	case domain.FieldFqdn:
		return m.Fqdn()
	case domain.FieldOwner:
		return m.Owner()
	case domain.FieldApproved:
		return m.Approved()
	}
	return nil, false
}

// OldField returns the old value of the field from the database. An error is
// returned if the mutation operation is not UpdateOne, or the query to the
// database failed.
func (m *DomainMutation) OldField(ctx context.Context, name string) (ent.Value, error) {
	switch name {
	case domain.FieldCreateTime:
		return m.OldCreateTime(ctx)
	case domain.FieldUpdateTime:
		return m.OldUpdateTime(ctx)
	case domain.FieldFqdn:
		return m.OldFqdn(ctx)
	case domain.FieldOwner:
		return m.OldOwner(ctx)
	case domain.FieldApproved:
		return m.OldApproved(ctx)
	}
	return nil, fmt.Errorf("unknown Domain field %s", name)
}

// SetField sets the value of a field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DomainMutation) SetField(name string, value ent.Value) error {
	switch name {
	case domain.FieldCreateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetCreateTime(v)
		return nil
	case domain.FieldUpdateTime:
		v, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetUpdateTime(v)
		return nil
	case domain.FieldFqdn:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetFqdn(v)
		return nil
	case domain.FieldOwner:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetOwner(v)
		return nil
	case domain.FieldApproved:
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("unexpected type %T for field %s", value, name)
		}
		m.SetApproved(v)
		return nil
	}
	return fmt.Errorf("unknown Domain field %s", name)
}

// AddedFields returns all numeric fields that were incremented/decremented during
// this mutation.
func (m *DomainMutation) AddedFields() []string {
	return nil
}

// AddedField returns the numeric value that was incremented/decremented on a field
// with the given name. The second boolean return value indicates that this field
// was not set, or was not defined in the schema.
func (m *DomainMutation) AddedField(name string) (ent.Value, bool) {
	return nil, false
}

// AddField adds the value to the field with the given name. It returns an error if
// the field is not defined in the schema, or if the type mismatched the field
// type.
func (m *DomainMutation) AddField(name string, value ent.Value) error {
	switch name {
	}
	return fmt.Errorf("unknown Domain numeric field %s", name)
}

// ClearedFields returns all nullable fields that were cleared during this
// mutation.
func (m *DomainMutation) ClearedFields() []string {
	return nil
}

// FieldCleared returns a boolean indicating if a field with the given name was
// cleared in this mutation.
func (m *DomainMutation) FieldCleared(name string) bool {
	_, ok := m.clearedFields[name]
	return ok
}

// ClearField clears the value of the field with the given name. It returns an
// error if the field is not defined in the schema.
func (m *DomainMutation) ClearField(name string) error {
	return fmt.Errorf("unknown Domain nullable field %s", name)
}

// ResetField resets all changes in the mutation for the field with the given name.
// It returns an error if the field is not defined in the schema.
func (m *DomainMutation) ResetField(name string) error {
	switch name {
	case domain.FieldCreateTime:
		m.ResetCreateTime()
		return nil
	case domain.FieldUpdateTime:
		m.ResetUpdateTime()
		return nil
	case domain.FieldFqdn:
		m.ResetFqdn()
		return nil
	case domain.FieldOwner:
		m.ResetOwner()
		return nil
	case domain.FieldApproved:
		m.ResetApproved()
		return nil
	}
	return fmt.Errorf("unknown Domain field %s", name)
}

// AddedEdges returns all edge names that were set/added in this mutation.
func (m *DomainMutation) AddedEdges() []string {
	edges := make([]string, 0, 1)
	if m.delegations != nil {
		edges = append(edges, domain.EdgeDelegations)
	}
	return edges
}

// AddedIDs returns all IDs (to other nodes) that were added for the given edge
// name in this mutation.
func (m *DomainMutation) AddedIDs(name string) []ent.Value {
	switch name {
	case domain.EdgeDelegations:
		ids := make([]ent.Value, 0, len(m.delegations))
		for id := range m.delegations {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// RemovedEdges returns all edge names that were removed in this mutation.
func (m *DomainMutation) RemovedEdges() []string {
	edges := make([]string, 0, 1)
	if m.removeddelegations != nil {
		edges = append(edges, domain.EdgeDelegations)
	}
	return edges
}

// RemovedIDs returns all IDs (to other nodes) that were removed for the edge with
// the given name in this mutation.
func (m *DomainMutation) RemovedIDs(name string) []ent.Value {
	switch name {
	case domain.EdgeDelegations:
		ids := make([]ent.Value, 0, len(m.removeddelegations))
		for id := range m.removeddelegations {
			ids = append(ids, id)
		}
		return ids
	}
	return nil
}

// ClearedEdges returns all edge names that were cleared in this mutation.
func (m *DomainMutation) ClearedEdges() []string {
	edges := make([]string, 0, 1)
	if m.cleareddelegations {
		edges = append(edges, domain.EdgeDelegations)
	}
	return edges
}

// EdgeCleared returns a boolean which indicates if the edge with the given name
// was cleared in this mutation.
func (m *DomainMutation) EdgeCleared(name string) bool {
	switch name {
	case domain.EdgeDelegations:
		return m.cleareddelegations
	}
	return false
}

// ClearEdge clears the value of the edge with the given name. It returns an error
// if that edge is not defined in the schema.
func (m *DomainMutation) ClearEdge(name string) error {
	switch name {
	}
	return fmt.Errorf("unknown Domain unique edge %s", name)
}

// ResetEdge resets all changes to the edge with the given name in this mutation.
// It returns an error if the edge is not defined in the schema.
func (m *DomainMutation) ResetEdge(name string) error {
	switch name {
	case domain.EdgeDelegations:
		m.ResetDelegations()
		return nil
	}
	return fmt.Errorf("unknown Domain edge %s", name)
}
