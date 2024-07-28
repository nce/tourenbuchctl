package sync

import (
	"fmt"
	"time"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewSyncCommand() *cobra.Command {
	var (
		parsedDate time.Time
		dateStr    string
	)

	act := activity.Activity{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync Strava data to Tourenbuch",
		Long:  "This parses strava activity data to the yaml format of Tourenbuch and exports the gpx track.",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			name, date, err := activity.GetActivityLocation()
			if err != nil {
				log.Fatal().Err(err).Msg("Error getting activity location")
			}

			if act.Meta.Name == "" {
				act.Meta.Name = name
			}

			if parsedDate.IsZero() {
				act.Tb.Date, err = time.Parse("02.01.2006", date)
				if err != nil {
					log.Fatal().Err(err).Str("date", date).Msg("Error parsing the date extracted from the activity file location")
				}
			} else {
				act.Tb.Date = parsedDate
			}

			err = act.StravaSync()
			if err != nil {
				log.Fatal().Err(err).Msg("Error fetching strava data")
			}
		},
	}

	cmd.Flags().BoolVarP(&act.Meta.StravaSync, "sync", "s", true, "Get activity stats from strava")
	cmd.Flags().BoolVarP(&act.Meta.StravaGpxSync, "gpx", "g", true, "Get gpx track from strava")
	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")

	//nolint: revive
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if dateStr != "" {
			var err error

			parsedDate, err = time.Parse("02.01.2006", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}
		}

		return nil
	}

	return cmd
}
