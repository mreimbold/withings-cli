package cmd

import "github.com/spf13/cobra"

func newUserCommand() *cobra.Command {
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User and profile commands",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userMeCmd := &cobra.Command{
		Use:   "me",
		Short: "Show current user profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUserMe(cmd)
		},
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	userListCmd := &cobra.Command{
		Use:   "list",
		Short: "List linked users",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUserList(cmd)
		},
	}

	userCmd.AddCommand(userMeCmd)
	userCmd.AddCommand(userListCmd)

	return userCmd
}
