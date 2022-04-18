package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hm-edu/pki-rest-interface/pkg/api"
	"github.com/hm-edu/pki-rest-interface/pkg/cfg"
	"github.com/hm-edu/pki-rest-interface/pkg/grpc"
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

		// load HTTP server config
		var srvCfg commonApi.Config
		if err := viper.Unmarshal(&srvCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load handler config
		var handlerCfg cfg.HandlerConfiguration
		if err := viper.Unmarshal(&srvCfg); err != nil {
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
	runCmd.Flags().String("smime_service", "", "The smime service to use")
	runCmd.Flags().String("ssl_service", "", "The ssl service to use")
	runCmd.Flags().String("domain_service", "", "The domain service to use")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
