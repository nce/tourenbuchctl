package activity

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// nolint: gochecknoglobals
var ActivityTypes []ActivityType

type ActivityType struct {
	Name      string
	TextPath  string
	AssetPath string
}

func GetActivityTypes() {
	validActivities := viper.GetStringSlice("activities")
	// var activityTypes []ActivityType

	textPath, err := GetTextLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Library Path")
	}

	assetPath, err := GetAssetLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Asset Path")
	}

	for _, validActivity := range validActivities {
		activity := ActivityType{
			Name:      validActivity,
			TextPath:  textPath + "/" + validActivity,
			AssetPath: assetPath + "/" + validActivity,
		}

		ActivityTypes = append(ActivityTypes, activity)
	}
}
