package store

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/ory/viper"
	"go.uber.org/zap"
)

// Preseed fills the database with given items
func (store *DomainStore) Preseed(logger *zap.Logger) {

	data, err := os.ReadFile(viper.GetString("preseed"))
	if err != nil {
		logger.Fatal("Error reading preseed file", zap.Error(err))
	}
	var preseed map[string]string
	err = json.Unmarshal(data, &preseed)
	if err != nil {
		logger.Fatal("Error unmarshalling preseed file", zap.Error(err))
	}
	for key, value := range preseed {
		exists, _ := database.DB.Db.Domain.Query().Where(domain.Fqdn(key)).Exist(context.Background())
		if exists {
			logger.Info("Domain already exists", zap.String("domain", key))
			continue
		}
		_, err = store.Create(context.Background(), &ent.Domain{Fqdn: key, Owner: value, Approved: true})
		if err != nil {
			logger.Fatal("Error creating domain", zap.Error(err))
		}
		logger.Info("Domain created", zap.String("domain", key))
	}
}
