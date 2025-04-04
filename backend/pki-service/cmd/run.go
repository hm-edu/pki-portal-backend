package cmd

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/pkg/cfg"
	"github.com/hm-edu/pki-service/pkg/database"
	"github.com/hm-edu/pki-service/pkg/grpc"
	"github.com/hm-edu/pki-service/pkg/worker"
	"github.com/hm-edu/portal-common/api"
	"github.com/hm-edu/portal-common/signals"
	"github.com/hm-edu/portal-common/tracing"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the servers",
	Long:  `Starts the GRPC server`,
	Run: func(cmd *cobra.Command, _ []string) {
		logger, deferFunc, viper := api.PrepareEnv(cmd)
		defer deferFunc(logger)
		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load Sectigo config
		var pkiCfg cfg.PKIConfiguration
		if err := viper.Unmarshal(&pkiCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger)

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		stopCh := signals.SetupSignalHandler()

		database.ConnectDb(logger, viper.GetString("db"))

		_, errUpdate := database.DB.Db.Certificate.Update().Where(certificate.CaIsNil()).SetCa("sectigo").Save(context.Background())
		if errUpdate != nil {
			logger.Fatal("Error updating certificates", zap.Error(errUpdate))
		}

		s := gocron.NewScheduler(time.UTC)
		if viper.GetBool("enable_notifications") {

			w := worker.Notifier{Db: database.DB.Db,
				MailHost:  viper.GetString("mail_host"),
				MailPort:  viper.GetInt("mail_port"),
				MailFrom:  viper.GetString("mail_from"),
				MailTo:    viper.GetString("mail_to"),
				MailToBcc: viper.GetString("mail_bcc"),
			}

			_, err := s.Every(1).Day().At("09:00").Do(func() {
				if err := w.Notify(logger); err != nil {
					logger.Error("Error while sending notifications", zap.Error(err))
				}
			})
			if err != nil {
				logger.Error("Error while scheduling notifications", zap.Error(err))
			}
		}
		_, err := s.Every(1).Day().At("01:00").Do(func() {
			if err := worker.Cleanup(logger, database.DB.Db); err != nil {
				logger.Error("Error while cleaning up", zap.Error(err))
			}
		})
		if err != nil {
			logger.Error("Error while scheduling cleanup", zap.Error(err))
		}
		s.StartAsync()
		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, &pkiCfg, database.DB.Db)
			grpcSrv.ListenAndServe(stopCh)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("host", "", "Host to bind service to")
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("sentry_dsn", "", "The sentry dsn to use")
	runCmd.Flags().Int("smime_key_length", 365, "The (expected) key length")
	runCmd.Flags().String("db", "", "connection string for the database")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
	runCmd.Flags().Bool("enable_notifications", false, "Enable notifications")
	runCmd.Flags().String("mail_host", "", "The mail host")
	runCmd.Flags().Int("mail_port", 25, "The mail port")
	runCmd.Flags().String("mail_to", "", "Optional param to send notifications to a specific mail address instead of the orignal issuer.")
	runCmd.Flags().String("mail_bcc", "", "Optional param to send notifications as blind copy to a specific mail address instead of the orignal issuer.")
	runCmd.Flags().String("mail_from", "", "The mail from")
	runCmd.Flags().String("user", "", "The user for the HARICA API")
	runCmd.Flags().String("password", "", "The password for the HARICA API")
	runCmd.Flags().String("totp_seed", "", "The totp seed for the HARICA API")
	runCmd.Flags().String("validation_user", "", "The user for the HARICA API")
	runCmd.Flags().String("validation_password", "", "The password for the HARICA API")
	runCmd.Flags().String("validation_totp_seed", "", "The totp seed for the HARICA API")
	runCmd.Flags().String("cert_type", "OV", "The certificate type to use")
}
