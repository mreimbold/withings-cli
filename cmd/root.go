package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
	exitCodeUsage   = 2
	exitCodeAuth    = 3
	exitCodeNetwork = 4
	exitCodeAPI     = 5
)

type globalOptions struct {
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

type exitError struct {
	code int
	err  error
}

func newExitError(code int, err error) *exitError {
	return &exitError{code: code, err: err}
}

// Error returns the wrapped error message.
func (e *exitError) Error() string {
	return e.err.Error()
}

type runEFunc func(*cobra.Command, []string) error

// Execute runs the CLI and returns the exit code.
func Execute() int {
	rootCmd := newRootCommand()

	err := rootCmd.Execute()
	if err == nil {
		return exitCodeSuccess
	}

	code := exitCodeFailure

	var exitErr *exitError

	if errors.As(err, &exitErr) {
		code = exitErr.code
		err = exitErr.err
	}

	_, writeErr := fmt.Fprintln(os.Stderr, err)
	if writeErr != nil {
		return exitCodeFailure
	}

	return code
}

func newRootCommand() *cobra.Command {
	var opts globalOptions

	rootCmd := buildRootCommand(&opts)
	rootCmd.Version = version

	addRootCommands(rootCmd)
	addRootFlags(rootCmd, &opts)

	return rootCmd
}

func buildRootCommand(opts *globalOptions) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use: "withings",
		Short: "Interact with Withings Health Solutions " +
			"data and OAuth tokens from the CLI.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return validateGlobalOptions(opts)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
}

func validateGlobalOptions(opts *globalOptions) error {
	if opts.JSON && opts.Plain {
		return newExitError(exitCodeUsage, errJSONPlainConflict)
	}

	if opts.Quiet && opts.Verbose > noVerbosity {
		return newExitError(exitCodeUsage, errQuietVerboseConflict)
	}

	if opts.Plain {
		opts.NoColor = true
	}

	switch opts.Cloud {
	case "eu", "us":
		return nil
	default:
		return newExitError(
			exitCodeUsage,
			fmt.Errorf("%w: %q", errInvalidCloud, opts.Cloud),
		)
	}
}

func addRootCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newActivityCommand(notImplementedHandler))
	rootCmd.AddCommand(newAPICommand())
	rootCmd.AddCommand(newAuthCommand())
	rootCmd.AddCommand(newHeartCommand())
	rootCmd.AddCommand(newMeasuresCommand())
	rootCmd.AddCommand(newSleepCommand(notImplementedHandler))
	rootCmd.AddCommand(newUserCommand(notImplementedHandler))
}

func addRootFlags(rootCmd *cobra.Command, opts *globalOptions) {
	rootCmd.PersistentFlags().CountVarP(
		&opts.Verbose,
		"verbose",
		"v",
		"increase diagnostic verbosity (repeatable)",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&opts.Quiet,
		"quiet",
		"q",
		false,
		"suppress non-error output",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.JSON,
		"json",
		false,
		"machine-readable JSON output",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.Plain,
		"plain",
		false,
		"stable line-based output (no tables, no colors)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.NoColor,
		"no-color",
		false,
		"disable ANSI color",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.NoInput,
		"no-input",
		false,
		"disable prompts",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.Config,
		"config",
		emptyString,
		"config file path (optional)",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.Cloud,
		"cloud",
		defaultCloud,
		"API cloud: eu or us",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.BaseURL,
		"base-url",
		emptyString,
		"override API base URL",
	)
}

func notImplementedHandler(cmd *cobra.Command, _ []string) error {
	return newExitError(
		exitCodeFailure,
		fmt.Errorf("%w: %s", errNotImplemented, cmd.CommandPath()),
	)
}
