package cmd

import (
	"context"
	"strings"

	"github.com/hm-edu/domain-rest-interface/pkg/api"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/hm-edu/domain-rest-interface/pkg/grpc"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	commonApi "github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	grpc_sentry "github.com/johnbellone/grpc-middleware-sentry"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
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

		tp := tracing.InitTracer(logger)

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		database.ConnectDb(logger, viper.GetString("db"))

		store := store.NewDomainStore(database.DB.Db)
		store.Preseed(logger)
		stopCh := signals.SetupSignalHandler()

		adminsVar := viper.GetStringSlice("admins")
		admins := []string{}
		for _, admin := range adminsVar {
			if strings.Contains(admin, ",") {
				admins = append(admins, strings.Split(admin, ",")...)
			} else {
				admins = append(admins, admin)
			}
		}

		logger.Info("Starting server with admins", zap.Strings("admins", admins))

		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, store, admins)
			go grpcSrv.ListenAndServe(stopCh)
		}

		client, err := sslClient(viper.GetString("ssl_service"), grpcCfg.SentryDSN)
		if err != nil {
			logger.Fatal("Error connecting to ssl service", zap.Error(err))
		}

		// start HTTP server
		srv := api.NewServer(logger, &srvCfg, store, client, admins)
		srv.ListenAndServe(stopCh)
	},
}

func sslClient(host string, sentryDSN string) (pb.SSLServiceClient, error) {
	var interceptor []googleGrpc.UnaryClientInterceptor
	if sentryDSN != "" {
		interceptor = append(interceptor, grpc_sentry.UnaryClientInterceptor())
	}
	conn, err := commonApi.ConnectGRPC(host, googleGrpc.WithChainUnaryInterceptor(interceptor...))
	if err != nil {
		return nil, err
	}
	return pb.NewSSLServiceClient(conn), nil
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
	runCmd.Flags().String("sentry_dsn", "", "sentry dsn")
	runCmd.Flags().String("ssl_service", "", "pki backend")
	runCmd.Flags().String("preseed", "", "path to the preseed file")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
	runCmd.Flags().StringSlice("admins", []string{}, "list of admin emails")
}
