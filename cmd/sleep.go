package cmd

import "github.com/spf13/cobra"

type sleepGetOptions struct {
	TimeRange  TimeRangeOptions
	Date       DateOption
	Pagination PaginationOptions
	User       UserOption
	LastUpdate LastUpdateOption
	Model      int
}

var sleepGetOpts sleepGetOptions

var sleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Sleep summaries",
}

var sleepGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch sleep summaries",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(sleepCmd)
	sleepCmd.AddCommand(sleepGetCmd)

	addTimeRangeFlags(sleepGetCmd, &sleepGetOpts.TimeRange)
	addDateFlag(sleepGetCmd, &sleepGetOpts.Date)
	addPaginationFlags(sleepGetCmd, &sleepGetOpts.Pagination)
	addUserIDFlag(sleepGetCmd, &sleepGetOpts.User)
	addLastUpdateFlag(sleepGetCmd, &sleepGetOpts.LastUpdate)

	sleepGetCmd.Flags().IntVar(&sleepGetOpts.Model, "model", 0, "sleep model (if supported)")
}
