package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version can be overridden at build time with -ldflags.
var version = "dev"

type GlobalOptions struct {
	Verbose int
	Quiet   bool
	JSON    bool
	Plain   bool
	NoColor bool
	NoInput bool
	Config  string
	Cloud   string
	BaseURL string
}

var globalOpts GlobalOptions

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

var rootCmd = &cobra.Command{
	Use:           "withings",
	Short:         "Interact with Withings Health Solutions data and OAuth tokens from the CLI.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globalOpts.JSON && globalOpts.Plain {
			return NewExitError(2, fmt.Errorf("--json and --plain are mutually exclusive"))
		}
		if globalOpts.Quiet && globalOpts.Verbose > 0 {
			return NewExitError(2, fmt.Errorf("--quiet and --verbose cannot be combined"))
		}
		if globalOpts.Plain {
			globalOpts.NoColor = true
		}
		switch globalOpts.Cloud {
		case "eu", "us":
		default:
			return NewExitError(2, fmt.Errorf("invalid --cloud %q (expected eu or us)", globalOpts.Cloud))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		code := 1
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.Code
			err = exitErr.Err
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}

func init() {
	rootCmd.Version = version

	rootCmd.PersistentFlags().CountVarP(&globalOpts.Verbose, "verbose", "v", "increase diagnostic verbosity (repeatable)")
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Quiet, "quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.JSON, "json", false, "machine-readable JSON output")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.Plain, "plain", false, "stable line-based output (no tables, no colors)")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.NoColor, "no-color", false, "disable ANSI color")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.NoInput, "no-input", false, "disable prompts")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Config, "config", "", "config file path (optional)")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Cloud, "cloud", "eu", "API cloud: eu or us")
	rootCmd.PersistentFlags().StringVar(&globalOpts.BaseURL, "base-url", "", "override API base URL")
}

func notImplemented(cmd *cobra.Command, _ []string) error {
	return NewExitError(1, fmt.Errorf("not implemented: %s", cmd.CommandPath()))
}
