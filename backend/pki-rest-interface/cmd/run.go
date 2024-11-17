package cmd

import (
	"context"

	"github.com/hm-edu/pki-rest-interface/pkg/api"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	"github.com/hm-edu/pki-rest-interface/pkg/grpc"
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
	Long:  `Starts the HTTP and the GRPC server`,
	Run: func(cmd *cobra.Command, _ []string) {

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

		// load handler config
		var handlerCfg cfg.HandlerConfiguration
		if err := viper.Unmarshal(&handlerCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}
		tp := tracing.InitTracer(logger, "pki-rest-interface")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		stopCh := signals.SetupSignalHandler()

		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger)
			go grpcSrv.ListenAndServe(stopCh)
		}

		// start HTTP server
		srv := api.NewServer(logger, &srvCfg, &handlerCfg)
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
	runCmd.Flags().String("sentry_dsn", "", "sentry dsn")
	runCmd.Flags().StringSlice("cors_allowed_origins", []string{}, "The allowed origin for CORS")
	runCmd.Flags().String("smime_service", "", "The smime service to use")
	runCmd.Flags().String("ssl_service", "", "The ssl service to use")
	runCmd.Flags().String("domain_service", "", "The domain service to use")
	runCmd.Flags().Bool("reject_students", false, "Reject students")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
