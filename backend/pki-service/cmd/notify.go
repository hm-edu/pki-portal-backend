package cmd

import (
	"github.com/hm-edu/pki-service/pkg/database"
	"github.com/hm-edu/pki-service/pkg/grpc"
	"github.com/hm-edu/pki-service/pkg/worker"
	"github.com/hm-edu/portal-common/api"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// notifyCmd represents the run command
var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Runs the notifications",
	Long:  `Starts the GRPC server`,
	Run: func(cmd *cobra.Command, _ []string) {
		logger, deferFunc, viper := api.PrepareEnv(cmd)
		defer deferFunc(logger)
		var grpcCfg grpc.Config
		if err := viper.Unmarshal(&grpcCfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}

		database.ConnectDb(logger, viper.GetString("db"))
		w := worker.Notifier{
			Db:        database.DB.Db,
			MailHost:  viper.GetString("mail_host"),
			MailPort:  viper.GetInt("mail_port"),
			MailFrom:  viper.GetString("mail_from"),
			MailTo:    viper.GetString("mail_to"),
			MailToBcc: viper.GetString("mail_bcc"),
			Force:     viper.GetBool("force"),
		}

		if err := w.Notify(logger); err != nil {
			logger.Error("Error while sending notifications", zap.Error(err))
		}

	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.Flags().String("db", "", "connection string for the database")
	notifyCmd.Flags().String("mail_host", "", "The mail host")
	notifyCmd.Flags().Int("mail_port", 25, "The mail port")
	notifyCmd.Flags().String("mail_from", "", "The mail from")
	notifyCmd.Flags().String("mail_to", "", "Optional param to send notifications to a specific mail address instead of the orignal issuer.")
	notifyCmd.Flags().String("mail_bcc", "", "Optional param to send notifications as blind copy to a specific mail address instead of the orignal issuer.")
	notifyCmd.Flags().Bool("force", false, "Optional param to force sending notifications.")
}
