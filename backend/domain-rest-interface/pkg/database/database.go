package database

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	"github.com/hm-edu/domain-rest-interface/ent"

	// Importing the pgx/v4/stdlib is required to create a pg database.
	_ "github.com/jackc/pgx/v4/stdlib"

	// Imprting the runtime is required to get the default hooks working.
	entsql "entgo.io/ent/dialect/sql"
	"go.uber.org/zap"
)

// DbInstance is a wrapper around a entgo database client.
type DbInstance struct {
	Db       *ent.Client
	Internal *sql.DB
}

// DB is a globally accessible domain instance.
var DB DbInstance

// Open creates a new databse client
func Open(log *zap.Logger, connectionString string) (*ent.Client, *sql.DB) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatal("Error connecting to database", zap.Error(err))
	}
	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), db
}

// ConnectDb establishs a new database connection.
func ConnectDb(log *zap.Logger, connectionString string) {

	client, db := Open(log, connectionString)

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal("failed creating schema resources", zap.Error(err))
	}

	DB = DbInstance{
		Db:       client,
		Internal: db,
	}
}
