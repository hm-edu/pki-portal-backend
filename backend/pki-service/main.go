package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hm-edu/pki-service/pkg/api"
	"github.com/hm-edu/pki-service/pkg/grpc"
	commonApi "github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/logging"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {

	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	fs.String("host", "", "Host to bind service to")
	fs.Int("port", 8080, "HTTP port to bind service to")
	fs.Int("grpc-port", 8081, "GRPC port to bind service to")
	fs.String("oauth2_id", "", "The client id used for token introspection")
	fs.String("oauth2_secret", "", "The client secret used for token introspection")
	fs.String("oauth2_endpoint", "", "The url used for token introspection")
	fs.String("level", "info", "log level debug, info, warn, error, flat or panic")

	// parse flags
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		_, err := fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		if err != nil {
			os.Exit(2)
		}
		fs.PrintDefaults()
		os.Exit(2)
	}
	err = viper.BindPFlags(fs)
	if err != nil {
		os.Exit(1)
	}
	_ = godotenv.Load()
	viper.AutomaticEnv()

	// configure logging
	logger, _ := logging.InitZap(viper.GetString("level"))
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	if _, err := strconv.Atoi(viper.GetString("port")); err != nil {
		port, _ := fs.GetInt("port")
		viper.Set("port", strconv.Itoa(port))
	}

	var grpcCfg grpc.Config
	if err := viper.Unmarshal(&grpcCfg); err != nil {
		logger.Panic("config unmarshal failed", zap.Error(err))
	}

	// load HTTP server config
	var srvCfg commonApi.Config
	if err := viper.Unmarshal(&srvCfg); err != nil {
		logger.Panic("config unmarshal failed", zap.Error(err))
	}

	tp := tracing.InitTracer(logger, "pki-service")

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
		}
	}()

	// start gRPC server
	if grpcCfg.Port > 0 {
		grpcSrv, _ := grpc.NewServer(&grpcCfg, logger)
		go grpcSrv.ListenAndServe()
	}

	// start HTTP server
	srv := api.NewServer(logger, &srvCfg)
	stopCh := signals.SetupSignalHandler()
	srv.ListenAndServe(stopCh)

}
