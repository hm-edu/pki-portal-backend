package cmd

import (
	"context"

	"github.com/hm-edu/dns-service/pkg/grpc"
	"github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var stopCh = signals.SetupSignalHandler()

func prepare(cmd *cobra.Command, _ []string) (grpc.Config, *zap.Logger, func()) {
	logger, deferLoggerFunc, viper := api.PrepareEnv(cmd)

	var grpcCfg grpc.Config
	if err := viper.Unmarshal(&grpcCfg); err != nil {
		logger.Panic("config unmarshal failed", zap.Error(err))
	}

	server := tracing.InitTracer(logger)

	deferFunc := func() {
		deferLoggerFunc(logger)
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
		}
	}
	return grpcCfg, logger, deferFunc
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the servers",
	Long:  `Starts the GRPC server`,
}

func registerFlags(runCmd *cobra.Command) {
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}

func init() {
	rootCmd.AddCommand(runCmd)
}
