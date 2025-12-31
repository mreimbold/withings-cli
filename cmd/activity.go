package cmd

import "github.com/spf13/cobra"

type activityGetOptions struct {
	TimeRange  timeRangeOptions
	Date       dateOption
	Pagination paginationOptions
	User       userOption
	LastUpdate lastUpdateOption
}

func newActivityCommand() *cobra.Command {
	var opts activityGetOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	activityCmd := &cobra.Command{
		Use:   "activity",
		Short: "Activity summaries",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	activityGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Fetch activity summaries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runActivityGet(cmd, opts)
		},
	}

	activityCmd.AddCommand(activityGetCmd)

	addTimeRangeFlags(activityGetCmd, &opts.TimeRange)
	addDateFlag(activityGetCmd, &opts.Date)
	addPaginationFlags(activityGetCmd, &opts.Pagination)
	addUserIDFlag(activityGetCmd, &opts.User)
	addLastUpdateFlag(activityGetCmd, &opts.LastUpdate)

	return activityCmd
}
