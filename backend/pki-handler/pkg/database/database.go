package database

import (
	"context"

	"github.com/hm-edu/pki-handler/ent"
	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
	// Imprting the runtime is required to get the default hooks working.
	_ "github.com/hm-edu/pki-handler/ent/runtime"
	"go.uber.org/zap"
)

// DbInstance is a wrapper around a entgo database client.
type DbInstance struct {
	Db *ent.Client
}

// DB is a globally accessible domain instance.
var DB DbInstance

// ConnectDb establishs a new database connection.
func ConnectDb(log *zap.Logger) {

	client, err := ent.Open("sqlite3", "file:db.sqlite3?cache=shared&mode=rwc&_fk=1")

	if err != nil {
		log.Fatal("failed opening DB Connection", zap.Error(err))
	}

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal("failed creating schema resources", zap.Error(err))
	}

	DB = DbInstance{
		Db: client,
	}
}
