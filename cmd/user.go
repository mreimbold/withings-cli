package cmd

import "github.com/spf13/cobra"

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User and profile commands",
}

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user profile",
	RunE:  notImplemented,
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List linked users",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userMeCmd)
	userCmd.AddCommand(userListCmd)
}
