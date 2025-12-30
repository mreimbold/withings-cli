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

var (
	authLoginOpts        authLoginOptions
	authAuthorizeURLOpts authAuthorizeURLOptions
	authExchangeOpts     authExchangeOptions
	authLogoutOpts       authLogoutOptions
	authSetClientOpts    authSetClientOptions
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage OAuth and client credentials",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Start browser OAuth flow and store tokens",
	RunE:  notImplemented,
}

var authAuthorizeURLCmd = &cobra.Command{
	Use:   "authorize-url",
	Short: "Print the OAuth authorize URL",
	RunE:  notImplemented,
}

var authExchangeCmd = &cobra.Command{
	Use:   "exchange",
	Short: "Exchange an authorization code for tokens",
	RunE:  notImplemented,
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the access token",
	RunE:  notImplemented,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show token scopes and expiry",
	RunE:  notImplemented,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Delete stored tokens",
	RunE:  notImplemented,
}

var authSetClientCmd = &cobra.Command{
	Use:   "set-client",
	Short: "Set OAuth client ID and secret",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authAuthorizeURLCmd)
	authCmd.AddCommand(authExchangeCmd)
	authCmd.AddCommand(authRefreshCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authSetClientCmd)

	authLoginCmd.Flags().StringVar(&authLoginOpts.RedirectURI, "redirect-uri", "", "override redirect URI")
	authLoginCmd.Flags().BoolVar(&authLoginOpts.NoOpen, "no-open", false, "print URL instead of opening a browser")
	authLoginCmd.Flags().StringVar(&authLoginOpts.Listen, "listen", "127.0.0.1:9876", "callback listen address")

	authAuthorizeURLCmd.Flags().StringVar(&authAuthorizeURLOpts.RedirectURI, "redirect-uri", "", "override redirect URI")
	authAuthorizeURLCmd.Flags().StringVar(&authAuthorizeURLOpts.Scope, "scope", "", "override OAuth scopes (comma-separated)")

	authExchangeCmd.Flags().StringVar(&authExchangeOpts.Code, "code", "", "authorization code (otherwise read from stdin)")

	authLogoutCmd.Flags().BoolVar(&authLogoutOpts.Force, "force", false, "skip confirmation")

	authSetClientCmd.Flags().StringVar(&authSetClientOpts.ClientID, "client-id", "", "OAuth client ID")
	authSetClientCmd.Flags().BoolVar(&authSetClientOpts.SecretStdin, "secret-stdin", false, "read client secret from stdin")
}
