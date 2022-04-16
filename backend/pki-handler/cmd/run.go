package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hm-edu/pki-handler/pkg/cfg"
	"github.com/hm-edu/pki-handler/pkg/database"
	"github.com/hm-edu/pki-handler/pkg/grpc"
	"github.com/hm-edu/portal-common/logging"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the servers",
	Long:  `Starts the HTTP and the GRPC server`,
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

		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load Sectigo config
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

		stopCh := signals.SetupSignalHandler()

		sectigoCfg.CheckSectigoConfiguration()

		database.ConnectDb(logger)
		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, &sectigoCfg, database.DB.Db)
			grpcSrv.ListenAndServe(stopCh)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("host", "", "Host to bind service to")
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("sectigo_user", "", "The sectigo user")
	runCmd.Flags().String("sectigo_password", "", "The password for the sectigo user")
	runCmd.Flags().String("sectigo_customeruri", "", "The sectigo customerUri")
	runCmd.Flags().Int("smime_profile", 0, "The (default) smime profile id")
	runCmd.Flags().Int("smime_org_id", 0, "The (default) org id")
	runCmd.Flags().Int("smime_term", 0, "The (default) lifetime")
	runCmd.Flags().Int("smime_key_length", 0, "The (expected) key length")
	runCmd.Flags().String("smime_key_type", "", "The (expected) key type")
	runCmd.Flags().Int("ssl_profile", 0, "The (default) ssl profile id")
	runCmd.Flags().Int("ssl_org_id", 0, "The (default) ssl org id")
	runCmd.Flags().Int("ssl_term", 0, "The (default) ssl lifetime")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
