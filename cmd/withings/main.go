// Package main is the entrypoint for the Withings CLI binary.
package main

import (
	"os"

	"github.com/mreimbold/withings-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
