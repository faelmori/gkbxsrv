package main

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/cmd"
	"github.com/faelmori/gkbxsrv/services"
	"os"
)

func main() {
	if os.Args[1] == "server" {
		go func() {
			f := services.NewFileSystemService("")
			fs := *f
			c := services.NewConfigService(fs.GetConfigFilePath(), fs.GetDefaultKeyPath(), fs.GetDefaultCertPath())
			_ = services.NewBrokerService(c)
		}()
	}

	if utlErr := cmd.RegX().Execute(); utlErr != nil {
		fmt.Println(fmt.Errorf("error: %v", utlErr))
		os.Exit(1)
	}

	os.Exit(0)
}
