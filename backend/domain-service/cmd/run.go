/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hm-edu/domain-service/pkg/api"
	"github.com/hm-edu/domain-service/pkg/database"
	"github.com/hm-edu/domain-service/pkg/grpc"
	"github.com/hm-edu/domain-service/pkg/store"
	commonApi "github.com/hm-edu/portal-common/api"
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
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
			os.Exit(1)
		}
		err = godotenv.Load()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
			os.Exit(1)
		}
		viper.AutomaticEnv()

		// configure logging
		logger, _ := logging.InitZap(viper.GetString("level"))
		defer func(logger *zap.Logger) {
			_ = logger.Sync()
		}(logger)
		stdLog := zap.RedirectStdLog(logger)
		defer stdLog()

		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load HTTP server config
		var srvCfg commonApi.Config
		if err := viper.Unmarshal(&srvCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "domain-service")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()
		database.ConnectDb(logger, true)

		store := store.NewDomainStore(database.DB.Db)

		stopCh := signals.SetupSignalHandler()

		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, store)
			go grpcSrv.ListenAndServe(stopCh)
		}

		// start HTTP server
		srv := api.NewServer(logger, &srvCfg, store)
		srv.ListenAndServe(stopCh)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("host", "", "Host to bind service to")
	runCmd.Flags().Int("port", 8080, "HTTP port to bind service to")
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("jwks_uri", "", "The location of the jwk set")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
