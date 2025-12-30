package cmd

import "github.com/spf13/cobra"

type authLoginOptions struct {
	RedirectURI string
	NoOpen      bool
	Listen      string
}

type authAuthorizeURLOptions struct {
	RedirectURI string
	Scope       string
}

type authExchangeOptions struct {
	Code string
}

type authLogoutOptions struct {
	Force bool
}

type authSetClientOptions struct {
	ClientID    string
	SecretStdin bool
}

func newAuthCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OAuth and client credentials",
	}

	authCmd.AddCommand(newAuthLoginCommand())
	authCmd.AddCommand(newAuthAuthorizeURLCommand())
	authCmd.AddCommand(newAuthExchangeCommand())
	authCmd.AddCommand(newAuthRefreshCommand())
	authCmd.AddCommand(newAuthStatusCommand())
	authCmd.AddCommand(newAuthLogoutCommand())
	authCmd.AddCommand(newAuthSetClientCommand())

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

func newAuthAuthorizeURLCommand() *cobra.Command {
	var opts authAuthorizeURLOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "authorize-url",
		Short: "Print the OAuth authorize URL",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthAuthorizeURL(cmd, opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.RedirectURI,
		"redirect-uri",
		emptyString,
		"override redirect URI",
	)
	cmd.Flags().StringVar(
		&opts.Scope,
		"scope",
		emptyString,
		"override OAuth scopes (comma-separated)",
	)

	return cmd
}

func newAuthExchangeCommand() *cobra.Command {
	var opts authExchangeOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "exchange",
		Short: "Exchange an authorization code for tokens",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthExchange(cmd, opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.Code,
		"code",
		emptyString,
		"authorization code (otherwise read from stdin)",
	)

	return cmd
}

func newAuthRefreshCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use:   "refresh",
		Short: "Refresh the access token",
		RunE:  runAuthRefresh,
	}
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

func newAuthSetClientCommand() *cobra.Command {
	var opts authSetClientOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "set-client",
		Short: "Set OAuth client ID and secret",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAuthSetClient(cmd, opts)
		},
	}

	cmd.Flags().StringVar(
		&opts.ClientID,
		"client-id",
		emptyString,
		"OAuth client ID",
	)
	cmd.Flags().BoolVar(
		&opts.SecretStdin,
		"secret-stdin",
		false,
		"read client secret from stdin",
	)

	return cmd
}
