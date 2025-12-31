package cli

import (
	"fmt"

	"github.com/mreimbold/withings-cli/internal/auth"
	"github.com/mreimbold/withings-cli/internal/services/sleep"
	"github.com/spf13/cobra"
)

func newSleepCommand() *cobra.Command {
	var opts sleep.Options

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
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			accessToken, err := auth.EnsureAccessToken(cmd.Context(), appOpts)
			if err != nil {
				return fmt.Errorf("ensure access token: %w", err)
			}

			return sleep.Run(cmd.Context(), opts, appOpts, accessToken)
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
