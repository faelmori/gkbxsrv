package main

import (
	"fmt"
	"github.com/faelmori/gkbxsrv/cmd"
	"os"
)

func main() {
	utl := cmd.RegX()
	if utlErr := utl.Execute(); utlErr != nil {
		fmt.Println(fmt.Errorf("error: %v", utlErr))
		os.Exit(1)
	}
	os.Exit(0)
}
