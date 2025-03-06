package main

import (
	"fmt"
	"os"
)

func main() {
	if utlErr := RegX().Execute(); utlErr != nil {
		fmt.Println(fmt.Errorf("error: %v", utlErr))
		os.Exit(1)
	}
}
