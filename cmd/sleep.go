package cmd

import "github.com/spf13/cobra"

type sleepGetOptions struct {
	TimeRange  timeRangeOptions
	Date       dateOption
	Pagination paginationOptions
	User       userOption
	LastUpdate lastUpdateOption
	Model      int
}

func newSleepCommand() *cobra.Command {
	var opts sleepGetOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	sleepCmd := &cobra.Command{
		Use:   "sleep",
		Short: "Sleep summaries",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	sleepGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Fetch sleep summaries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSleepGet(cmd, opts)
		},
	}

	sleepCmd.AddCommand(sleepGetCmd)

	addTimeRangeFlags(sleepGetCmd, &opts.TimeRange)
	addDateFlag(sleepGetCmd, &opts.Date)
	addPaginationFlags(sleepGetCmd, &opts.Pagination)
	addUserIDFlag(sleepGetCmd, &opts.User)
	addLastUpdateFlag(sleepGetCmd, &opts.LastUpdate)

	sleepGetCmd.Flags().IntVar(
		&opts.Model,
		"model",
		defaultInt,
		"sleep model (if supported)",
	)

	return sleepCmd
}
