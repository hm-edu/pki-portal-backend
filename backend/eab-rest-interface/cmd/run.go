package cmd

import (
	"context"

	"github.com/hm-edu/eab-rest-interface/pkg/api"
	"github.com/hm-edu/eab-rest-interface/pkg/database"
	"github.com/hm-edu/eab-rest-interface/pkg/grpc"
	commonApi "github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the servers",
	Long:  `Starts the HTTP server`,
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

		database.ConnectDb(logger, viper.GetString("db"), viper.GetString("acme_db"))

		tp := tracing.InitTracer(logger, "eab-rest-interface")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		stopCh := signals.SetupSignalHandler()
		prov := viper.GetString("provisioner_id")
		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, prov)
			go grpcSrv.ListenAndServe(stopCh)
		}

		// start HTTP server
		srv := api.NewServer(logger, &srvCfg, prov)
		srv.ListenAndServe(stopCh)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("host", "", "Host to bind service to")
	runCmd.Flags().Int("port", 8080, "HTTP port to bind service to")
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("jwks_uri", "", "The location of the jwk set")
	runCmd.Flags().String("audience", "", "The expected audience")
	runCmd.Flags().StringSlice("cors_allowed_origins", []string{}, "The allowed origin for CORS")
	runCmd.Flags().String("db", "", "connection string for the database")
	runCmd.Flags().String("acme_db", "", "connection string for the acme database")
	runCmd.Flags().String("preseed", "", "path to the preseed file")
	runCmd.Flags().String("provisioner_id", "", "id of the smallstep provisioner to configure")
	runCmd.Flags().String("domain_service", "", "The domain service to use")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
