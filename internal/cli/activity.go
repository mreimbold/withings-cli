package cli

import (
	"fmt"

	"github.com/mreimbold/withings-cli/internal/auth"
	"github.com/mreimbold/withings-cli/internal/services/activity"
	"github.com/spf13/cobra"
)

func newActivityCommand() *cobra.Command {
	var opts activity.Options

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
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			accessToken, err := auth.EnsureAccessToken(cmd.Context(), appOpts)
			if err != nil {
				return fmt.Errorf("ensure access token: %w", err)
			}

			return activity.Run(cmd.Context(), opts, appOpts, accessToken)
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
