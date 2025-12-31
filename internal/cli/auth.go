package cli

import (
	"github.com/mreimbold/withings-cli/internal/auth"
	"github.com/spf13/cobra"
)

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
	var opts auth.LoginOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Start browser OAuth flow and store tokens",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			return auth.Login(cmd.Context(), opts, appOpts)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			return auth.Status(appOpts)
		},
	}
}

func newAuthLogoutCommand() *cobra.Command {
	var opts auth.LogoutOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Delete stored tokens",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			return auth.Logout(opts, appOpts)
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
