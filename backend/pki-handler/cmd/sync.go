package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hm-edu/pki-handler/pkg/cfg"
	"github.com/hm-edu/pki-handler/pkg/database"
	"github.com/hm-edu/pki-handler/pkg/worker"
	"github.com/hm-edu/portal-common/logging"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// runCmd represents the run command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the database with the Sectigo API",
	Long:  `Adds all missing entries from the Sectigo API to the database`,
	Run: func(cmd *cobra.Command, args []string) {

		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
		err = godotenv.Load()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Hint: %s\n", err.Error())
		}
		viper.AutomaticEnv()

		// configure logging
		logger, _ := logging.InitZap(viper.GetString("level"))
		zap.ReplaceGlobals(logger)
		defer func(logger *zap.Logger) {
			_ = logger.Sync()
		}(logger)
		stdLog := zap.RedirectStdLog(logger)
		defer stdLog()

		// load HTTP server config
		var sectigoCfg cfg.SectigoConfiguration
		if err := viper.Unmarshal(&sectigoCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "pki-handler")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		database.ConnectDb(logger)
		syncer := worker.Syncer{
			Db:     database.DB.Db,
			Client: sectigo.NewClient(http.DefaultClient, logger, sectigoCfg.User, sectigoCfg.Password, sectigoCfg.CustomerURI),
		}
		syncer.SyncCertificates()

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("sectigo_user", "", "The sectigo user")
	syncCmd.Flags().String("sectigo_password", "", "The password for the sectigo user")
	syncCmd.Flags().String("sectigo_customeruri", "", "The sectigo customerUri")
	syncCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
