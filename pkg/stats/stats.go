package stats

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nce/tourenbuchctl/pkg/activity"
	"github.com/rs/zerolog/log"
)

func WriteStats(activityTypes string, outputFormat string, regionalGrouping bool) {
	validActivities, err := filterActivityTypes(activityTypes)
	if err != nil {
		log.Error().Msgf("Error reading directory contents: %v", err)
	}

	activityCollection, err := gatherActivites(validActivities)
	if err != nil {
		log.Error().Msgf("Error gathering all activities: %v", err)
	}

	if outputFormat == "md" {
		printMarkdown(activityCollection, regionalGrouping)
	}
}

// filter activity types by string inputs like "mtb" or "mtb, skitour".
func filterActivityTypes(activityTypes string) ([]activity.ActivityType, error) {
	var validActivityTypes []activity.ActivityType

	if activityTypes == "all" {
		validActivityTypes = append(validActivityTypes, activity.ActivityTypes...)

		return validActivityTypes, nil
	}

	unfilteredActivityTypes := strings.Split(strings.ReplaceAll(activityTypes, " ", ""), ",")

	var filteredActivityTypes []activity.ActivityType

	for _, unfilteredActivityType := range unfilteredActivityTypes {
		// check if it's a valid activity type
		if activity.ValidActivityType(unfilteredActivityType) {
			for _, validType := range activity.ActivityTypes {
				// add this valid type to the slice
				if validType.Name == unfilteredActivityType {
					filteredActivityTypes = append(filteredActivityTypes, validType)
				}
			}
		}
	}

	if len(filteredActivityTypes) == 0 {
		return nil, ErrNoValidActivityTypes
	}

	return filteredActivityTypes, nil
}

type activityData struct {
	Title        string
	Dirname      string
	Date         string
	Region       string
	Ascent       string
	Distance     string
	Duration     string
	Participants string
}

func gatherActivites(activityTypes []activity.ActivityType) ([]activityData, error) {
	var skippedActivities int

	var validActivities int

	var activityCollection []activityData

	for _, activityFolder := range activityTypes {
		folders, err := os.ReadDir(activityFolder.TextPath)
		if err != nil {
			log.Error().Str("folder", activityFolder.TextPath).Msg("Error reading directory contents")

			return nil, fmt.Errorf("reading directory %w", err)
		}

		for _, folder := range folders {
			if folder.IsDir() {
				if folder.Name() == "multidaytrip" {
					continue
				}

				var activityData activityData

				headerPath := filepath.Join(activityFolder.TextPath, folder.Name(), "header.yaml")

				if _, err := os.Stat(headerPath); errors.Is(err, os.ErrNotExist) {
					skippedActivities++

					continue
				}

				activityData.Dirname = folder.Name()

				act, err := activity.GetFromHeader[string](
					filepath.Join(activityFolder.TextPath, folder.Name()),
					"Activity.Title",
					"Activity.Date",
					"Activity.PointOfOrigin.Region",
					"Activity.Company",
					"Stats.Ascent",
					"Stats.Distance",
					"Stats.OverallTime",
				)
				if err != nil {
					log.Error().Str("folder", headerPath).Msgf("Error reading header content: %v", err)
				}

				activityData.Title = act["Activity.Title"]
				activityData.Date = act["Activity.Date"]
				activityData.Region = act["Activity.PointOfOrigin.Region"]
				activityData.Participants = act["Activity.Company"]
				activityData.Ascent = act["Stats.Ascent"]
				activityData.Distance = act["Stats.Distance"]
				activityData.Duration = act["Stats.OverallTime"]

				validActivities++

				activityCollection = append(activityCollection, activityData.normalizeData())
			}
		}

		log.Info().Msgf("Skipped Activities %d", skippedActivities)
		log.Info().Msgf("Evaluated Activities %d", validActivities)
	}

	return activityCollection, nil
}

func (a *activityData) normalizeData() activityData {
	a.Title = strings.ReplaceAll(a.Title, "$\\rightarrow$", "↝")

	a.Region, _, _ = strings.Cut(a.Region, "-")
	a.Region = strings.TrimSpace(a.Region)

	return *a
}

type Region struct {
	Region string
	Count  int
}

func uniqueRegions(activityCollection []activityData) []Region {
	seen := make(map[string]int)

	for _, activity := range activityCollection {
		if _, ok := seen[activity.Region]; !ok {
			seen[activity.Region] = 0
		}

		seen[activity.Region]++
	}

	result := make([]Region, 0, len(seen))
	for region, count := range seen {
		result = append(result, Region{Region: region, Count: count})
		log.Info().Msgf("Region: %s; Count: %d", region, count)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}
