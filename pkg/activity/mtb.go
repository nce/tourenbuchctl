package activity

import (
	"errors"
	"fmt"
	"os"

	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/rs/zerolog/log"
)

var (
	ErrTextLocationNotInitialized = errors.New("textLocation not properly initialized")
	ErrDateNotInitialized         = errors.New("date not properly initialized")
)

func (a *Activity) StravaSync() error {
	if a.Tb.Date.IsZero() {
		return fmt.Errorf("found empty: %w", ErrDateNotInitialized)
	}

	//nolint: nestif
	if a.Meta.StravaSync {
		stats, err := strava.FetchStravaData(a.Tb.Date)
		if err != nil {
			log.Error().Str("date", a.Tb.Date.String()).Msg("Strava Sync failing for date")

			return fmt.Errorf("fetching data from strava: %w", err)
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
		a.Meta.StravaID = stats.ID
		a.Meta.Category = stats.SportType

		// this is duplicated and should be refactored
		if a.Meta.Multiday {
			a.Meta.TextLocation = text + a.Meta.Category + "/" + "multidaytrip/" + a.Meta.Name + "/"
			a.Meta.AssetLocation = asset + a.Meta.Category + "/" + "multidaytrip/" + a.Meta.Name + "/"
		} else {
			a.Meta.TextLocation = text + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/"
			a.Meta.AssetLocation = asset + a.Meta.Category + "/" + a.Meta.Name + "-" + a.normalizeDate() + "/"
		}

		if a.Meta.TextLocation == "" {
			return fmt.Errorf("found empty: %w", ErrTextLocationNotInitialized)
		}

		err = a.updateActivity(a.Meta.TextLocation)
		if err != nil {
			log.Error().Str("filePath", a.Meta.TextLocation).Msg("Error updating activity")

			return fmt.Errorf("updating activity: %w", err)
		}

		log.Info().Str("filePath", a.Meta.TextLocation).Msg("Updated strava stats in tourenbuch")
	}

	if a.Meta.StravaGpxSync {
		//nolint: gosimple
		if a.Meta.StravaSync == false {
			stats, err := strava.FetchStravaData(a.Tb.Date)
			if err != nil {
				return fmt.Errorf("fetching data from strava: %w", err)
			}

			a.Meta.StravaID = stats.ID
			a.Meta.Category = stats.SportType
		}

		err := strava.ExportStravaToGpx(a.Meta.StravaID, a.Meta.AssetLocation+"input.gpx")
		if err != nil {
			return fmt.Errorf("exporting gpx from strava: %w", err)
		}

		log.Info().Str("gpxFile", a.Meta.AssetLocation+"input.gpx").Msg("Exported Strava data to GPX")
	}

	return nil
}

func (a *Activity) CreateActivity() error {
	err := a.createFolder()
	if err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	for _, file := range []string{"description.md", "header.yaml", "elevation.plt"} {
		text, err := a.initSkeleton(file)
		if err != nil {
			return fmt.Errorf("creating init skelton %w", err)
		}

		file, err := os.Create(a.Meta.TextLocation + "/" + file)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	}

	err = a.StravaSync()
	if err != nil {
		return err
	}

	return nil
}
