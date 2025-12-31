package cli

import (
	"github.com/mreimbold/withings-cli/internal/params"
	"github.com/spf13/cobra"
)

func addTimeRangeFlags(cmd *cobra.Command, opts *params.TimeRange) {
	cmd.Flags().StringVar(
		&opts.Start,
		"start",
		emptyString,
		"start time (RFC3339 or epoch)",
	)
	cmd.Flags().StringVar(
		&opts.End,
		"end",
		emptyString,
		"end time (RFC3339 or epoch)",
	)
}

func addDateFlag(cmd *cobra.Command, opts *params.Date) {
	cmd.Flags().StringVar(
		&opts.Date,
		"date",
		emptyString,
		"date (YYYY-MM-DD)",
	)
}

func addPaginationFlags(cmd *cobra.Command, opts *params.Pagination) {
	cmd.Flags().IntVar(
		&opts.Limit,
		"limit",
		defaultInt,
		"limit number of results",
	)
	cmd.Flags().IntVar(
		&opts.Offset,
		"offset",
		defaultInt,
		"offset into result set",
	)
}

func addUserIDFlag(cmd *cobra.Command, opts *params.User) {
	cmd.Flags().StringVar(
		&opts.UserID,
		"user-id",
		emptyString,
		"Withings user ID",
	)
}

func addLastUpdateFlag(cmd *cobra.Command, opts *params.LastUpdate) {
	cmd.Flags().Int64Var(
		&opts.LastUpdate,
		"last-update",
		defaultInt64,
		"last update timestamp (epoch)",
	)
}
