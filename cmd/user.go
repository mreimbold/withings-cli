package cmd

import "github.com/spf13/cobra"

func newUserCommand(notImplemented runEFunc) *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User and profile commands",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userMeCmd := &cobra.Command{
		Use:   "me",
		Short: "Show current user profile",
		RunE:  notImplemented,
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userListCmd := &cobra.Command{
		Use:   "list",
		Short: "List linked users",
		RunE:  notImplemented,
	}

	userCmd.AddCommand(userMeCmd)
	userCmd.AddCommand(userListCmd)

	return userCmd
}
