package cmd

import (
	"github.com/hm-edu/dns-service/pkg/core"
	"github.com/hm-edu/dns-service/pkg/grpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var mwnCmd = &cobra.Command{
	Use:   "mwn",
	Short: "Uses the mwn provider",
	Run: func(cmd *cobra.Command, args []string) {
		grpcCfg, logger, deferFunc := prepare(cmd, args)
		defer deferFunc()
		cfg := &core.MwnProviderConfig{}
		if err := viper.Unmarshal(&cfg); err != nil {
			logger.Panic("config unmarshal failed", zap.Error(err))
		}
		provider := core.NewMwnProvider(logger, cfg)
		// start gRPC server
		if grpcCfg.Port > 0 {
			grpcSrv, _ := grpc.NewServer(&grpcCfg, logger, provider)
			grpcSrv.ListenAndServe(stopCh)
		}
	},
}

func init() {
	runCmd.AddCommand(mwnCmd)
	registerFlags(mwnCmd)
	mwnCmd.Flags().String("write_nameserver", "", "Sets the nameserver")
	mwnCmd.Flags().String("read_nameserver", "", "Sets the nameserver")
	mwnCmd.Flags().String("tsig_key_name", "", "Sets the tsig Key Name")
	mwnCmd.Flags().String("tsig_secret", "", "Sets the tsig Secret")
	mwnCmd.Flags().String("tsig_secret_alg", "", "Sets the tsig Secret Alg")
}
