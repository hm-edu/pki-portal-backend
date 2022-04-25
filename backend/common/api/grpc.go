package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hm-edu/portal-common/logging"
	"github.com/joho/godotenv"
	"github.com/ory/viper"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/xds"
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
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	if strings.HasPrefix(host, "xds") {
		creds, err := xds.NewClientCredentials(xds.ClientOptions{
			FallbackCreds: insecure.NewCredentials(),
		})
		if err != nil {
			return nil, err
		}
		conn, err := grpc.DialContext(
			ctx,
			host,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	conn, err := grpc.DialContext(ctx, host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return conn, nil

}
