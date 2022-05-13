package cmd

import (
	"context"

	"github.com/hm-edu/domain-rest-interface/pkg/api"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/hm-edu/domain-rest-interface/pkg/grpc"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
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

		tp := tracing.InitTracer(logger, "domain-rest-interface")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		database.ConnectDb(logger, viper.GetString("db"))

		store := store.NewDomainStore(database.DB.Db)
		store.Preseed(logger)
		stopCh := signals.SetupSignalHandler()

		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, store)
			go grpcSrv.ListenAndServe(stopCh)
		}

		client, err := sslClient(viper.GetString("ssl_service"))
		if err != nil {
			logger.Fatal("Error connecting to ssl service", zap.Error(err))
		}

		// start HTTP server
		srv := api.NewServer(logger, &srvCfg, store, client)
		srv.ListenAndServe(stopCh)
	},
}

func sslClient(host string) (pb.SSLServiceClient, error) {
	conn, err := commonApi.ConnectGRPC(host)
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
	runCmd.Flags().String("db", "", "connection string for the database")
	runCmd.Flags().String("ssl_service", "", "pki backend")
	runCmd.Flags().String("preseed", "", "path to the preseed file")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
