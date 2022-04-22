package cmd

import (
	"github.com/hm-edu/portal-common/helper"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the servers",
	Long:  `Starts the GRPC server`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, deferFunc := helper.PrepareEnv(cmd)
		defer deferFunc(logger)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
