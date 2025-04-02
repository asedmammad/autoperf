package main

import (
	"os"

	"github.com/asedmammad/autoperf/cmd/autoperf"
)

func main() {
	if err := autoperf.Execute(); err != nil {
		os.Exit(1)
	}
}
