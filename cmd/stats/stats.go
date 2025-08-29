package stats

import (
	"github.com/nce/tourenbuchctl/pkg/stats"
	"github.com/spf13/cobra"
)

func NewStatsCommand() *cobra.Command {
	var outputFormat string

	var activityType string

	var regionalGrouping bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "create statistics of activities",
		Long:  "Create a sortable list of activities with properties like length, elevation, duration,...",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			stats.WriteStats(activityType, outputFormat, regionalGrouping)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output-format", "o", "md", "the output format (md/csv)")
	cmd.Flags().BoolVarP(&regionalGrouping, "regional-grouping", "r", false, "group activities by region")
	cmd.Flags().StringVarP(&activityType, "activity-type", "t", "all", "which type of activities "+
		"should be parsed")

	return cmd
}
