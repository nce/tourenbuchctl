package newactivity

import (
	"fmt"
	"time"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewNewCommand() *cobra.Command {
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new activity in Tourenbuch",
		Long:  "Create a new selected activity",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Fatal().Err(err).Msg("Error printing help")
			}
		},
	}
	newCmd.AddCommand(newMtbCommand())
	newCmd.AddCommand(newSkitourCommand())

	return newCmd
}

func newMtbCommand() *cobra.Command {
	act := activity.Activity{}

	cmd := &cobra.Command{
		Use:   "mtb [name]",
		Short: "Create a new mtb activity in Tourenbuch",
		Long:  "Create a new mtb activity",
		Args:  cobra.ExactArgs(1),
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			act.Meta.Name = args[0]
			act.Meta.Category = "mtb"
			log.Info().Msg("Creating new mtb activity")

			if act.Meta.QueryStartLocation {
				var err error
				act.Tb.StartLocationQr, err = activity.GetStartLocationQr()
				if err != nil {
					log.Fatal().Err(err).Msg("Error getting start location")
				}
				log.Debug().Str("startLocationQr", act.Tb.StartLocationQr).Msg("Start location set")
			}

			err := act.CreateActivity()
			if err != nil {
				log.Fatal().Err(err).Msg("Error creating activity")
			}
		},
	}

	var dateStr string

	// there is no maxHeight in mtb
	// cmd.Flags().IntVar(&flags.maxHeight, "height", "h", "Maximium absolute elevation in meter.")
	cmd.Flags().StringVarP(&act.Tb.Title, "title", "t", "", "Title of the activity")
	cmd.Flags().StringVarP(&act.Tb.Company, "company", "c", "", "Names of people who participated")
	cmd.Flags().StringVar(&act.Tb.Restaurant, "restaurant", "", "Names of people who participated")
	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")
	cmd.Flags().BoolVarP(&act.Meta.StravaSync, "sync", "s", true, "Get activity stats from strava")
	cmd.Flags().BoolVarP(&act.Meta.StravaGpxSync, "gpx", "g", true, "Get gpx track from strava")
	cmd.Flags().BoolVarP(&act.Meta.Multiday, "mutli", "m", false, "Is part of a multiday activity")
	cmd.Flags().BoolVarP(&act.Meta.QueryStartLocation, "start-location", "l", true, "Interactive"+
		"query for starting locations")
	cmd.Flags().IntVarP(&act.Tb.Rating, "rating", "r", 3, "Rating of the activity in the format '1-5'."+
		"This will be later displayed as stars")
	cmd.Flags().IntVarP(&act.Tb.Difficulty, "difficulty", "y", 3, "Difficulty of trails in S-Scale")

	err := cmd.MarkFlagRequired("date")
	if err != nil {
		log.Fatal().Err(err).Msg("Error in marking flag as required")
	}

	err = cmd.MarkFlagRequired("title")
	if err != nil {
		log.Fatal().Err(err).Msg("Error in marking flag as required")
	}

	//nolint: revive
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if dateStr != "" {
			parsedDate, err := time.Parse("02.01.2006", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}

			act.Tb.Date = parsedDate
		}

		return nil
	}

	return cmd
}

func newSkitourCommand() *cobra.Command {
	newSkitourCmd := &cobra.Command{
		Use:   "skitour",
		Short: "Create a new skitour activity in Tourenbuch",
		Long:  "Create a new skitour activity",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
		},
	}

	return newSkitourCmd
}
