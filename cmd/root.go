package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nce/tourenbuchctl/cmd/gen"
	"github.com/nce/tourenbuchctl/cmd/migrate"
	newactivity "github.com/nce/tourenbuchctl/cmd/newActivity"
	"github.com/nce/tourenbuchctl/cmd/stats"
	"github.com/nce/tourenbuchctl/cmd/sync"
	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	//nolint: gochecknoglobals
	Version = "dev"
	//nolint: gochecknoglobals
	Commit = "none"
	//nolint: gochecknoglobals
	Date = "unknown"
)

func newRootCmd() *cobra.Command {
	var debug bool

	rootCmd := &cobra.Command{
		Use:   "tourenbuchctl",
		Short: "tourenbuch CLI application",
		Long:  "A CLI application to interact with Tourenbuch.",
		//nolint: revive
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initLogging(debug)
			initConfig()
			log.Debug().Msg("Logging initialized")
			if v, _ := cmd.Flags().GetBool("version"); v {
				printVersion()
			}
		},
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			// Default action if no subcommands are specified
			// add usage
		},
	}

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable logging in debug mode")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print application version")

	return rootCmd
}

func printVersion() {
	logger := zerolog.New(os.Stderr).With().Logger()
	logger.Info().
		Str("application", Version).
		Str("commit", Commit).
		Str("date", Date).
		Send()
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print application version",
		Long:  "Print application version and quit",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	rootCmd := newRootCmd()
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(sync.NewSyncCommand())
	rootCmd.AddCommand(newactivity.NewNewCommand())
	rootCmd.AddCommand(gen.NewGenCommand())
	rootCmd.AddCommand(migrate.NewMigrateCommand())
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(stats.NewStatsCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}

func initLogging(debug bool) {
	// setup code line caller for logging
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
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
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get home directory")
	}

	viper.SetConfigName(".tourenbuchctl")
	// viper.SetConfigType("env")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Fatal().Msg("Config file ~/.tourenbuchctl not found")
		} else {
			log.Fatal().Msg("Parsing config file not possible. The file should contain " +
				"the following environment variables: STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET")
		}
	}

	log.Debug().Msg("Environment config initialized")
	viper.AutomaticEnv()

	activity.GetActivityKinds()
}
