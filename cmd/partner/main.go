// Package main is the entry point for the AI Assistant CLI
package main

import (
	"fmt"
	"os"

	"partner/internal/cli"
)

func main() {
	app := cli.NewApp()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
