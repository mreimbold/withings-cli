package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/spf13/cobra"
)

var version = "dev"

// Execute runs the CLI and returns the exit code.
func Execute() int {
	rootCmd := newRootCommand()

	err := rootCmd.Execute()
	if err == nil {
		return app.ExitCodeSuccess
	}

	code := app.ExitCodeFailure

	var exitErr *app.ExitError

	if errors.As(err, &exitErr) {
		code = exitErr.Code
		err = exitErr.Err
	}

	_, writeErr := fmt.Fprintln(os.Stderr, err)
	if writeErr != nil {
		return app.ExitCodeFailure
	}

	return code
}

func newRootCommand() *cobra.Command {
	var opts app.Options

	rootCmd := buildRootCommand(&opts)
	rootCmd.Version = version

	addRootCommands(rootCmd)
	addRootFlags(rootCmd, &opts)

	return rootCmd
}

func buildRootCommand(opts *app.Options) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use: "withings",
		Short: "Interact with Withings Health Solutions " +
			"data and OAuth tokens from Withings CLI.",
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

func validateGlobalOptions(opts *app.Options) error {
	if opts.JSON && opts.Plain {
		return app.NewExitError(app.ExitCodeUsage, errJSONPlainConflict)
	}

	if opts.Quiet && opts.Verbose > noVerbosity {
		return app.NewExitError(app.ExitCodeUsage, errQuietVerboseConflict)
	}

	if opts.Plain {
		opts.NoColor = true
	}

	switch opts.Cloud {
	case "eu", "us":
		return nil
	default:
		return app.NewExitError(
			app.ExitCodeUsage,
			fmt.Errorf("%w: %q", errInvalidCloud, opts.Cloud),
		)
	}
}

func addRootCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newActivityCommand())
	rootCmd.AddCommand(newAPICommand())
	rootCmd.AddCommand(newAuthCommand())
	rootCmd.AddCommand(newHeartCommand())
	rootCmd.AddCommand(newMeasuresCommand())
	rootCmd.AddCommand(newSleepCommand())
}

func addRootFlags(rootCmd *cobra.Command, opts *app.Options) {
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
