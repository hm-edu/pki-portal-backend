// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/hm-edu/pki-service/ent/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/domain"
	"github.com/hm-edu/pki-service/ent/smimecertificate"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Certificate is the client for interacting with the Certificate builders.
	Certificate *CertificateClient
	// Domain is the client for interacting with the Domain builders.
	Domain *DomainClient
	// SmimeCertificate is the client for interacting with the SmimeCertificate builders.
	SmimeCertificate *SmimeCertificateClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Certificate = NewCertificateClient(c.config)
	c.Domain = NewDomainClient(c.config)
	c.SmimeCertificate = NewSmimeCertificateClient(c.config)
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
		ctx:              ctx,
		config:           cfg,
		Certificate:      NewCertificateClient(cfg),
		Domain:           NewDomainClient(cfg),
		SmimeCertificate: NewSmimeCertificateClient(cfg),
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
		ctx:              ctx,
		config:           cfg,
		Certificate:      NewCertificateClient(cfg),
		Domain:           NewDomainClient(cfg),
		SmimeCertificate: NewSmimeCertificateClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Certificate.
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
	c.Certificate.Use(hooks...)
	c.Domain.Use(hooks...)
	c.SmimeCertificate.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Certificate.Intercept(interceptors...)
	c.Domain.Intercept(interceptors...)
	c.SmimeCertificate.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *CertificateMutation:
		return c.Certificate.mutate(ctx, m)
	case *DomainMutation:
		return c.Domain.mutate(ctx, m)
	case *SmimeCertificateMutation:
		return c.SmimeCertificate.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// CertificateClient is a client for the Certificate schema.
type CertificateClient struct {
	config
}

