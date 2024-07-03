package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/nce/tourenbuchctl/pkg/oauth"
	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/nce/tourenbuchctl/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Strava data to Tourenbuch",
	Long:  "This parses strava activity data to the yaml format of Tourenbuch",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize configuration before running any command
		initConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		tokenFile := "/tmp/stravatoken.json"

		token, err := utils.LoadToken(tokenFile)
		if err == nil && token.Valid() {
			log.Println("Using existing token")
			client := oauth.StravaOauthConfig.Client(context.Background(), token)
			strava.FetchStravaData(client)
		} else {

			oauth.InitStravaOauthConfig()

			log.Println("Using no token")
			log.Println("Sent to:", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString))
			err := exec.Command("open", oauth.StravaOauthConfig.AuthCodeURL(oauth.OauthStateString)).Start()
			if err != nil {
				log.Fatal(err)
			}

			oauth.AuthStrava(tokenFile)

			token, err := utils.LoadToken(tokenFile)

			if err == nil && token.Valid() {
				log.Println("Using new token")
				client := oauth.StravaOauthConfig.Client(context.Background(), token)
				strava.FetchStravaData(client)
			}

		}
	},
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

func init() {
	rootCmd.AddCommand(syncCmd)
}
