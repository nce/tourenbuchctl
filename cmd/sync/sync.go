package sync

import (
	"fmt"
	"os"
	"time"

	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	parsedDate time.Time
)

func NewSyncCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync Strava data to Tourenbuch",
		Long:  "This parses strava activity data to the yaml format of Tourenbuch",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize configuration before running any command
			initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			foo := strava.FetchStravaData(parsedDate)
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
