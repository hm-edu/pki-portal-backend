package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/hm-edu/domain-rest-interface/pkg/grpc"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	commonApi "github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runCmd represents the run command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the domain storage and fills in some predefined datasets",
	Run: func(cmd *cobra.Command, args []string) {

		logger, deferFunc, viper := commonApi.PrepareEnv(cmd)
		defer deferFunc(logger)

		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load HTTP server config
		var srvCfg commonApi.Config
		if err := viper.Unmarshal(&srvCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "domain-rest-interface")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		database.ConnectDb(logger, viper.GetString("db"))

		store := store.NewDomainStore(database.DB.Db)

		data, err := os.ReadFile(viper.GetString("config"))
		if err != nil {
			logger.Fatal("Error reading config file", zap.Error(err))
		}
		var config map[string]string
		err = json.Unmarshal(data, &config)
		if err != nil {
			logger.Fatal("Error unmarshalling config file", zap.Error(err))
		}
		for key, value := range config {
			exists, _ := database.DB.Db.Domain.Query().Where(domain.Fqdn(key)).Exist(context.Background())
			if exists {
				continue
			}
			_, err = store.Create(context.Background(), &ent.Domain{Fqdn: key, Owner: value, Approved: true})
			if err != nil {
				logger.Fatal("Error creating domain", zap.Error(err))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("db", "", "connection string for the database")
	initCmd.Flags().String("config", "", "config file")
	initCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
