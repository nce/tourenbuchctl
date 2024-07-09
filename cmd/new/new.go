package new

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nce/tourenbuchctl/cmd/flags"
	"github.com/nce/tourenbuchctl/pkg/activity"
)

var (
	flag = &flags.CreateMtbFlags{}
)

func NewNewCommand() *cobra.Command {
	var newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new activity in Tourenbuch",
		Long:  "Create a new selected activity",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize configuration before running any command
			initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
			fmt.Println("newnewnew")

		},
	}
	newCmd.AddCommand(newMtbCommand())
	newCmd.AddCommand(newSkitourCommand())

	return newCmd
}

func newMtbCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "mtb [name]",
		Short: "Create a new mtb activity in Tourenbuch",
		Long:  "Create a new mtb activity",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			flag.Core.Name = args[0]

			flag.Core.StartLocationQr = activity.GetStartLocationQr()
			err := activity.CreateActivity(flag)
			if err != nil {
				panic(err)
			}
		},
	}

	var dateStr string

	// there is no maxHeight in mtb
	// cmd.Flags().IntVar(&flags.maxHeight, "height", "h", "Maximium absolute elevation in meter.")
	cmd.Flags().StringVarP(&flag.Core.Title, "title", "t", "", "Title of the activity")
	cmd.Flags().StringVarP(&flag.Company, "company", "c", "", "Names of people who participated")
	cmd.Flags().StringVar(&flag.Restaurant, "restaurant", "", "Names of people who participated")
	cmd.Flags().StringVarP(&dateStr, "date", "d", "", "Date of the activity in the format 'DD.MM.YYYY'")
	cmd.Flags().BoolVarP(&flag.Core.StravaSync, "sync", "s", true, "Get activity stats from strava")
	cmd.Flags().BoolVarP(&flag.Core.QueryStartLocation, "start-location", "l", true, "Interactive query for starting locations")
	cmd.Flags().IntVarP(&flag.Rating, "rating", "r", 3, "Rating of the activity in the format '1-5'."+
		"This will be later displayed as stars")
	cmd.Flags().IntVarP(&flag.Difficulty, "difficulty", "y", 3, "Difficulty of trails in S-Scale")

	err := cmd.MarkFlagRequired("date")
	if err != nil {
		panic(err)
	}
	err = cmd.MarkFlagRequired("title")
	if err != nil {
		panic(err)
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if dateStr != "" {
			parsedDate, err := time.Parse("02.01.2006", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %v", err)
			}
			flag.Core.Date = parsedDate
		}
		return nil
	}

	return cmd

}

func newSkitourCommand() *cobra.Command {
	var newSkitourCmd = &cobra.Command{
		Use:   "skitour",
		Short: "Create a new skitour activity in Tourenbuch",
		Long:  "Create a new skitour activity",
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
			fmt.Println("Strava CLI application")
		},
	}
	return newSkitourCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Error: Config file .env not found")
			os.Exit(1)
		} else {
			fmt.Printf("Error reading config file, %s\n", err)
			os.Exit(1)
		}
	}
	viper.AutomaticEnv() // read in environment variables that match
}