// NewCertificateClient returns a client for the Certificate from the given config.
func NewCertificateClient(c config) *CertificateClient {
	return &CertificateClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `certificate.Hooks(f(g(h())))`.
func (c *CertificateClient) Use(hooks ...Hook) {
	c.hooks.Certificate = append(c.hooks.Certificate, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `certificate.Intercept(f(g(h())))`.
func (c *CertificateClient) Intercept(interceptors ...Interceptor) {
	c.inters.Certificate = append(c.inters.Certificate, interceptors...)
}

// Create returns a builder for creating a Certificate entity.
func (c *CertificateClient) Create() *CertificateCreate {
	mutation := newCertificateMutation(c.config, OpCreate)
	return &CertificateCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Certificate entities.
func (c *CertificateClient) CreateBulk(builders ...*CertificateCreate) *CertificateCreateBulk {
	return &CertificateCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *CertificateClient) MapCreateBulk(slice any, setFunc func(*CertificateCreate, int)) *CertificateCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &CertificateCreateBulk{err: fmt.Errorf("calling to CertificateClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*CertificateCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &CertificateCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Certificate.
func (c *CertificateClient) Update() *CertificateUpdate {
	mutation := newCertificateMutation(c.config, OpUpdate)
	return &CertificateUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *CertificateClient) UpdateOne(ce *Certificate) *CertificateUpdateOne {
	mutation := newCertificateMutation(c.config, OpUpdateOne, withCertificate(ce))
	return &CertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *CertificateClient) UpdateOneID(id int) *CertificateUpdateOne {
	mutation := newCertificateMutation(c.config, OpUpdateOne, withCertificateID(id))
	return &CertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Certificate.
func (c *CertificateClient) Delete() *CertificateDelete {
	mutation := newCertificateMutation(c.config, OpDelete)
	return &CertificateDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *CertificateClient) DeleteOne(ce *Certificate) *CertificateDeleteOne {
	return c.DeleteOneID(ce.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *CertificateClient) DeleteOneID(id int) *CertificateDeleteOne {
	builder := c.Delete().Where(certificate.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &CertificateDeleteOne{builder}
}

// Query returns a query builder for Certificate.
func (c *CertificateClient) Query() *CertificateQuery {
	return &CertificateQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeCertificate},
		inters: c.Interceptors(),
	}
}

// Get returns a Certificate entity by its id.
func (c *CertificateClient) Get(ctx context.Context, id int) (*Certificate, error) {
	return c.Query().Where(certificate.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *CertificateClient) GetX(ctx context.Context, id int) *Certificate {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryDomains queries the domains edge of a Certificate.
func (c *CertificateClient) QueryDomains(ce *Certificate) *DomainQuery {
	query := (&DomainClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := ce.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(certificate.Table, certificate.FieldID, id),
			sqlgraph.To(domain.Table, domain.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, certificate.DomainsTable, certificate.DomainsPrimaryKey...),
		)
		fromV = sqlgraph.Neighbors(ce.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *CertificateClient) Hooks() []Hook {
	hooks := c.hooks.Certificate
	return append(hooks[:len(hooks):len(hooks)], certificate.Hooks[:]...)
}

// Interceptors returns the client interceptors.
func (c *CertificateClient) Interceptors() []Interceptor {
	return c.inters.Certificate
}

func (c *CertificateClient) mutate(ctx context.Context, m *CertificateMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&CertificateCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&CertificateUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&CertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&CertificateDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Certificate mutation op: %q", m.Op())
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

// QueryCertificates queries the certificates edge of a Domain.
func (c *DomainClient) QueryCertificates(d *Domain) *CertificateQuery {
	query := (&CertificateClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := d.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(domain.Table, domain.FieldID, id),
			sqlgraph.To(certificate.Table, certificate.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, domain.CertificatesTable, domain.CertificatesPrimaryKey...),
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

// SmimeCertificateClient is a client for the SmimeCertificate schema.
type SmimeCertificateClient struct {
	config
}

// NewSmimeCertificateClient returns a client for the SmimeCertificate from the given config.
func NewSmimeCertificateClient(c config) *SmimeCertificateClient {
	return &SmimeCertificateClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `smimecertificate.Hooks(f(g(h())))`.
func (c *SmimeCertificateClient) Use(hooks ...Hook) {
	c.hooks.SmimeCertificate = append(c.hooks.SmimeCertificate, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `smimecertificate.Intercept(f(g(h())))`.
func (c *SmimeCertificateClient) Intercept(interceptors ...Interceptor) {
	c.inters.SmimeCertificate = append(c.inters.SmimeCertificate, interceptors...)
}

// Create returns a builder for creating a SmimeCertificate entity.
func (c *SmimeCertificateClient) Create() *SmimeCertificateCreate {
	mutation := newSmimeCertificateMutation(c.config, OpCreate)
	return &SmimeCertificateCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of SmimeCertificate entities.
func (c *SmimeCertificateClient) CreateBulk(builders ...*SmimeCertificateCreate) *SmimeCertificateCreateBulk {
	return &SmimeCertificateCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *SmimeCertificateClient) MapCreateBulk(slice any, setFunc func(*SmimeCertificateCreate, int)) *SmimeCertificateCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &SmimeCertificateCreateBulk{err: fmt.Errorf("calling to SmimeCertificateClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*SmimeCertificateCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &SmimeCertificateCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for SmimeCertificate.
func (c *SmimeCertificateClient) Update() *SmimeCertificateUpdate {
	mutation := newSmimeCertificateMutation(c.config, OpUpdate)
	return &SmimeCertificateUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *SmimeCertificateClient) UpdateOne(sc *SmimeCertificate) *SmimeCertificateUpdateOne {
	mutation := newSmimeCertificateMutation(c.config, OpUpdateOne, withSmimeCertificate(sc))
	return &SmimeCertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *SmimeCertificateClient) UpdateOneID(id int) *SmimeCertificateUpdateOne {
	mutation := newSmimeCertificateMutation(c.config, OpUpdateOne, withSmimeCertificateID(id))
	return &SmimeCertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for SmimeCertificate.
func (c *SmimeCertificateClient) Delete() *SmimeCertificateDelete {
	mutation := newSmimeCertificateMutation(c.config, OpDelete)
	return &SmimeCertificateDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *SmimeCertificateClient) DeleteOne(sc *SmimeCertificate) *SmimeCertificateDeleteOne {
	return c.DeleteOneID(sc.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *SmimeCertificateClient) DeleteOneID(id int) *SmimeCertificateDeleteOne {
	builder := c.Delete().Where(smimecertificate.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &SmimeCertificateDeleteOne{builder}
}

// Query returns a query builder for SmimeCertificate.
func (c *SmimeCertificateClient) Query() *SmimeCertificateQuery {
	return &SmimeCertificateQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeSmimeCertificate},
		inters: c.Interceptors(),
	}
}

// Get returns a SmimeCertificate entity by its id.
func (c *SmimeCertificateClient) Get(ctx context.Context, id int) (*SmimeCertificate, error) {
	return c.Query().Where(smimecertificate.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *SmimeCertificateClient) GetX(ctx context.Context, id int) *SmimeCertificate {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *SmimeCertificateClient) Hooks() []Hook {
	hooks := c.hooks.SmimeCertificate
	return append(hooks[:len(hooks):len(hooks)], smimecertificate.Hooks[:]...)
}

// Interceptors returns the client interceptors.
func (c *SmimeCertificateClient) Interceptors() []Interceptor {
	return c.inters.SmimeCertificate
}

func (c *SmimeCertificateClient) mutate(ctx context.Context, m *SmimeCertificateMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&SmimeCertificateCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&SmimeCertificateUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&SmimeCertificateUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&SmimeCertificateDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown SmimeCertificate mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Certificate, Domain, SmimeCertificate []ent.Hook
	}
	inters struct {
		Certificate, Domain, SmimeCertificate []ent.Interceptor
	}
)
