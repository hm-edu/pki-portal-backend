// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/hm-edu/domain-rest-interface/ent/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/hm-edu/domain-rest-interface/ent/delegation"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Delegation is the client for interacting with the Delegation builders.
	Delegation *DelegationClient
	// Domain is the client for interacting with the Domain builders.
	Domain *DomainClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Delegation = NewDelegationClient(c.config)
	c.Domain = NewDomainClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("ent: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:        ctx,
		config:     cfg,
		Delegation: NewDelegationClient(cfg),
		Domain:     NewDomainClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:        ctx,
		config:     cfg,
		Delegation: NewDelegationClient(cfg),
		Domain:     NewDomainClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Delegation.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Delegation.Use(hooks...)
	c.Domain.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Delegation.Intercept(interceptors...)
	c.Domain.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *DelegationMutation:
		return c.Delegation.mutate(ctx, m)
	case *DomainMutation:
		return c.Domain.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// DelegationClient is a client for the Delegation schema.
type DelegationClient struct {
	config
}

// NewDelegationClient returns a client for the Delegation from the given config.
func NewDelegationClient(c config) *DelegationClient {
	return &DelegationClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `delegation.Hooks(f(g(h())))`.
func (c *DelegationClient) Use(hooks ...Hook) {
	c.hooks.Delegation = append(c.hooks.Delegation, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `delegation.Intercept(f(g(h())))`.
func (c *DelegationClient) Intercept(interceptors ...Interceptor) {
	c.inters.Delegation = append(c.inters.Delegation, interceptors...)
}

// Create returns a builder for creating a Delegation entity.
func (c *DelegationClient) Create() *DelegationCreate {
	mutation := newDelegationMutation(c.config, OpCreate)
	return &DelegationCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Delegation entities.
func (c *DelegationClient) CreateBulk(builders ...*DelegationCreate) *DelegationCreateBulk {
	return &DelegationCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *DelegationClient) MapCreateBulk(slice any, setFunc func(*DelegationCreate, int)) *DelegationCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &DelegationCreateBulk{err: fmt.Errorf("calling to DelegationClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*DelegationCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &DelegationCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Delegation.
func (c *DelegationClient) Update() *DelegationUpdate {
	mutation := newDelegationMutation(c.config, OpUpdate)
	return &DelegationUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *DelegationClient) UpdateOne(d *Delegation) *DelegationUpdateOne {
	mutation := newDelegationMutation(c.config, OpUpdateOne, withDelegation(d))
	return &DelegationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *DelegationClient) UpdateOneID(id int) *DelegationUpdateOne {
	mutation := newDelegationMutation(c.config, OpUpdateOne, withDelegationID(id))
	return &DelegationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Delegation.
func (c *DelegationClient) Delete() *DelegationDelete {
	mutation := newDelegationMutation(c.config, OpDelete)
	return &DelegationDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *DelegationClient) DeleteOne(d *Delegation) *DelegationDeleteOne {
	return c.DeleteOneID(d.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *DelegationClient) DeleteOneID(id int) *DelegationDeleteOne {
	builder := c.Delete().Where(delegation.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &DelegationDeleteOne{builder}
}

// Query returns a query builder for Delegation.
func (c *DelegationClient) Query() *DelegationQuery {
	return &DelegationQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeDelegation},
		inters: c.Interceptors(),
	}
}

// Get returns a Delegation entity by its id.
func (c *DelegationClient) Get(ctx context.Context, id int) (*Delegation, error) {
	return c.Query().Where(delegation.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *DelegationClient) GetX(ctx context.Context, id int) *Delegation {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryDomain queries the domain edge of a Delegation.
func (c *DelegationClient) QueryDomain(d *Delegation) *DomainQuery {
	query := (&DomainClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := d.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(delegation.Table, delegation.FieldID, id),
			sqlgraph.To(domain.Table, domain.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, delegation.DomainTable, delegation.DomainColumn),
		)
		fromV = sqlgraph.Neighbors(d.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *DelegationClient) Hooks() []Hook {
	return c.hooks.Delegation
}

// Interceptors returns the client interceptors.
func (c *DelegationClient) Interceptors() []Interceptor {
	return c.inters.Delegation
}

func (c *DelegationClient) mutate(ctx context.Context, m *DelegationMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&DelegationCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&DelegationUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&DelegationUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&DelegationDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Delegation mutation op: %q", m.Op())
	}
}

// DomainClient is a client for the Domain schema.
type DomainClient struct {
	config
}

// NewDomainClient returns a client for the Domain from the given config.
func NewDomainClient(c config) *DomainClient {
	return &DomainClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `domain.Hooks(f(g(h())))`.
func (c *DomainClient) Use(hooks ...Hook) {
	c.hooks.Domain = append(c.hooks.Domain, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `domain.Intercept(f(g(h())))`.
func (c *DomainClient) Intercept(interceptors ...Interceptor) {
	c.inters.Domain = append(c.inters.Domain, interceptors...)
}

// Create returns a builder for creating a Domain entity.
func (c *DomainClient) Create() *DomainCreate {
	mutation := newDomainMutation(c.config, OpCreate)
	return &DomainCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Domain entities.
func (c *DomainClient) CreateBulk(builders ...*DomainCreate) *DomainCreateBulk {
	return &DomainCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *DomainClient) MapCreateBulk(slice any, setFunc func(*DomainCreate, int)) *DomainCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &DomainCreateBulk{err: fmt.Errorf("calling to DomainClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*DomainCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &DomainCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Domain.
func (c *DomainClient) Update() *DomainUpdate {
	mutation := newDomainMutation(c.config, OpUpdate)
	return &DomainUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *DomainClient) UpdateOne(d *Domain) *DomainUpdateOne {
	mutation := newDomainMutation(c.config, OpUpdateOne, withDomain(d))
	return &DomainUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *DomainClient) UpdateOneID(id int) *DomainUpdateOne {
	mutation := newDomainMutation(c.config, OpUpdateOne, withDomainID(id))
	return &DomainUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Domain.
func (c *DomainClient) Delete() *DomainDelete {
	mutation := newDomainMutation(c.config, OpDelete)
	return &DomainDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *DomainClient) DeleteOne(d *Domain) *DomainDeleteOne {
	return c.DeleteOneID(d.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *DomainClient) DeleteOneID(id int) *DomainDeleteOne {
	builder := c.Delete().Where(domain.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &DomainDeleteOne{builder}
}

// Query returns a query builder for Domain.
func (c *DomainClient) Query() *DomainQuery {
	return &DomainQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeDomain},
		inters: c.Interceptors(),
	}
}

// Get returns a Domain entity by its id.
func (c *DomainClient) Get(ctx context.Context, id int) (*Domain, error) {
	return c.Query().Where(domain.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *DomainClient) GetX(ctx context.Context, id int) *Domain {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryDelegations queries the delegations edge of a Domain.
func (c *DomainClient) QueryDelegations(d *Domain) *DelegationQuery {
	query := (&DelegationClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := d.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(domain.Table, domain.FieldID, id),
			sqlgraph.To(delegation.Table, delegation.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, domain.DelegationsTable, domain.DelegationsColumn),
		)
		fromV = sqlgraph.Neighbors(d.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *DomainClient) Hooks() []Hook {
	return c.hooks.Domain
}

// Interceptors returns the client interceptors.
func (c *DomainClient) Interceptors() []Interceptor {
	return c.inters.Domain
}

func (c *DomainClient) mutate(ctx context.Context, m *DomainMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&DomainCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&DomainUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&DomainUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&DomainDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Domain mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Delegation, Domain []ent.Hook
	}
	inters struct {
		Delegation, Domain []ent.Interceptor
	}
)
