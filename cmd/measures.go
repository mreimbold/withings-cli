package cmd

import "github.com/spf13/cobra"

type measuresGetOptions struct {
	TimeRange  TimeRangeOptions
	Pagination PaginationOptions
	User       UserOption
	LastUpdate LastUpdateOption
	Types      string
	Category   string
}

var measuresGetOpts measuresGetOptions

var measuresCmd = &cobra.Command{
	Use:   "measures",
	Short: "Body measures (weight, blood pressure, composition)",
}

var measuresGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch body measures",
	RunE:  notImplemented,
}

func init() {
	rootCmd.AddCommand(measuresCmd)
	measuresCmd.AddCommand(measuresGetCmd)

	addTimeRangeFlags(measuresGetCmd, &measuresGetOpts.TimeRange)
	addPaginationFlags(measuresGetCmd, &measuresGetOpts.Pagination)
	addUserIDFlag(measuresGetCmd, &measuresGetOpts.User)
	addLastUpdateFlag(measuresGetCmd, &measuresGetOpts.LastUpdate)

	measuresGetCmd.Flags().StringVar(&measuresGetOpts.Types, "type", "", "measure types (comma-separated)")
	measuresGetCmd.Flags().StringVar(&measuresGetOpts.Category, "category", "", "category: real or goal")
}
