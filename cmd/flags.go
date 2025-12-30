package cmd

import "github.com/spf13/cobra"

type TimeRangeOptions struct {
	Start string
	End   string
}

type DateOption struct {
	Date string
}

type PaginationOptions struct {
	Limit  int
	Offset int
}

type UserOption struct {
	UserID string
}

type LastUpdateOption struct {
	LastUpdate int64
}

func addTimeRangeFlags(cmd *cobra.Command, opts *TimeRangeOptions) {
	cmd.Flags().StringVar(&opts.Start, "start", "", "start time (RFC3339 or epoch)")
	cmd.Flags().StringVar(&opts.End, "end", "", "end time (RFC3339 or epoch)")
}

func addDateFlag(cmd *cobra.Command, opts *DateOption) {
	cmd.Flags().StringVar(&opts.Date, "date", "", "date (YYYY-MM-DD)")
}

func addPaginationFlags(cmd *cobra.Command, opts *PaginationOptions) {
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "limit number of results")
	cmd.Flags().IntVar(&opts.Offset, "offset", 0, "offset into result set")
}

func addUserIDFlag(cmd *cobra.Command, opts *UserOption) {
	cmd.Flags().StringVar(&opts.UserID, "user-id", "", "Withings user ID")
}

func addLastUpdateFlag(cmd *cobra.Command, opts *LastUpdateOption) {
	cmd.Flags().Int64Var(&opts.LastUpdate, "last-update", 0, "last update timestamp (epoch)")
}
