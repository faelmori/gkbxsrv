package cli

import (
	"github.com/faelmori/gkbxsrv/internal/services"
	l "github.com/faelmori/logz"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func BrokerCommands() []*cobra.Command {
	return []*cobra.Command{
		brokerCommand(),
	}
}

func brokerCommand() *cobra.Command {
	if defaultConfitFile == "" {
		//if fs == nil {
		//	tfs := kbxApi.NewFileSystemService("")
		//	fs = *tfs
		//}
		//defaultConfitFile = databases.NewFileSystemService(defaultConfitFile)
	}

	var brokerExp = []string{
		"gkbxsrv broker start --config='config.json'",
		"gkbxsrv broker stop",
	}

	var ws sync.WaitGroup

	cmd := &cobra.Command{
		Use:     "broker",
		Aliases: []string{"brkr", "bkr"},
		Example: concatenateExamples(brokerExp),
		Annotations: getDescriptions([]string{
			"Broker server and manager for many things",
			"Broker for interacting with the database, models, and many other services",
		}, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			chanSig := make(chan os.Signal, 1)
			signal.Notify(chanSig, syscall.SIGINT, syscall.SIGTERM)

			ws.Add(1)
			go func() {
				defer ws.Done()
				l.GetLogger("GKBXSrv").Info("Starting broker...", map[string]interface{}{"configFile": configFile, "host": host, "port": port})

				///if _, brkErr := services.NewBrokerConn(port); brkErr != nil {
				if _, brkErr := services.NewBroker(true); brkErr != nil {
					l.GetLogger("GKBXSrv").Fatalln("Error starting broker", map[string]interface{}{
						"context":  "gkbxsrv",
						"action":   "broker",
						"showData": true,
						"error":    brkErr.Error(),
						"port":     port,
					})

					chanSig <- syscall.SIGTERM
					return
				}
			}()
			l.GetLogger("GKBXSrv").Info("Broker started successfully!", nil)

			<-chanSig
			ws.Wait()
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", defaultConfitFile, "config file")
	cmd.Flags().StringVarP(&host, "host", "H", "", "host")
	cmd.Flags().StringVarP(&port, "port", "P", "5555", "port")

	return cmd
}
