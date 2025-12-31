package cli

import (
	"fmt"

	"github.com/mreimbold/withings-cli/internal/auth"
	"github.com/mreimbold/withings-cli/internal/services/api"
	"github.com/spf13/cobra"
)

func newAPICommand() *cobra.Command {
	var opts api.Options

	//nolint:exhaustruct // Cobra command defaults are intentional.
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level API access",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	apiCallCmd := &cobra.Command{
		Use:   "call",
		Short: "Call a Withings API service/action",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			accessToken, err := auth.EnsureAccessToken(cmd.Context(), appOpts)
			if err != nil {
				return fmt.Errorf("ensure access token: %w", err)
			}

			return api.Run(cmd.Context(), opts, appOpts, accessToken)
		},
	}

	apiCmd.AddCommand(apiCallCmd)

	apiCallCmd.Flags().StringVar(
		&opts.Service,
		"service",
		emptyString,
		"API service name",
	)
	apiCallCmd.Flags().StringVar(
		&opts.Action,
		"action",
		emptyString,
		"API action name",
	)
	apiCallCmd.Flags().StringVar(
		&opts.Params,
		"params",
		emptyString,
		"JSON params, @file.json, or - for stdin",
	)
	apiCallCmd.Flags().BoolVar(
		&opts.DryRun,
		"dry-run",
		false,
		"print request without executing",
	)

	_ = apiCallCmd.MarkFlagRequired("service")
	_ = apiCallCmd.MarkFlagRequired("action")

	return apiCmd
}
