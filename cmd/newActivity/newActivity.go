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
	newCmd.AddCommand(newHikeCommand())
	newCmd.AddCommand(newSkitourCommand())
	newCmd.AddCommand(newAlpineSkiCommand())

	return newCmd
}

func addActivityFlags(cmd *cobra.Command, act *activity.Activity) {
	var dateStr string
	// there is no maxHeight in mtb
	cmd.Flags().StringVarP(&act.Tb.Title, "title", "t", "", "Title of the activity")
	cmd.Flags().StringVarP(&act.Tb.Company, "company", "c", "", "Names of people who participated")
	cmd.Flags().StringVar(&act.Tb.Restaurant, "restaurant", "", "Name of restaurant/pause location")
	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")
	cmd.Flags().BoolVarP(&act.Meta.StravaSync, "sync", "s", true, "Get activity stats from strava")
	cmd.Flags().BoolVarP(&act.Meta.StravaGpxSync, "gpx", "g", true, "Get gpx track from strava")
	cmd.Flags().BoolVarP(&act.Meta.Multiday, "multi", "m", false, "Is part of a multiday activity")
	cmd.Flags().BoolVarP(&act.Meta.QueryStartLocation, "start-location", "l", true, "Interactive"+
		"query for starting locations")
	cmd.Flags().IntVarP(&act.Tb.Rating, "rating", "r", 3, "Rating of the activity in the format '1-5'."+
		"This will be later displayed as stars")

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
}

func createNewActivity(act *activity.Activity) {
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
			log.Info().Msgf("Creating new %s activity", act.Meta.Category)

			createNewActivity(&act)
		},
	}

	addActivityFlags(cmd, &act)
	cmd.Flags().IntVarP(&act.Tb.TrailDifficulty, "difficulty", "y", 3, "Difficulty of trails in S-Scale")

	return cmd
}

func newAlpineSkiCommand() *cobra.Command {
	act := activity.Activity{}

	cmd := &cobra.Command{
		Use:   "alpineSki [name]",
		Short: "Create a new alpine Ski activity in Tourenbuch",
		Long:  "Create a new alspine Ski activity",
		Args:  cobra.ExactArgs(1),
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			act.Meta.Name = args[0]
			act.Meta.Category = "alpineSki"
			log.Info().Msgf("Creating new %s activity", act.Meta.Category)

			createNewActivity(&act)
		},
	}

	addActivityFlags(cmd, &act)
	cmd.Flags().IntVar(&act.Tb.AlpineSki.Runs, "runs", 0, "Number of descends")
	cmd.Flags().IntVar(&act.Tb.AlpineSki.Vertical, "vertical", 0, "Vertical meters skied")
	cmd.Flags().IntVar(&act.Tb.Distance, "distance", 0, "Kilometers skied")

	return cmd
}

func newHikeCommand() *cobra.Command {
	act := activity.Activity{}

	cmd := &cobra.Command{
		Use:   "hike [name]",
		Short: "Create a new hike activity in Tourenbuch",
		Long:  "Create a new hike activity",
		Args:  cobra.ExactArgs(1),
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			act.Meta.Name = args[0]
			act.Meta.Category = "wandern"
			log.Info().Msgf("Creating new %s activity", act.Meta.Category)

			createNewActivity(&act)
		},
	}

	addActivityFlags(cmd, &act)

	return cmd
}

func newSkitourCommand() *cobra.Command {
	act := activity.Activity{}

	cmd := &cobra.Command{
		Use:   "skitour [name]",
		Short: "Create a new skitour activity in Tourenbuch",
		Long:  "Create a new skitour activity",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			act.Meta.Name = args[0]
			act.Meta.Category = "skitour"
			log.Info().Msgf("Creating new %s activity", act.Meta.Category)

			createNewActivity(&act)
		},
	}

	addActivityFlags(cmd, &act)
	cmd.Flags().IntVar(&act.Tb.MaxElevation, "max-elevation", 0, "Maximium absolute elevation in meter.")

	return cmd
}
