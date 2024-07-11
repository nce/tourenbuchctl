package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/nce/tourenbuchctl/cmd/new"
	"github.com/nce/tourenbuchctl/cmd/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debug bool
var rootCmd = &cobra.Command{
	Use:   "tourenbuchctl",
	Short: "tourenbuch CLI application",
	Long:  "A CLI application to interact with Tourenbuch.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging(debug)
		initConfig()
		log.Debug().Msg("Logging initialized")
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default action if no subcommands are specified
		fmt.Println("Tourenbuch CLI application")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(sync.NewSyncCommand())
	rootCmd.AddCommand(new.NewNewCommand())

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable logging in debug mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func initLogging(debug bool) {
	// setup code line caller for logging
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if debug {
		log.Logger = log.Level(zerolog.DebugLevel)
	} else {
		log.Logger = log.Level(zerolog.InfoLevel)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal().Msg("Config file .env not found")
		} else {
			log.Fatal().Msg("Parsing config file not possible. The file should contain " +
				"the following environment variables: STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET")
		}
	}
	log.Debug().Msg("Environment config initialized")
	viper.AutomaticEnv()
}
