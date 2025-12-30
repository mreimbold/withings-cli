package cmd

import "github.com/spf13/cobra"

type heartGetOptions struct {
	TimeRange  timeRangeOptions
	Pagination paginationOptions
	User       userOption
	LastUpdate lastUpdateOption
	Signal     bool
}

func newHeartCommand(notImplemented runEFunc) *cobra.Command {
	var opts heartGetOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	heartCmd := &cobra.Command{
		Use:   "heart",
		Short: "Heart data",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	heartGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Fetch heart data",
		RunE:  notImplemented,
	}

	heartCmd.AddCommand(heartGetCmd)

	addTimeRangeFlags(heartGetCmd, &opts.TimeRange)
	addPaginationFlags(heartGetCmd, &opts.Pagination)
	addUserIDFlag(heartGetCmd, &opts.User)
	addLastUpdateFlag(heartGetCmd, &opts.LastUpdate)

	heartGetCmd.Flags().BoolVar(
		&opts.Signal,
		"signal",
		false,
		"include signal metadata when available",
	)

	return heartCmd
}
