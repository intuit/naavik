package main

import (
	"os"

	"github.com/intuit/naavik/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
