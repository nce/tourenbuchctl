package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nce/strava2tourenbuch/pkg/oauth"
	"github.com/nce/strava2tourenbuch/pkg/strava"
	"github.com/nce/strava2tourenbuch/pkg/utils"
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

			server := &http.Server{Addr: ":8080"}

			http.HandleFunc("/", handleMain)
			http.HandleFunc("/login", oauth.HandleStravaLogin)
			http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
				oauth.HandleStravaCallback(w, r, server, tokenFile)
			})

			log.Println("Started running on http://localhost:8080")
			log.Fatal(server.ListenAndServe())
		}
	},
}

// initConfig reads in config file and ENV variables if set.
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

func handleMain(w http.ResponseWriter, r *http.Request) {
	var html = `<html><body><a href="/login">Log in with Strava</a></body></html>`
	fmt.Fprint(w, html)
}
