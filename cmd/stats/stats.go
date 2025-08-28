package stats

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewStatsCommand() *cobra.Command {

	var outputFormat string
	var activityType string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "create statistics of activities",
		Long:  "Create a sortable list of activities with properties like length, elevation, duration,...",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {

			log.Info().Msg("Single Page rendered")
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output-format", "o", "md", "the output format (md/csv)")
	cmd.Flags().StringVarP(&activityType, "activity-type", "t", "all", "which type of activites "+
		"should be parsed")

	return cmd
}
