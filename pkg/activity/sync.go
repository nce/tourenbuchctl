package activity

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/nce/tourenbuchctl/pkg/maprender"
	"github.com/nce/tourenbuchctl/pkg/strava"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	ErrTextLocationNotInitialized = errors.New("textLocation not properly initialized")
	ErrDateNotInitialized         = errors.New("date not properly initialized")
)

//nolint:nestif
func (a *Activity) StravaSync() error {
	if a.Tb.Date.IsZero() {
		return fmt.Errorf("found empty: %w", ErrDateNotInitialized)
	}

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
		if !a.Meta.StravaSync {
			stats, err := strava.FetchStravaData(a.Tb.Date)
			if err != nil {
				return fmt.Errorf("fetching data from strava: %w", err)
			}

			a.Meta.StravaID = stats.ID
			a.Meta.Category = stats.SportType
		}

		gpxFile := a.Meta.AssetLocation + "input.gpx"

		err := strava.ExportStravaToGpx(a.Meta.StravaID, gpxFile)
		if err != nil {
			return fmt.Errorf("exporting gpx from strava: %w", err)
		}

		log.Info().Str("gpxFile", gpxFile).Msg("Exported Strava data to GPX")

		generateMapForGpx(gpxFile, a.Meta.Category)
	}

	return nil
}

func generateMapForGpx(gpxFile string, activityType string) {
	apiKey := viper.GetString("THUNDERFOREST_API_KEY")
	if apiKey == "" {
		log.Info().Msg("THUNDERFOREST_API_KEY not configured; skipping map generation")

		return
	}

	mapFile := filepath.Join(filepath.Dir(gpxFile), "map.png")

	err := maprender.GenerateForActivity(context.Background(), gpxFile, mapFile, apiKey, activityType)
	if err != nil {
		if errors.Is(err, maprender.ErrMissingAPIKey) {
			log.Info().Msg("THUNDERFOREST_API_KEY not configured; skipping map generation")

			return
		}

		log.Warn().
			Err(err).
			Str("gpxFile", gpxFile).
			Msg("Generating Thunderforest map failed; gen will retry if map.png is missing")

		return
	}

	log.Info().Str("mapFile", mapFile).Msg("Generated Thunderforest map")
}
