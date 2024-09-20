package api

import (
	"fmt"
	"os"

	"github.com/hm-edu/portal-common/logging"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PrepareEnv loads the basic data and e.g. configures the logger.
func PrepareEnv(cmd *cobra.Command) (*zap.Logger, func(*zap.Logger), *viper.Viper) {
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
	stdLog := zap.RedirectStdLog(logger)

	return logger, func(logger *zap.Logger) {
		_ = logger.Sync()
		stdLog()
	}, viper.GetViper()
}

// ConnectGRPC connects to the GRPC server.
func ConnectGRPC(host string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		return nil, err
	}
	return conn, nil

}
