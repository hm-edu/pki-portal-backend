package database

import (
	"context"
	"database/sql"

	"github.com/hm-edu/eab-rest-interface/ent"

	"github.com/smallstep/certificates/acme/db/nosql"
	nosqlDb "github.com/smallstep/nosql"

	// Importing the pgx/v4/stdlib is required to create a pg database.
	_ "github.com/jackc/pgx/v4/stdlib"

	// Imprting the runtime is required to get the default hooks working.
	entsql "entgo.io/ent/dialect/sql"

	"entgo.io/ent/dialect"

	"go.uber.org/zap"
)

// DbInstance is a wrapper around a entgo database client.
type DbInstance struct {
	Db            *ent.Client
	Internal      *sql.DB
	NoSQL         *nosql.DB
	NoSQLInternal nosqlDb.DB
}

// DB is a globally accessible domain instance.
var DB DbInstance

func openNoSQL(log *zap.Logger, connectionString string) (*nosql.DB, nosqlDb.DB) {
	connection, err := nosqlDb.New("postgres", connectionString)
	if err != nil {
		log.Fatal("Error connecting to database", zap.Error(err))
		return nil, nil
	}
	db, err := nosql.New(connection)
	if err != nil {
		log.Fatal("Error connecting to database", zap.Error(err))
		return nil, nil
	}
	return db, connection
}

// open creates a new databse client
func open(log *zap.Logger, connectionString string) (*ent.Client, *sql.DB) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatal("Error connecting to database", zap.Error(err))
		return nil, nil
	}
	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), db
}

// ConnectDb establishs a new database connection.
func ConnectDb(log *zap.Logger, mappingConnectionString, smallStep string) {

	client, db := open(log, mappingConnectionString)

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal("failed creating schema resources", zap.Error(err))
	}

	noSQLClient, noSQLDB := openNoSQL(log, smallStep)

	DB = DbInstance{
		Db:            client,
		Internal:      db,
		NoSQL:         noSQLClient,
		NoSQLInternal: noSQLDB,
	}
}
