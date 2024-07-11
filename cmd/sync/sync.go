package sync

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/spf13/cobra"
)

var (
	parsedDate time.Time
)

func NewSyncCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync Strava data to Tourenbuch",
		Long:  "This parses strava activity data to the yaml format of Tourenbuch",
		Run: func(cmd *cobra.Command, args []string) {
			foo, err := strava.FetchStravaData(parsedDate)
			if err != nil {
				log.Fatal().Err(err).Msg("Error fetching strava data")
			}
			fmt.Println(foo.Distance)
		},
	}

	var dateStr string

	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")
	err := cmd.MarkFlagRequired("date")
	if err != nil {
		panic(err)
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if dateStr != "" {
			var err error
			parsedDate, err = time.Parse("02.01.2006", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %v", err)
			}
		}
		return nil
	}

	return cmd
}
