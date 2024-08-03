package migrate

import (
	"os"

	"github.com/nce/tourenbuchctl/pkg/migrate"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "migrate the old activity format to the newest one",
		Long:  "parses the current state of the activity and migrates it to the current format",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			path, err := os.Getwd()
			if err != nil {
				log.Fatal().Err(err).Msg("Error getting the current working directory")
			}

			migrated, err := migrate.SplitDescriptionFile(path + "/")
			if err != nil {
				log.Fatal().Err(err).Str("filename", path).Msg("migration error")
			}

			if migrated {
				log.Info().Str("filename", path).
					Msg("Description split into header.yaml and description.md")
			}

			log.Info().Msg("Activity migrated")
		},
	}

	// cmd.Flags().BoolVarP(&act.Meta.StravaSync, "sync", "s", true, "Get activity stats from strava")

	return cmd
}
