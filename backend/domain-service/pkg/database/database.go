package database

import (
	"context"

	"github.com/hm-edu/domain-service/ent"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type DbInstance struct {
	Db *ent.Client
}

var DB DbInstance

// ConnectDb
func ConnectDb(log *zap.Logger, dev bool) {

	client, err := ent.Open("sqlite3", "file:db.sqlite3?cache=shared&_fk=1")

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
