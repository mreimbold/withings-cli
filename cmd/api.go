package cmd

import "github.com/spf13/cobra"

type apiCallOptions struct {
	Service string
	Action  string
	Params  string
	DryRun  bool
}

var apiCallOpts apiCallOptions

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Low-level API access",
}

var apiCallCmd = &cobra.Command{
	Use:   "call",
	Short: "Call a Withings API service/action",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.AddCommand(apiCallCmd)

	apiCallCmd.Flags().StringVar(&apiCallOpts.Service, "service", "", "API service name")
	apiCallCmd.Flags().StringVar(&apiCallOpts.Action, "action", "", "API action name")
	apiCallCmd.Flags().StringVar(&apiCallOpts.Params, "params", "", "JSON params, @file.json, or - for stdin")
	apiCallCmd.Flags().BoolVar(&apiCallOpts.DryRun, "dry-run", false, "print request without executing")

	_ = apiCallCmd.MarkFlagRequired("service")
	_ = apiCallCmd.MarkFlagRequired("action")
}
