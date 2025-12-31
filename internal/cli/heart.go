package cli

import (
	"fmt"

	"github.com/mreimbold/withings-cli/internal/auth"
	"github.com/mreimbold/withings-cli/internal/services/heart"
	"github.com/spf13/cobra"
)

func newHeartCommand() *cobra.Command {
	var opts heart.Options

	//nolint:exhaustruct // Cobra command defaults are intentional.
	heartCmd := &cobra.Command{
		Use:   "heart",
		Short: "Heart data",
	}
	//nolint:exhaustruct // Cobra command defaults are intentional.
	heartGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Fetch heart data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			appOpts, err := readGlobalOptions(cmd.Root().PersistentFlags())
			if err != nil {
				return err
			}

			accessToken, err := auth.EnsureAccessToken(cmd.Context(), appOpts)
			if err != nil {
				return fmt.Errorf("ensure access token: %w", err)
			}

			return heart.Run(cmd.Context(), opts, appOpts, accessToken)
		},
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
