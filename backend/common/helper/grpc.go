package helper

import (
	"fmt"
	"net"
	"os"

	"github.com/hm-edu/portal-common/logging"
	"github.com/joho/godotenv"
	"github.com/ory/viper"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServerWrapper is the basic structure of the GRPC server.
type ServerWrapper interface {
	grpc.ServiceRegistrar
	reflection.ServiceInfoProvider

	Serve(net.Listener) error
	GracefulStop()
}

// PrepareEnv loads the basic data and e.g. configures the logger.
func PrepareEnv(cmd *cobra.Command) (*zap.Logger, func(*zap.Logger)) {
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
	}
}
