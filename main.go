// Package main is the entrypoint for the withings-cli binary.
package main

import (
	"os"

	"github.com/mreimbold/withings-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
