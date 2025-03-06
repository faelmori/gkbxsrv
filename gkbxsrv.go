package main

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/cmd"
	"os"
)

func main() {
	if utlErr := cmd.RegX().Execute(); utlErr != nil {
		fmt.Println(fmt.Errorf("error: %v", utlErr))
		os.Exit(1)
	}
}
