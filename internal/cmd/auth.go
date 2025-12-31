package cmd

import "github.com/spf13/cobra"

type authLoginOptions struct {
	RedirectURI string
	NoOpen      bool
	Listen      string
}

type authLogoutOptions struct {
	Force bool
}

func newAuthCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OAuth tokens",
	}

	authCmd.AddCommand(newAuthLoginCommand())
	authCmd.AddCommand(newAuthStatusCommand())
	authCmd.AddCommand(newAuthLogoutCommand())

	return authCmd
}

func newAuthLoginCommand() *cobra.Command {
	var opts authLoginOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Start browser OAuth flow and store tokens",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthLogin(cmd, opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.RedirectURI,
		"redirect-uri",
		emptyString,
		"override redirect URI",
	)
	cmd.Flags().BoolVar(
		&opts.NoOpen,
		"no-open",
		false,
		"print URL instead of opening a browser",
	)
	cmd.Flags().StringVar(
		&opts.Listen,
		"listen",
		defaultListenAddr,
		"callback listen address",
	)

	return cmd
}

func newAuthStatusCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use:   "status",
		Short: "Show token scopes and expiry",
		RunE:  runAuthStatus,
	}
}

func newAuthLogoutCommand() *cobra.Command {
	var opts authLogoutOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Delete stored tokens",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthLogout(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(
		&opts.Force,
		"force",
		false,
		"skip confirmation",
	)

	return cmd
}
