package main

import (
	"github.com/faelmori/gokubexfs/cmd"
)

func main() {
	utl := cmd.RegX()
	if utlErr := utl.Execute(); utlErr != nil {
		panic(utlErr)
	}
}
