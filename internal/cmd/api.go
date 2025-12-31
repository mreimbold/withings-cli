package cmd

import "github.com/spf13/cobra"

type apiCallOptions struct {
	Service string
	Action  string
	Params  string
	DryRun  bool
}

func newAPICommand() *cobra.Command {
	var opts apiCallOptions

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
			return runAPICall(cmd, opts)
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
