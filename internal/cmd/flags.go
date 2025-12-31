package cmd

import "github.com/spf13/cobra"

type timeRangeOptions struct {
	Start string
	End   string
}

type dateOption struct {
	Date string
}

type paginationOptions struct {
	Limit  int
	Offset int
}

type userOption struct {
	UserID string
}

type lastUpdateOption struct {
	LastUpdate int64
}

func addTimeRangeFlags(cmd *cobra.Command, opts *timeRangeOptions) {
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

func addDateFlag(cmd *cobra.Command, opts *dateOption) {
	cmd.Flags().StringVar(
		&opts.Date,
		"date",
		emptyString,
		"date (YYYY-MM-DD)",
	)
}

func addPaginationFlags(cmd *cobra.Command, opts *paginationOptions) {
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

func addUserIDFlag(cmd *cobra.Command, opts *userOption) {
	cmd.Flags().StringVar(
		&opts.UserID,
		"user-id",
		emptyString,
		"Withings user ID",
	)
}

func addLastUpdateFlag(cmd *cobra.Command, opts *lastUpdateOption) {
	cmd.Flags().Int64Var(
		&opts.LastUpdate,
		"last-update",
		defaultInt64,
		"last update timestamp (epoch)",
	)
}
