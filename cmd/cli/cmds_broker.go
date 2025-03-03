package cli

import (
	"github.com/faelmori/gkbxsrv/internal/services"
	databases "github.com/faelmori/gkbxsrv/services"
	"github.com/spf13/cobra"
)

func BrokerCommands() []*cobra.Command {
	return []*cobra.Command{
		gdbaseCommand(),
	}
}

func brokerCommand() *cobra.Command {
	var configFile, host, port string

	if defaultConfitFile == "" {
		if fs == nil {
			tfs := databases.NewFileSystemService("")
			fs = *tfs
		}
		defaultConfitFile = fs.GetConfigFilePath()
	}

	var brokerExp = []string{
		"gkbxsrv broker start --config='config.json'",
		"gkbxsrv broker stop",
	}

	cmd := &cobra.Command{
		Use:     "broker",
		Aliases: []string{"brkr", "bkr"},
		Example: concatenateExamples(brokerExp),
		Annotations: getDescriptions([]string{
			"Broker server and manager for many things",
			"Broker for interacting with the database, models, and many other services",
		}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := services.NewConfigSrv(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
			brk := services.NewBroker(cfg)

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", defaultConfitFile, "config file")
	cmd.Flags().StringVarP(&host, "host", "H", "localhost", "host")
	cmd.Flags().StringVarP(&port, "port", "P", "5432", "port")

	return cmd
}
