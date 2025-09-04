package activity

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// nolint: gochecknoglobals
var ActivityTypes []Kind

type Kind struct {
	Name      string
	TextPath  string
	AssetPath string
}

func SetupActivityKinds() {
	validActivities := viper.GetStringSlice("activities")
	if len(validActivities) == 0 {
		log.Fatal().Msg("no valid activities found in configuration")
	}

	textPath, err := GetTextLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Library Path")
	}

	assetPath, err := GetAssetLibraryPath()
	if err != nil {
		log.Error().Msg("Error getting Asset Path")
	}

	for _, validActivity := range validActivities {
		activity := Kind{
			Name:      validActivity,
			TextPath:  textPath + "/" + validActivity,
			AssetPath: assetPath + "/" + validActivity,
		}

		ActivityTypes = append(ActivityTypes, activity)
	}
}
