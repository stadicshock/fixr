package main

import (
	"os"

	"github.com/fixr-app/fixr/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
