package cmd

import (
	"context"
	"net/http"

	"github.com/hm-edu/pki-service/pkg/cfg"
	"github.com/hm-edu/pki-service/pkg/database"
	"github.com/hm-edu/pki-service/pkg/worker"
	"github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the database with the Sectigo API",
	Long:  `Adds all missing entries from the Sectigo API to the database`,
	Run: func(cmd *cobra.Command, _ []string) {

		logger, deferFunc, viper := api.PrepareEnv(cmd)
		defer deferFunc(logger)

		// load HTTP server config
		var sectigoCfg cfg.SectigoConfiguration
		if err := viper.Unmarshal(&sectigoCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "pki-service")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		database.ConnectDb(logger, viper.GetString("db"))
		syncer := worker.Syncer{
			Db:     database.DB.Db,
			Client: sectigo.NewClient(http.DefaultClient, logger, sectigoCfg.User, sectigoCfg.Password, sectigoCfg.CustomerURI),
		}
		syncer.SyncAllCertificates()

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("sectigo_user", "", "The sectigo user")
	syncCmd.Flags().String("sectigo_password", "", "The password for the sectigo user")
	syncCmd.Flags().String("sectigo_customeruri", "", "The sectigo customerUri")
	syncCmd.Flags().String("db", "", "connection string for the database")
	syncCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
