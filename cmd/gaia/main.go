// Package main is the entry point for the Gaia CLI
package main

import (
	"fmt"
	"os"

	"gaia/internal/cli"
)

func main() {
	app := cli.NewApp()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
