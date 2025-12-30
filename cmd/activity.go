package cmd

import "github.com/spf13/cobra"

type activityGetOptions struct {
	TimeRange  TimeRangeOptions
	Date       DateOption
	Pagination PaginationOptions
	User       UserOption
	LastUpdate LastUpdateOption
}

var activityGetOpts activityGetOptions

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Activity summaries",
}

var activityGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch activity summaries",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(activityCmd)
	activityCmd.AddCommand(activityGetCmd)

	addTimeRangeFlags(activityGetCmd, &activityGetOpts.TimeRange)
	addDateFlag(activityGetCmd, &activityGetOpts.Date)
	addPaginationFlags(activityGetCmd, &activityGetOpts.Pagination)
	addUserIDFlag(activityGetCmd, &activityGetOpts.User)
	addLastUpdateFlag(activityGetCmd, &activityGetOpts.LastUpdate)
}
