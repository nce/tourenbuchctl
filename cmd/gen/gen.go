package gen

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/nce/tourenbuchctl/pkg/render"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewGenCommand() *cobra.Command {
	var preventCleanup bool

	var exportToDisk bool

	var exportToS3 bool

	var compression bool

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

			spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			spin.Start()

			page, err := render.NewPage(path, exportToDisk, exportToS3, compression)
			if err != nil {
				log.Fatal().Err(err).Str("cwd", path).Msg("Error creating a new page")
			}

			err = page.GenerateSinglePageActivity(preventCleanup)
			if err != nil {
				log.Fatal().Err(err).Msg("Rendering the single page failed")
			}

			spin.Stop()
			log.Info().Msg("Single Page rendered")
		},
	}

	cmd.Flags().BoolVarP(&exportToDisk, "save", "s", false, "Save the rendered pdf to assetdir on disk")
	cmd.Flags().BoolVarP(&exportToS3, "upload", "u", false, "upload the pdf to cloud s3")
	cmd.Flags().BoolVarP(&preventCleanup, "prevent-cleanup", "x", false, "Don't remove the temporary "+
		"rendering folders. Useful for debugging")
	cmd.Flags().BoolVarP(&compression, "compress", "c", false, "compress the pdf after generation")

	return cmd
}
