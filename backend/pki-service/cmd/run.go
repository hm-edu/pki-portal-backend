package cmd

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
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
	Run: func(cmd *cobra.Command, args []string) {
		logger, deferFunc, viper := api.PrepareEnv(cmd)
		defer deferFunc(logger)
		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		// load Sectigo config
		var sectigoCfg cfg.SectigoConfiguration
		if err := viper.Unmarshal(&sectigoCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		tp := tracing.InitTracer(logger, "pki-service")

		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				logger.Fatal("Error shutting down tracer provider.", zap.Error(err))
			}
		}()

		stopCh := signals.SetupSignalHandler()

		sectigoCfg.CheckSectigoConfiguration()

		database.ConnectDb(logger, viper.GetString("db"))
		w := worker.Notifier{Db: database.DB.Db,
			MailHost:  viper.GetString("mail_host"),
			MailPort:  viper.GetInt("mail_port"),
			MailFrom:  viper.GetString("mail_from"),
			MailTo:    viper.GetString("mail_to"),
			MailToBcc: viper.GetString("mail_bcc"),
		}

		s := gocron.NewScheduler(time.UTC)
		s.Every(1).Day().At("09:00").Do(func() {
			w.Notify(logger)
		})
		s.StartAsync()

		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, &sectigoCfg, database.DB.Db)
			grpcSrv.ListenAndServe(stopCh)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().String("host", "", "Host to bind service to")
	runCmd.Flags().Int("grpc-port", 8081, "GRPC port to bind service to")
	runCmd.Flags().String("sectigo_user", "", "The sectigo user")
	runCmd.Flags().String("sectigo_password", "", "The password for the sectigo user")
	runCmd.Flags().String("sectigo_customeruri", "", "The sectigo customerUri")
	runCmd.Flags().Int("smime_profile", 0, "The (default) smime profile id")
	runCmd.Flags().Int("smime_org_id", 0, "The (default) org id")
	runCmd.Flags().Int("smime_term", 0, "The (default) lifetime of an employee certificate")
	runCmd.Flags().Int("smime_student_term", 0, "The (default) lifetime of a student certificate")
	runCmd.Flags().Int("smime_key_length", 365, "The (expected) key length")
	runCmd.Flags().String("smime_key_type", "", "The (expected) key type")
	runCmd.Flags().String("db", "", "connection string for the database")
	runCmd.Flags().Int("ssl_profile", 0, "The (default) ssl profile id")
	runCmd.Flags().Int("ssl_org_id", 0, "The (default) ssl org id")
	runCmd.Flags().Int("ssl_term", 0, "The (default) ssl lifetime")
	runCmd.Flags().String("level", "info", "log level debug, info, warn, error, flat or panic")
	runCmd.Flags().Bool("enable_notifications", false, "Enable notifications")
	runCmd.Flags().String("mail_host", "", "The mail host")
	runCmd.Flags().Int("mail_port", 25, "The mail port")
	runCmd.Flags().String("mail_to", "", "Optional param to send notifications to a specific mail address instead of the orignal issuer.")
	runCmd.Flags().String("mail_bcc", "", "Optional param to send notifications as blind copy to a specific mail address instead of the orignal issuer.")
	runCmd.Flags().String("mail_from", "", "The mail from")
}
