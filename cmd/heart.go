package cmd

import "github.com/spf13/cobra"

type heartGetOptions struct {
	TimeRange  TimeRangeOptions
	Pagination PaginationOptions
	User       UserOption
	LastUpdate LastUpdateOption
	Signal     bool
}

var heartGetOpts heartGetOptions

var heartCmd = &cobra.Command{
	Use:   "heart",
	Short: "Heart data",
}

var heartGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch heart data",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(heartCmd)
	heartCmd.AddCommand(heartGetCmd)

	addTimeRangeFlags(heartGetCmd, &heartGetOpts.TimeRange)
	addPaginationFlags(heartGetCmd, &heartGetOpts.Pagination)
	addUserIDFlag(heartGetCmd, &heartGetOpts.User)
	addLastUpdateFlag(heartGetCmd, &heartGetOpts.LastUpdate)

	heartGetCmd.Flags().BoolVar(&heartGetOpts.Signal, "signal", false, "include signal metadata when available")
}
