package activity

import (
	"errors"
	"os"

	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/rs/zerolog/log"
)

func (a *Activity) StravaSync() error {
	if a.Tb.Date.IsZero() {
		return errors.New("Date not properly initialized")
	}

	if a.Meta.StravaSync {
		stats, err := strava.FetchStravaData(a.Tb.Date)
		if err != nil {
			log.Error().Str("date", a.Tb.Date.String()).Msg("Strava Sync failing for date")
			return err
		}
		text, err := GetTextLibraryPath()
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting text activity location")
		}
		asset, err := GetAssetLibraryPath()
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting asset activity location")
		}

		a.Tb.Distance = stats.Distance
		a.Tb.Ascent = stats.Ascent
		a.Tb.StartTime = stats.StartDate
		a.Tb.ElapsedTime = stats.ElapsedTime
		a.Tb.MovingTime = stats.MovingTime
		a.Meta.StravaId = stats.Id
		a.Meta.Category = stats.SportType

		a.Meta.TextLocation = text + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/"
		a.Meta.AssetLocation = asset + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/"

		if a.Meta.TextLocation == "" {
			return errors.New("TextLocation not properly initialized")
		}

		err = a.updateActivity(a.Meta.TextLocation + "description.md")
		if err != nil {
			log.Error().Str("filePath", a.Meta.TextLocation+"description.md").Msg("Error updating activity")
			return err
		}
		log.Info().Str("filePath", a.Meta.TextLocation+"description.md").Msg("Updated strava stats in tourenbuch")

	}

	if a.Meta.StravaGpxSync {
		//nolint: gosimple
		if a.Meta.StravaSync == false {
			stats, err := strava.FetchStravaData(a.Tb.Date)
			if err != nil {
				return err
			}

			a.Meta.StravaId = stats.Id
			a.Meta.Category = stats.SportType

		}
		err := strava.ExportStravaToGpx(a.Meta.StravaId, a.Meta.AssetLocation+"input.gpx")
		if err != nil {
			return err
		}

		log.Info().Str("gpxFile", a.Meta.AssetLocation+"input.gpx").Msg("Exported Strava data to GPX")
	}

	return nil

}

func (a *Activity) CreateActivity() error {

	err := a.createFolder()
	if err != nil {
		return err
	}

	for _, file := range []string{"description.md", "elevation.plt", "Makefile", "img-even.tex"} {

		text, err := a.initSkeleton(file)
		if err != nil {
			return err
		}

		file, err := os.Create(a.Meta.TextLocation + "/" + file)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			return err
		}
	}

	err = a.StravaSync()
	if err != nil {
		return err
	}

	return nil
}
