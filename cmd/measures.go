package cmd

import "github.com/spf13/cobra"

type measuresGetOptions struct {
	TimeRange  timeRangeOptions
	Pagination paginationOptions
	User       userOption
	LastUpdate lastUpdateOption
	Types      string
	Category   string
}

func newMeasuresCommand() *cobra.Command {
	var opts measuresGetOptions

	//nolint:exhaustruct // Cobra command defaults are intentional.
	measuresCmd := &cobra.Command{
		Use:   "measures",
		Short: "Body measures (weight, blood pressure, composition)",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	measuresGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Fetch body measures",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMeasuresGet(cmd, opts)
		},
	}

	measuresCmd.AddCommand(measuresGetCmd)

	addTimeRangeFlags(measuresGetCmd, &opts.TimeRange)
	addPaginationFlags(measuresGetCmd, &opts.Pagination)
	addUserIDFlag(measuresGetCmd, &opts.User)
	addLastUpdateFlag(measuresGetCmd, &opts.LastUpdate)

	measuresGetCmd.Flags().StringVar(
		&opts.Types,
		"type",
		emptyString,
		"measure types (comma-separated)",
	)
	measuresGetCmd.Flags().StringVar(
		&opts.Category,
		"category",
		emptyString,
		"category: real or goal",
	)

	return measuresCmd
}
