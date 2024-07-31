package gen

import (
	"os"

	"github.com/nce/tourenbuchctl/pkg/render"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewGenCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate a pdf from a single activity. This needs to be run from a directory containing the activity",
		Long:  "It parses the yaml/md description, compiles the elevation profile and map and generates a pdf file.",
		//nolint: revive
		Run: func(cmd *cobra.Command, args []string) {
			path, err := os.Getwd()
			if err != nil {
				log.Fatal().Err(err).Msg("Error getting the current working directory")
			}
			//path = filepath.Base(path)

			page, err := render.NewPage(path)
			if err != nil {
				log.Fatal().Err(err).Str("cwd", path).Msg("Error creating a new page")
			}

			err = page.GenerateSinglePageActivity()
			if err != nil {
				log.Fatal().Err(err).Msg("Rendering the single page failed")
			}
			log.Info().Msg("Single Page rendered")
		},
	}

	// cmd.Flags().BoolVarP(&act.Meta.StravaSync, "sync", "s", true, "Get activity stats from strava")

	return cmd
}
