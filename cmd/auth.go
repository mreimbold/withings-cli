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

func newAuthCommand(notImplemented runEFunc) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage OAuth and client credentials",
	}

	authCmd.AddCommand(newAuthLoginCommand(notImplemented))
	authCmd.AddCommand(newAuthAuthorizeURLCommand(notImplemented))
	authCmd.AddCommand(newAuthExchangeCommand(notImplemented))
	authCmd.AddCommand(newAuthRefreshCommand(notImplemented))
	authCmd.AddCommand(newAuthStatusCommand(notImplemented))
	authCmd.AddCommand(newAuthLogoutCommand(notImplemented))
	authCmd.AddCommand(newAuthSetClientCommand(notImplemented))

	return authCmd
}

func newAuthLoginCommand(notImplemented runEFunc) *cobra.Command {
	var opts authLoginOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Start browser OAuth flow and store tokens",
		RunE:  notImplemented,
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

func newAuthAuthorizeURLCommand(notImplemented runEFunc) *cobra.Command {
	var opts authAuthorizeURLOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "authorize-url",
		Short: "Print the OAuth authorize URL",
		RunE:  notImplemented,
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

func newAuthExchangeCommand(notImplemented runEFunc) *cobra.Command {
	var opts authExchangeOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "exchange",
		Short: "Exchange an authorization code for tokens",
		RunE:  notImplemented,
	}

	cmd.Flags().StringVar(
		&opts.Code,
		"code",
		emptyString,
		"authorization code (otherwise read from stdin)",
	)

	return cmd
}

func newAuthRefreshCommand(notImplemented runEFunc) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use:   "refresh",
		Short: "Refresh the access token",
		RunE:  notImplemented,
	}
}

func newAuthStatusCommand(notImplemented runEFunc) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	return &cobra.Command{
		Use:   "status",
		Short: "Show token scopes and expiry",
		RunE:  notImplemented,
	}
}

func newAuthLogoutCommand(notImplemented runEFunc) *cobra.Command {
	var opts authLogoutOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Delete stored tokens",
		RunE:  notImplemented,
	}

	cmd.Flags().BoolVar(
		&opts.Force,
		"force",
		false,
		"skip confirmation",
	)

	return cmd
}

func newAuthSetClientCommand(notImplemented runEFunc) *cobra.Command {
	var opts authSetClientOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "set-client",
		Short: "Set OAuth client ID and secret",
		RunE:  notImplemented,
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
