package cmd

import (
	"context"
	"net/http"

	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/hm-edu/sectigo-client/sectigo"
	"github.com/hm-edu/validation-service/pkg/cfg"
	"github.com/hm-edu/validation-service/pkg/worker"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Syncs the database with the Sectigo API",
	Long:  `Adds all missing entries from the Sectigo API to the database`,
	Run: func(cmd *cobra.Command, _ []string) {

		logger, deferFunc, viper := api.PrepareEnv(cmd)
		defer deferFunc(logger)

		// load HTTP server config
		var sectigoCfg cfg.SectigoConfiguration
		if err := viper.Unmarshal(&sectigoCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "validation-service")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		conn, err := api.ConnectGRPC(viper.GetString("dns_service"))
		if err != nil {
			logger.Fatal("Could not connect to gRPC server", zap.Error(err))
		}

		validator := worker.DomainValidator{
			Client:     sectigo.NewClient(http.DefaultClient, logger, sectigoCfg.User, sectigoCfg.Password, sectigoCfg.CustomerURI),
			DNSService: pb.NewDNSServiceClient(conn),
			Domains:    viper.GetStringSlice("domains"),
			Force:      viper.GetBool("force"),
		}

		stopCh := signals.SetupSignalHandler()
		validator.ValidateDomains()
		<-stopCh
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().String("sectigo_user", "", "The sectigo user")
	validateCmd.Flags().String("sectigo_password", "", "The password for the sectigo user")
	validateCmd.Flags().String("sectigo_customeruri", "", "The sectigo customerUri")
	validateCmd.Flags().String("dns_service", "", "The dns service to use")
	validateCmd.Flags().Bool("force", false, "Force the validation of all domains")
	validateCmd.Flags().StringSlice("domains", nil, "The domains to validate")
	validateCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
}
